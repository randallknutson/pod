package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetType5xStatus struct {
	Seq uint8
	ID  []byte
}

func UnmarshalType5xStatus(data []byte) (*GetType5xStatus, error) {
	ret := &GetType5xStatus{}
	log.Debugf("GetStatus, 0x0e, received, type %x, data %x", data[1], data)
	return ret, nil
}

func (g *GetType5xStatus) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.Type5xStatusResponse{}, nil
}

func (g *GetType5xStatus) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetType5xStatus) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}
