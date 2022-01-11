package response

import (
	"encoding/hex"
)

type Type2StatusResponse struct {
	Seq uint16
}

func (r *Type2StatusResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("021602080200000001b200000003ff01cc0000001fff030d")

	return response, nil
}
