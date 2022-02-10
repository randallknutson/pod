package command

import (
	"fmt"

	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetVersion struct {
	Seq        uint8
	ID         []byte
	TheOtherID []byte
}

func UnmarshalGetVersion(data []byte) (*GetVersion, error) {
	if data[0] != 4 {
		return nil, fmt.Errorf("invalid length when unmarshaling GetVersion %d :: %x", data[0], data)
	}
	ret := &GetVersion{}
	ret.TheOtherID = make([]byte, 4)
	copy(ret.TheOtherID, data[1:])
	log.Debugf("GetVersion, 0x07, rest of podMsgBody: %x", data)
	return ret, nil
}

func (g *GetVersion) GetSeq() uint8 {
	return g.Seq
}

func (g *GetVersion) IsResponseHardcoded() bool {
	return true
}

func (g *GetVersion) DoesMutatePodState() bool {
	return false
}

func (g *GetVersion) GetResponse() (response.Response, error) {
	// TODO improve responses
	return &response.VersionResponse{}, nil
}

func (g *GetVersion) SetHeaderData(seq uint8, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}

func (g *GetVersion) GetHeaderData() (uint8, []byte, error) {
	return g.Seq, g.ID, nil
}

func (g *GetVersion) GetPayload() Payload {
	return nil
}

func (g *GetVersion) GetType() Type {
	return GET_VERSION
}
