package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type SilenceAlerts struct {
	Seq       uint8
	ID        []byte
	AlertMask uint8
}

func UnmarshalSilenceAlerts(data []byte) (*SilenceAlerts, error) {
	ret := &SilenceAlerts{}
	ret.AlertMask = data[5]
	log.Debugf("SilenceAlerts, 0x11, received, alert mask %x", ret.AlertMask)
	return ret, nil
}

func (g *SilenceAlerts) GetSeq() uint8 {
	return g.Seq
}

func (g *SilenceAlerts) IsResponseHardcoded() bool {
	return false
}

func (g *SilenceAlerts) DoesMutatePodState() bool {
	return true
}

func (g *SilenceAlerts) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *SilenceAlerts) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *SilenceAlerts) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *SilenceAlerts) GetPayload() Payload {
	return nil
}

func (g *SilenceAlerts) GetType() Type {
	return SILENCE_ALERTS
}
