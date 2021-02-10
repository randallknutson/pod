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
	activationCmd, _ := p.ble.Read()
	log.Infof("Got activation command: %s", activationCmd)
}
