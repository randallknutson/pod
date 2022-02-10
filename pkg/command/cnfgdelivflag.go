package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type CnfgDelivFlag struct {
	Seq uint8
	ID  []byte
}

func UnmarshalCnfgDelivFlag(data []byte) (*CnfgDelivFlag, error) {
	ret := &CnfgDelivFlag{}
	// TODO deserialize this command
	log.Debugf("CnfgDelivFlag, 0x08, received, data %x", data)
	return ret, nil
}

func (g *CnfgDelivFlag) GetSeq() uint8 {
	return g.Seq
}

func (g *CnfgDelivFlag) IsResponseHardcoded() bool {
	return true
}

func (g *CnfgDelivFlag) DoesMutatePodState() bool {
	return false
}

func (g *CnfgDelivFlag) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *CnfgDelivFlag) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *CnfgDelivFlag) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *CnfgDelivFlag) GetPayload() Payload {
	return nil
}

func (g *CnfgDelivFlag) GetType() Type {
	return CNFG_DELIV_FLAG
}
