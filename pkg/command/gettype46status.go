package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetType46Status struct {
	Seq uint8
	ID  []byte
}

func UnmarshalType46Status(data []byte) (*GetType46Status, error) {
	ret := &GetType46Status{}
	log.Debugf("GetStatus, 0x0e, received, type %x, data %x", data[1], data)
	return ret, nil
}

func (g *GetType46Status) GetResponseType() (CommandResponseType) {
	return Hardcoded
}

func (g *GetType46Status) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.Type46StatusResponse{}, nil
}

func (g *GetType46Status) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetType46Status) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *GetType46Status) GetPayload() Payload {
	return nil
}

func (g *GetType46Status) GetType() Type {
	return GET_STATUS
}
