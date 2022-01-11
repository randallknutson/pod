package response

import (
	"encoding/hex"
)

type StatusResponseInsert struct {
	Seq uint16
}

func (r *StatusResponseInsert) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D570016F011000023FF") // Insert
	return response, nil
}
