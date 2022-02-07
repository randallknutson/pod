package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramPrimeComplete struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramPrimeComplete(data []byte) (*ProgramPrimeComplete, error) {
	ret := &ProgramPrimeComplete{}
	// TODO deserialize this command
	log.Debugf("ProgramPrimeComplete, rest of message %x", data)
	return ret, nil
}

func (g *ProgramPrimeComplete) GetResponseType() (CommandResponseType) {
	return Hardcoded
}

func (g *ProgramPrimeComplete) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.StatusResponsePrimeComplete{}, nil
}

func (g *ProgramPrimeComplete) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramPrimeComplete) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramPrimeComplete) GetPayload() Payload {
	return nil
}

func (g *ProgramPrimeComplete) GetType() Type {
	return PROGRAM_INSULIN
}
