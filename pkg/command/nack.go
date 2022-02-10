package command

import (
	"github.com/avereha/pod/pkg/response"

	log "github.com/sirupsen/logrus"
)

type Nack struct {
	Seq uint8
	ID  []byte
}

func UnmarshalNack(data []byte) (*Nack, error) {
	ret := &Nack{}
	log.Infof("unmarshal Nack. we do not implement this command yet, data %x", data)
	return ret, nil
}

func (g *Nack) IsResponseHardcoded() bool {
	return true
}

func (g *Nack) GetSeq() uint8 {
	return g.Seq
}

func (g *Nack) GetResponse() (response.Response, error) {
	return &response.NackResponse{}, nil
}

func (g *Nack) DoesMutatePodState() bool {
	return false
}

func (g *Nack) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *Nack) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *Nack) GetPayload() Payload {
	return nil
}

func (g *Nack) GetType() Type {
	return NACK
}
