package response

import (
	"encoding/hex"
)

// This is the special case - sent after config and before prime

type GeneralStatusResponseBeforePrime struct {
	Seq uint16
}

func (r *GeneralStatusResponseBeforePrime) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D0300002000000003FF") // Default
	return response, nil
}
