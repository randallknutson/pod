package response

import (
	"encoding/hex"
)

type GeneralStatusResponse struct {
	Seq uint16
}

func (r *GeneralStatusResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D1800A02800000463FF")

	return response, nil
}
