package pod

import (
	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/avereha/pod/pkg/eap"

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
	ble   *bluetooth.Ble
	ltk   []byte
	id    []byte // 4 byte
	ck    []byte
	nonce []byte
	seq   uint32 // or 16?
}

func New(ble *bluetooth.Ble) *Pod {
	ret := &Pod{
		ble: ble,
	}
	return ret
}

func (p *Pod) StartActivation() {

	activationCmd, _ := p.ble.ReadCmd()
	log.Infof("Got activation command: %s", activationCmd)

	pair := &Pair{}
	msg, _ := p.ble.ReadMessage()
	if err := pair.ParseSP1SP2(msg); err != nil {
		log.Fatalf("Error parsing SP1SP2 %s", err)
	}
	// read PDM public key and nonce
	msg, _ = p.ble.ReadMessage()
	if err := pair.ParseSPS1(msg); err != nil {
		log.Fatalf("Error parsing SPS1 %s", err)
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
	if string(msg.Payload) != sp0gp0 {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Expected SP0GP0, got %x", msg.Payload)
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
		log.Fatalf("Error parsing the EAP-AKA challenge")
	}

	msg, err = pair.GenerateChallengeResponse()
	if err != nil {
		log.Fatalf("error generating the eap-aka challenge response")
	}

	msg, _ = p.ble.ReadMessage()
	err = pair.ParseSuccess(msg)
	if err != nil {
		log.Fatalf("Error parsing the EAP-AKA Success packet")
	}
	ck, nonce := pair.CKNonce()
	log.Infof("Got CK: %x", ck)
	log.Infof("Got Nonce: %x", nonce)

	// ??? Start encryption ???
}
