package pod

import (
	"github.com/avereha/pod/pkg/bluetooth"
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
	log.Infof("Received message: %s", msg)

	msg, _ = p.ble.ReadMessage()
	log.Infof("Received message2: %s", msg)

	// let's send back what we received, wrong for now
	p.ble.WriteMessage(msg)

	msg, _ = p.ble.ReadMessage()
	log.Infof("Received message 3: %s", msg)

	// let's send back what we received, wrong for now
	p.ble.WriteMessage(msg)

	// what we want to see in logs:
	//  EapAkaMasterModule Reached Eap-Aka SUCCESSFULLY
}
