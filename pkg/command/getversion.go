package command

import (
	"fmt"

	"github.com/avereha/pod/pkg/response"
)

type GetVersion struct {
	Seq        uint16
	ID         []byte
	TheOtherID []byte
}

func UnmarshalGetVersion(data []byte) (*GetVersion, error) {
	if data[0] != 0 {
		return nil, fmt.Errorf("invalid length when unmarshaling GetVersion %d :: %x", data[0], data)
	}
	ret := &GetVersion{}
	ret.TheOtherID = make([]byte, 4)
	copy(ret.TheOtherID, data[1:])
	return ret, nil
}

func (g *GetVersion) GetResponse() (response.Response, error) {
	// TODO improve responses

	return nil, nil
}

func (g *GetVersion) SetHeaderData(seq uint16, id []byte) error {
	g.ID = id
	g.Seq = seq
	return nil
}
