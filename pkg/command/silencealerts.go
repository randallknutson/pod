package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

	type SilenceAlerts struct {
	Seq uint8
	ID  []byte
}

func UnmarshalSilenceAlerts(data []byte) (*SilenceAlerts, error) {
	ret := &SilenceAlerts{}
	// TODO deserialize this command
	log.Infof("SilenceAlerts, 0x11, received, data 0x%x", data)
	return ret, nil
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
