package response

import (
	"encoding/hex"
)

type NackResponse struct {
	Seq uint16
}

func (r *NackResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("0603070009")

	return response, nil
}
