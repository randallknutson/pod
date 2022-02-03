package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramAlertsBeforeInsert struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramAlertsBeforeInsert(data []byte) (*ProgramAlertsBeforeInsert, error) {
	ret := &ProgramAlertsBeforeInsert{}
	// TODO deserialize this command
	log.Debugf("ProgramAlertsBeforeInsert, 0x19, received, data %x", data)
	return ret, nil
}

func (g *ProgramAlertsBeforeInsert) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponseBeforeInsert{}, nil
}

func (g *ProgramAlertsBeforeInsert) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramAlertsBeforeInsert) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *ProgramAlertsBeforeInsert) GetPayload() Payload {
	return nil
}

func (g *ProgramAlertsBeforeInsert) GetType() Type {
	return PROGRAM_ALERTS
}
