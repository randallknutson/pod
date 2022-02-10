package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetStatus struct {
	Seq         uint8
	ID          []byte
	RequestType byte
}

func UnmarshalGetStatus(data []byte) (*GetStatus, error) {
	ret := &GetStatus{}

	ret.RequestType = data[1]
	// TODO deserialize this command
	log.Debugf("GetStatus, 0x0e, received, data %x", data)

	return ret, nil
}

func (g *GetStatus) GetSeq() uint8 {
	return g.Seq
}

func (g *GetStatus) IsResponseHardcoded() bool {
	if g.RequestType == 0 || g.RequestType == 7 || g.RequestType == 2 {
		return false
	} else {
		return true
	}
}

func (g *GetStatus) DoesMutatePodState() bool {
	return false
}

// TODO remove this once all other message types return something other than
// Hardcoded for GetResponseType()
func (g *GetStatus) GetResponse() (response.Response, error) {
	if g.RequestType == 0x2 {
		return &response.DetailedStatusResponse{}, nil
	} else if g.RequestType == 0x46 {
		return &response.Type46StatusResponse{}, nil
	} else if g.RequestType == 0x50 {
		return &response.Type50StatusResponse{}, nil
	} else if g.RequestType == 0x51 {
		return &response.Type51StatusResponse{}, nil
	} else {
		return &response.NackResponse{}, nil
	}
}

func (g *GetStatus) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetStatus) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *GetStatus) GetPayload() Payload {
	return nil
}

func (g *GetStatus) GetType() Type {
	return GET_STATUS
}
