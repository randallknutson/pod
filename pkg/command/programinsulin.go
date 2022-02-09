package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramInsulin struct {
	Seq uint8
	ID  []byte
	TableNum byte
	Pulses uint16
}

func UnmarshalProgramInsulin(data []byte) (*ProgramInsulin, error) {
	ret := &ProgramInsulin{}
	// TODO deserialize this command
	log.Debugf("ProgramInsulin, 0x1a, received, data %x", data)

	// 1a LL NNNNNNNN 02 CCCC HH SSSS PPPP 0ppp
	//    00 01020304 05 0607 08 0910 1112 1314
	ret.TableNum = data[5]
	ret.Pulses = (uint16(data[11]) << 8) + uint16(data[12])
	return ret, nil
}

func (g *ProgramInsulin) GetResponseType() (CommandResponseType) {
	return Hardcoded
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
