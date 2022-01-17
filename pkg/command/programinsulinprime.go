package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramInsulinPrime struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramInsulinPrime(data []byte) (*ProgramInsulinPrime, error) {
	ret := &ProgramInsulinPrime{}
	// TODO deserialize this command
	log.Debugf("ProgramInsulinPrime, 0x1a17, rest of message %x", data)
	return ret, nil
}

func (g *ProgramInsulinPrime) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.StatusResponsePrime{}, nil
}

func (g *ProgramInsulinPrime) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramInsulinPrime) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramInsulinPrime) GetPayload() Payload {
	return nil
}

func (g *ProgramInsulinPrime) GetType() Type {
	return PROGRAM_INSULIN
}
