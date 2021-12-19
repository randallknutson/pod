package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type ProgramAlertsBeforePrime struct {
	Seq uint8
	ID  []byte
}

func UnmarshalProgramAlertsBeforePrime(data []byte) (*ProgramAlertsBeforePrime, error) {
	ret := &ProgramAlertsBeforePrime{}
	// TODO deserialize this command
	log.Debugf("ProgramAlertsBeforePrime, 0x19, received, data %x", data)
	return ret, nil
}

func (g *ProgramAlertsBeforePrime) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponseBeforePrime{}, nil
}

func (g *ProgramAlertsBeforePrime) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *ProgramAlertsBeforePrime) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}
