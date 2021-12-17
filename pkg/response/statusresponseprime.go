package response

import (
	"encoding/hex"
)

type StatusResponsePrime struct {
	Seq uint16
}

func (r *StatusResponsePrime) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D4400004034000003FF") // Prime
	return response, nil
}
