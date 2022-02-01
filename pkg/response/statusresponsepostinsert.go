package response

import (
	"encoding/hex"
)

type StatusResponsePostInsert struct {
	Seq uint16
}

func (r *StatusResponsePostInsert) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D18001DE00000001BFF") // Insert
	return response, nil
}
