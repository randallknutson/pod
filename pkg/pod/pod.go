package pod

import (
	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/avereha/pod/pkg/command"
	"github.com/avereha/pod/pkg/eap"
	"github.com/avereha/pod/pkg/pair"

	"github.com/avereha/pod/pkg/encrypt"
	"github.com/avereha/pod/pkg/response"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type podState int

const (
	podStateNotInitialized podState = iota
	podStateWithID
	podStateWithLTK
	podStateWithCK
)

type Pod struct {
	ble *bluetooth.Ble
	ltk []byte

	id []byte // 4 byte

	msgSeq      uint8  // TODO: is this the same as nonceSeq?
	cmdSeq      uint8  // TODO: are all those 3 the same number ???
	nonceSeq    uint64 // or 16?
	noncePrefix []byte
	ck          []byte
}

func New(ble *bluetooth.Ble) *Pod {
	ret := &Pod{
		ble: ble,
	}
	return ret
}

func (p *Pod) StartActivation() {

	activationCmd, _ := p.ble.ReadCmd()
	log.Infof("got activation command: %s", activationCmd)

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
		log.Fatalf("could not parse SP0GP0: %w", err)
	}

	// send P0 constant
	msg, err = pair.GenerateP0()
	if err != nil {
		log.Fatal(err)
	}
	p.ble.WriteMessage(msg)

	p.ltk, err = pair.LTK()
	if err != nil {
		log.Fatalf("could not get LTK %s", err)
	}
	log.Infof("LTK %x", p.ltk)
	p.EapAka()
	// here we reached Eap AKA!
}

func (p *Pod) EapAka() {

	pair := eap.NewEapAkaChallenge(p.ltk)

	msg, _ := p.ble.ReadMessage()
	err := pair.ParseChallenge(msg)
	if err != nil {
		log.Fatalf("error parsing the EAP-AKA challenge: %s", err)
	}

	msg, err = pair.GenerateChallengeResponse()
	if err != nil {
		log.Fatalf("error generating the eap-aka challenge response")
	}
	p.ble.WriteMessage(msg)

	msg, _ = p.ble.ReadMessage()
	log.Debugf("success? %x", msg.Payload)
	err = pair.ParseSuccess(msg)
	if err != nil {
		log.Fatalf("Error parsing the EAP-AKA Success packet: %s", err)
	}
	p.ck, p.noncePrefix = pair.CKNoncePrefix()
	p.nonceSeq = 1
	p.msgSeq = 1
	log.Infof("got CK: %x", p.ck)
	log.Infof("got Nonce: %x", p.noncePrefix)
	log.Infof("using SEQ: %d", p.nonceSeq)

	p.CommandLoop()
	// ??? Start encryption ???
}

func (p *Pod) CommandLoop() {
	var lastMsgSeq uint8 = 0
	for {
		msg, _ := p.ble.ReadMessage()
		if msg.SequenceNumber == lastMsgSeq {
			// this is a retry because we did not answer yet
			// ignore duplicate commands/mesages
			continue
		}
		lastMsgSeq = msg.SequenceNumber

		log.Tracef("got command message: %s", spew.Sdump(msg))
		decrypted, err := encrypt.DecryptMessage(p.ck, p.noncePrefix, p.nonceSeq, msg)
		if err != nil {
			log.Fatalf("could not decrypt message: %s", err)
		}
		p.nonceSeq++

		cmd, err := command.Unmarshal(decrypted.Payload)
		if err != nil {
			log.Fatalf("could not unmarshal command: %s", err)
		}
		cmdSeq, requestID, err := cmd.GetHeaderData()
		if err != nil {
			log.Fatalf("could not get command header data", err)
		}
		p.cmdSeq = cmdSeq
		rsp, err := cmd.GetResponse()
		if err != nil {
			log.Fatalf("could not get command response: %s", err)
		}

		p.msgSeq++
		p.cmdSeq++
		responseMetadata := &response.ResponseMetadata{
			Dst:       msg.Source,
			Src:       msg.Destination,
			CmdSeq:    p.cmdSeq,
			MsgSeq:    p.msgSeq,
			RequestID: requestID,
			AckSeq:    msg.SequenceNumber + 1,
		}
		msg, err = response.Marshal(rsp, responseMetadata)
		if err != nil {
			log.Fatalf("could not marshal command response: %s", err)
		}
		msg, err = encrypt.EncryptMessage(p.ck, p.noncePrefix, p.nonceSeq, msg)
		if err != nil {
			log.Fatalf("could not encrypt response: %s", err)
		}
		p.nonceSeq++

		p.ble.WriteMessage(msg)
		log.Tracef("sending response: %s", spew.Sdump(msg))

		log.Trace("reading response ACK. Nonce seq %d", p.nonceSeq)
		msg, _ = p.ble.ReadMessage()
		// TODO check for SEQ numbers here and the Ack flag
		decrypted, err = encrypt.DecryptMessage(p.ck, p.noncePrefix, p.nonceSeq, msg)
		if err != nil {
			log.Fatalf("could not decrypt message: %s", err)
		}
		p.nonceSeq++
		if len(decrypted.Payload) != 0 {
			log.Fatalf("this should be empty message with ACK header %s", spew.Sdump(msg))
		}
	}
}
