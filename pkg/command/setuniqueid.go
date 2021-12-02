package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type SetUniqueID struct {
	Seq uint8
	ID  []byte
}

func UnmarshalSetUniqueID(data []byte) (*SetUniqueID, error) {
	ret := &SetUniqueID{}
	// TODO deserialize this command
	log.Infof("SetUniqueID, 0x03, received, data 0x%x", data)
	return ret, nil
}

func (g *SetUniqueID) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.SetUniqueID{}, nil
}

func (g *SetUniqueID) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *SetUniqueID) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}
