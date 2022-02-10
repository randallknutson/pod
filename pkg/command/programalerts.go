package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramAlerts struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramAlerts(data []byte) (*ProgramAlerts, error) {
	ret := &ProgramAlerts{}
	// TODO deserialize this command
	log.Debugf("ProgramAlerts, 0x19, received, data %x", data)
	return ret, nil
}

func (g *ProgramAlerts) GetSeq() uint8 {
	return g.Seq
}

func (g *ProgramAlerts) IsResponseHardcoded() bool {
	return false
}

func (g *ProgramAlerts) DoesMutatePodState() bool {
	return true
}

func (g *ProgramAlerts) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *ProgramAlerts) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramAlerts) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramAlerts) GetPayload() Payload {
	return nil
}

func (g *ProgramAlerts) GetType() Type {
	return PROGRAM_ALERTS
}
