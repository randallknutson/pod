package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetType50Status struct {
	Seq uint8
	ID  []byte
}

func UnmarshalType50Status(data []byte) (*GetType50Status, error) {
	ret := &GetType50Status{}
	log.Debugf("GetStatus, 0x0e, received, type %x, data %x", data[1], data)
	return ret, nil
}

func (g *GetType50Status) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.Type50StatusResponse{}, nil
}

func (g *GetType50Status) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetType50Status) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *GetType50Status) GetPayload() Payload {
	return nil
}

func (g *GetType50Status) GetType() Type {
	// TODO: Differentiate between normal, type 2 and type 5 get status
	return GET_STATUS
}
