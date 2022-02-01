package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramPostInsulinInsert struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramPostInsert(data []byte) (*ProgramPostInsulinInsert, error) {
	ret := &ProgramPostInsulinInsert{}
	// TODO deserialize this command
	log.Debugf("ProgramPostInsulinInsert, next message that expects an 0x1d, rest of message %x", data)
	return ret, nil
}

func (g *ProgramPostInsulinInsert) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.StatusResponsePostInsert{}, nil
}

func (g *ProgramPostInsulinInsert) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramPostInsulinInsert) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramPostInsulinInsert) GetPayload() Payload {
	return nil
}

func (g *ProgramPostInsulinInsert) GetType() Type {
	return PROGRAM_INSULIN
}
