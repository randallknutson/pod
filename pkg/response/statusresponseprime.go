package response

import (
	"encoding/hex"
)

// This is the special case - sent in response to the prime command

type StatusResponsePrime struct {
	Seq uint16
}

func (r *StatusResponsePrime) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D4400004034000003FF") // Prime
	return response, nil
}
