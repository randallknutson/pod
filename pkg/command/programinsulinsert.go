package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramInsulinInsert struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramInsulinInsert(data []byte) (*ProgramInsulinInsert, error) {
	ret := &ProgramInsulinInsert{}
	// TODO deserialize this command
	log.Debugf("ProgramInsulinInsert, 0x1a17, rest of message %x", data)
	return ret, nil
}

func (g *ProgramInsulinInsert) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.StatusResponseInsert{}, nil
}

func (g *ProgramInsulinInsert) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramInsulinInsert) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramInsulinInsert) GetPayload() Payload {
	return nil
}

func (g *ProgramInsulinInsert) GetType() Type {
	return PROGRAM_INSULIN
}
