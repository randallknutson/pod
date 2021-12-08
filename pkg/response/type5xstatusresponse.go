package response

import (
	"encoding/hex"
)

type Type5xStatusResponse struct {
	Seq uint16
}

func (r *Type5xStatusResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("02cb5000900063298005622f80086229800d622f80")

	return response, nil
}
