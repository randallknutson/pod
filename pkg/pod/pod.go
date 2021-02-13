package pod

import (
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
	sp, err := parseStringByte([]string{"SP1=", ",SP2="}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Error reading SP1, SP2 %s", err)
	}
	log.Infof("Received SP1 SP2: %x :: %x", sp["SP1="], sp[",SP2="])

	msg, _ = p.ble.ReadMessage()
	sp, err = parseStringByte([]string{"SPS1="}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Error reading SPS1 %s", err)
	}
	log.Infof("Received SPS1  %x", sp["SPS1="])

	// let's send back what we received, wrong for now
	p.ble.WriteMessage(msg)

	msg, _ = p.ble.ReadMessage()
	sp, err = parseStringByte([]string{"SPS2="}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		log.Fatalf("Error reading SPS2 %s", err)
	}
	log.Infof("Received SPS2  %x", sp["SPS2="])

	// let's send back what we received, wrong for now
	p.ble.WriteMessage(msg)

	// what we want to see in logs:
	//  EapAkaMasterModule Reached Eap-Aka SUCCESSFULLY
}
