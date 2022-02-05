package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetType51Status struct {
	Seq uint8
	ID  []byte
}

func UnmarshalType51Status(data []byte) (*GetType51Status, error) {
	ret := &GetType51Status{}
	log.Debugf("GetStatus, 0x0e, received, type %x, data %x", data[1], data)
	return ret, nil
}

func (g *GetType51Status) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.Type51StatusResponse{}, nil
}

func (g *GetType51Status) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetType51Status) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *GetType51Status) GetPayload() Payload {
	return nil
}

func (g *GetType51Status) GetType() Type {
	// TODO: Differentiate between normal, type 2 and type 5 get status
	return GET_STATUS
}
