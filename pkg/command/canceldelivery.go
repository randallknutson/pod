package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type CancelDelivery struct {
	Seq uint8
	ID  []byte
}

func UnmarshalCancelDelivery(data []byte) (*CancelDelivery, error) {
	ret := &CancelDelivery{}
	// TODO deserialize this command
	log.Infof("CancelDelivery, 0x1f, received, data 0x%x", data)
	return ret, nil
}

func (g *CancelDelivery) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *CancelDelivery) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *CancelDelivery) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}
