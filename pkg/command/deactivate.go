package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type Deactivate struct {
	Seq uint8
	ID  []byte
}

func UnmarshalDeactivate(data []byte) (*Deactivate, error) {
	ret := &Deactivate{}
	// TODO deserialize this command
	log.Debugf("Deactivate, 0x1c, received, data %x", data)
	return ret, nil
}

func (g *Deactivate) GetSeq() uint8 {
	return g.Seq
}

func (g *Deactivate) IsResponseHardcoded() bool {
	return true
}

func (g *Deactivate) DoesMutatePodState() bool {
	return true
}

func (g *Deactivate) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.DeactivateResponse{}, nil
}

func (g *Deactivate) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *Deactivate) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *Deactivate) GetPayload() Payload {
	return nil
}

func (g *Deactivate) GetType() Type {
	return DEACTIVATE
}
