package response

import (
	"encoding/hex"
)

type Type46StatusResponse struct {
	Seq uint16
}

func (r *Type46StatusResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("0204460000")
	return response, nil
}
