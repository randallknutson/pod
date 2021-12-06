package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type StopDelivery struct {
	Seq uint8
	ID  []byte
}

func UnmarshalStopDelivery(data []byte) (*StopDelivery, error) {
	ret := &StopDelivery{}
	// TODO deserialize this command
	log.Debugf("StopDelivery, 0x1f, received, data 0x%x", data)
	return ret, nil
}

func (g *StopDelivery) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *StopDelivery) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *StopDelivery) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}
