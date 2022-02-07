package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetType2Status struct {
	Seq uint8
	ID  []byte
}

func UnmarshalType2Status(data []byte) (*GetType2Status, error) {
	ret := &GetType2Status{}
	log.Debugf("GetStatus, 0x0e, received, type %x, data %x", data[1], data)
	return ret, nil
}

func (g *GetType2Status) GetResponseType() (CommandResponseType) {
	return Hardcoded
}

func (g *GetType2Status) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.Type2StatusResponse{}, nil
}

func (g *GetType2Status) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetType2Status) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *GetType2Status) GetPayload() Payload {
	return nil
}

func (g *GetType2Status) GetType() Type {
	// TODO: Differentiate between normal, type 2 and type 5 get status
	return GET_STATUS
}
