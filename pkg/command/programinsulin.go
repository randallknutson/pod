package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramInsulin struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramInsulin(data []byte) (*ProgramInsulin, error) {
	ret := &ProgramInsulin{}
	// TODO deserialize this command
	log.Debugf("ProgramInsulin, 0x1a, received, data %x", data)
	return ret, nil
}

func (g *ProgramInsulin) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *ProgramInsulin) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramInsulin) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramInsulin) GetPayload() Payload {
	return nil
}

func (g *ProgramInsulin) GetType() Type {
	return PROGRAM_INSULIN
}
