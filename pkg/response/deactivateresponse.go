package response

import (
	"encoding/hex"
)

type DeactivateResponse struct {
	Seq uint16
}

func (r *DeactivateResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D0F050648000038B6F3")

	return response, nil
}
