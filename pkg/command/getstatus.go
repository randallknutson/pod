package command

import (
	"github.com/avereha/pod/pkg/response"
	log "github.com/sirupsen/logrus"
)

type GetStatus struct {
	Seq uint8
	ID  []byte
}

func UnmarshalGetStatus(data []byte) (*GetStatus, error) {
	ret := &GetStatus{}
	// TODO deserialize this command
	log.Debugf("GetStatus, 0x0e, received, data %x", data)
	return ret, nil
}

func (g *GetStatus) GetResponseType() (CommandResponseType) {
	return Dynamic
}

// TODO remove this once all other message types return something other than
// Hardcoded for GetResponseType()
func (g *GetStatus) GetResponse() (response.Response, error) {
	return &response.GeneralStatusResponse{}, nil
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
