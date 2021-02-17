package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type Nack struct {
	Seq uint16
	Id  []byte
}

func UnmarshalNack(data []byte) (*Nack, error) {
	ret := &Nack{}
	log.Debugf("Unmarshal Nack: %x", data)
	return ret, nil
}

func (g *Nack) GetResponse() (response.Response, error) {
	return nil, nil
}
func (g *Nack) SetHeaderData(seq uint16, id []byte) error {
	g.Id = id
	g.Seq = seq
	return nil
}
