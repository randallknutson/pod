package pod

import (
	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/avereha/pod/pkg/command"
	"github.com/avereha/pod/pkg/eap"
	"github.com/avereha/pod/pkg/pair"
	"github.com/avereha/pod/pkg/message"

	"github.com/avereha/pod/pkg/encrypt"
	"github.com/avereha/pod/pkg/response"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type Pod struct {
	ble   *bluetooth.Ble
	state *PODState
}

func New(ble *bluetooth.Ble, stateFile string, freshState bool) *Pod {
	var err error

	state := &PODState{
		filename: stateFile,
	}
	if !freshState {
		state, err = NewState(stateFile)
		if err != nil {
			log.Fatalf("could not restore pod state from %s: %+v", stateFile, err)
		}
	}

	ret := &Pod{
		ble:   ble,
		state: state,
	}

	return ret
}

func (p *Pod) StartAcceptingCommands() {
	log.Infof("got a new connection, start accepting commands")
	firstCmd, _ := p.ble.ReadCmd()
	log.Infof("got first command: %s", firstCmd)

	p.ble.StartMessageLoop()

	if p.state.LTK != nil { // paired, just establish new session
		msg, _ := p.ble.ReadMessageExpectingCommand(firstCmd)
		p.EapAka(msg)
	} else {
		p.StartActivation() // not paired, get the LTK
	}
}

func (p *Pod) StartActivation() {

	pair := &pair.Pair{}
	msg, _ := p.ble.ReadMessage()
	if err := pair.ParseSP1SP2(msg); err != nil {
		log.Fatalf("error parsing SP1SP2 %s", err)
	}
	// read PDM public key and nonce
	msg, _ = p.ble.ReadMessage()
	if err := pair.ParseSPS1(msg); err != nil {
		log.Fatalf("error parsing SPS1 %s", err)
	}

	msg, err := pair.GenerateSPS1()
	if err != nil {
		log.Fatal(err)
	}
	// send POD public key and nonce
	p.ble.WriteMessage(msg)

	// read PDM conf value
	msg, _ = p.ble.ReadMessage()
	pair.ParseSPS2(msg)

	// send POD conf value
	msg, err = pair.GenerateSPS2()
	if err != nil {
		log.Fatal(err)
	}
	p.ble.WriteMessage(msg)

	// receive SP0GP0 constant from PDM
	msg, _ = p.ble.ReadMessage()
	err = pair.ParseSP0GP0(msg)
	if err != nil {
		log.Fatalf("could not parse SP0GP0: %s", err)
	}

	// send P0 constant
	msg, err = pair.GenerateP0()
	if err != nil {
		log.Fatal(err)
	}
	p.ble.WriteMessage(msg)

	p.state.LTK, err = pair.LTK()
	if err != nil {
		log.Fatalf("could not get LTK %s", err)
	}
	log.Infof("LTK %x", p.state.LTK)
	p.state.EapAkaSeq = 1
	p.state.Save()

	msg, err = p.ble.ReadMessage()
	if err != nil {
	       log.Fatalf("could not get EapAka msg %s", err)
	}
	p.EapAka(msg)
}

func (p *Pod) EapAka(msg *message.Message) {

	session := eap.NewEapAkaChallenge(p.state.LTK, p.state.EapAkaSeq)

	err := session.ParseChallenge(msg)
	if err != nil {
		log.Fatalf("error parsing the EAP-AKA challenge: %s", err)
	}

	msg, err = session.GenerateChallengeResponse()
	if err != nil {
		log.Fatalf("error generating the eap-aka challenge response")
	}
	p.ble.WriteMessage(msg)

	msg, _ = p.ble.ReadMessage()
	log.Debugf("success? %x", msg.Payload) // TODO: figure out how error looks like
	err = session.ParseSuccess(msg)
	if err != nil {
		log.Fatalf("error parsing the EAP-AKA Success packet: %s", err)
	}
	p.state.CK, p.state.NoncePrefix = session.CKNoncePrefix()

	p.state.NonceSeq = 1
	p.state.MsgSeq = 1
	p.state.EapAkaSeq = session.Sqn
	log.Infof("got CK: %x", p.state.CK)
	log.Infof("got NONCE: %x", p.state.NoncePrefix)
	log.Infof("using NONCE SEQ: %d", p.state.NonceSeq)
	log.Infof("EAP-AKA session SQN: %d", p.state.EapAkaSeq)

	err = p.state.Save()
	if err != nil {
		log.Fatalf("Could not save the pod state: %s", err)
	}

	p.CommandLoop()
}

func (p *Pod) CommandLoop() {
	var lastMsgSeq uint8 = 0
	for {
		log.Debugf("reading the next command")
		msg, _ := p.ble.ReadMessage()
		log.Tracef("got command message: %s", spew.Sdump(msg))

		if msg.SequenceNumber == lastMsgSeq {
			// this is a retry because we did not answer yet
			// ignore duplicate commands/mesages
			continue
		}
		lastMsgSeq = msg.SequenceNumber

		decrypted, err := encrypt.DecryptMessage(p.state.CK, p.state.NoncePrefix, p.state.NonceSeq, msg)
		if err != nil {
			log.Fatalf("could not decrypt message: %s", err)
		}
		p.state.NonceSeq++

		cmd, err := command.Unmarshal(decrypted.Payload)
		if err != nil {
			log.Fatalf("could not unmarshal command: %s", err)
		}
		cmdSeq, requestID, err := cmd.GetHeaderData()
		if err != nil {
			log.Fatalf("could not get command header data: %s", err)
		}
		p.state.CmdSeq = cmdSeq
		rsp, err := cmd.GetResponse()
		if err != nil {
			log.Fatalf("could not get command response: %s", err)
		}

		p.state.MsgSeq++
		p.state.CmdSeq++
		p.state.Save()
		responseMetadata := &response.ResponseMetadata{
			Dst:       msg.Source,
			Src:       msg.Destination,
			CmdSeq:    p.state.CmdSeq,
			MsgSeq:    p.state.MsgSeq,
			RequestID: requestID,
			AckSeq:    msg.SequenceNumber + 1,
		}
		msg, err = response.Marshal(rsp, responseMetadata)
		if err != nil {
			log.Fatalf("could not marshal command response: %s", err)
		}
		msg, err = encrypt.EncryptMessage(p.state.CK, p.state.NoncePrefix, p.state.NonceSeq, msg)
		if err != nil {
			log.Fatalf("could not encrypt response: %s", err)
		}
		p.state.NonceSeq++
		p.state.Save()

		log.Tracef("sending response: %s", spew.Sdump(msg))
		p.ble.WriteMessage(msg)

		log.Debugf("reading response ACK. Nonce seq %d", p.state.NonceSeq)
		msg, _ = p.ble.ReadMessage()
		// TODO check for SEQ numbers here and the Ack flag
		decrypted, err = encrypt.DecryptMessage(p.state.CK, p.state.NoncePrefix, p.state.NonceSeq, msg)
		if err != nil {
			log.Fatalf("could not decrypt message: %s", err)
		}
		p.state.NonceSeq++
		if len(decrypted.Payload) != 0 {
			log.Fatalf("this should be empty message with ACK header %s", spew.Sdump(msg))
		}
		p.state.Save()
	}
}
