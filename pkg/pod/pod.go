package pod

import (
	"bytes"

	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type podState int

const (
	podStateNotInitialized podState = iota
	podStateWithID
	podStateWithLTK
	podStateWithCK

	sp1 = "SP1="
	sp2 = ",SP2="

	sps1   = "SPS1="
	sps2   = "SPS2="
	sp0gp0 = "SP0,GP0"
	p0     = "P0="
)

type Pod struct {
	ble *bluetooth.Ble
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

	msg, _ := p.ble.ReadMessage()
	sp, err := parseStringByte([]string{sp1, sp2}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Error reading SP1, SP2 %s", err)
	}
	log.Infof("Received SP1 SP2: %x :: %x", sp[sp1], sp[sp2])

	msg, _ = p.ble.ReadMessage()
	sp, err = parseStringByte([]string{sps1}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Error reading SPS1 %s", err)
	}
	log.Infof("Received SPS1  %x", sp[sps1])
	pdmPublic := sp[sps1][:32]
	pdmNonce := sp[sps1][32:]

	podSecret, podPublic, podNonce, err := generateSPS1()
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	buf.Write(podPublic)
	buf.Write(podNonce)
	sp = make(map[string][]byte)
	sp[sps1] = buf.Bytes()
	msg.Payload, err = buildStringByte([]string{sps1}, sp)
	if err != nil {
		log.Fatal(err)
	}
	// send pod public key
	p.ble.WriteMessage(msg)

	ltk, err := getLTK(pdmPublic, podSecret)
	if err != nil {
		log.Fatalf("LTK: %s", err)
	}
	log.Infof("PDM public %x", pdmPublic)
	log.Infof("PDM nonce: %x", pdmNonce)
	log.Infof("POD public: %x", podPublic)
	log.Infof("POD secret: %x", podSecret)
	log.Infof("POD Nonce: %x", podNonce)
	log.Infof("LTK: %x", ltk)

	msg, _ = p.ble.ReadMessage()
	sp, err = parseStringByte([]string{sps2}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Error reading SPS2 %s", err)
	}
	log.Infof("Received SPS2  %x", sp[sps2])

	// let's send back what we received, wrong for now
	p.ble.WriteMessage(msg)

	msg, _ = p.ble.ReadMessage()
	if string(msg.Payload) != sp0gp0 {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Expected SP0GP0, got %x", msg.Payload)
	}

	sp = make(map[string][]byte)
	sp[p0] = []byte{0xa5} // magic constant ???
	msg.Payload, err = buildStringByte([]string{p0}, sp)
	if err != nil {
		log.Fatal(err)
	}
	// send p0
	p.ble.WriteMessage(msg)
	// what we want to see in logs:
	//  EapAkaMasterModule Reached Eap-Aka SUCCESSFULLY
}
