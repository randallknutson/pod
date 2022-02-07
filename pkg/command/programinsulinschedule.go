package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramInsulinSchedule struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramInsulinSchedule(data []byte) (*ProgramInsulinSchedule, error) {
	ret := &ProgramInsulinSchedule{}
	// TODO deserialize this command
	log.Debugf("ProgramInsulinSchedule, 0x1a13, rest of message %x", data)
	return ret, nil
}

func (g *ProgramInsulinSchedule) GetResponseType() (CommandResponseType) {
	return Hardcoded
}

func (g *ProgramInsulinSchedule) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.StatusResponseSchedule{}, nil
}

func (g *ProgramInsulinSchedule) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramInsulinSchedule) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramInsulinSchedule) GetPayload() Payload {
	return nil
}

func (g *ProgramInsulinSchedule) GetType() Type {
	return PROGRAM_INSULIN
}
