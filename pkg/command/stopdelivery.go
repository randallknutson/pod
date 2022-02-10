package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type StopDelivery struct {
	Seq           uint8
	ID            []byte
	StopBolus     bool
	StopTempBasal bool
	StopBasal     bool
}

func UnmarshalStopDelivery(data []byte) (*StopDelivery, error) {
	ret := &StopDelivery{
		StopBolus:     (data[5] & 0b100) != 0,
		StopTempBasal: (data[5] & 0b10) != 0,
		StopBasal:     (data[5] & 0b1) != 0,
	}
	// 05 49 4e 53 2e 07 1910494e532e580f000f06046800001e0302
	log.Debugf("StopDelivery, 0x1f, received, data %x, stop_bits = %x", data, data[5]&0b1)
	return ret, nil
}

func (g *StopDelivery) GetSeq() uint8 {
	return g.Seq
}

func (g *StopDelivery) IsResponseHardcoded() bool {
	return false
}

func (g *StopDelivery) DoesMutatePodState() bool {
	return true
}

func (g *StopDelivery) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.GeneralStatusResponse{}, nil
}

func (g *StopDelivery) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *StopDelivery) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *StopDelivery) GetPayload() Payload {
	return nil
}

func (g *StopDelivery) GetType() Type {
	return STOP_DELIVERY
}
