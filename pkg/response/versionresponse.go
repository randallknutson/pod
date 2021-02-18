package response

import (
	"encoding/hex"
)

type VersionResponse struct {
	Seq uint16
}

func (r *VersionResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("0115040A00010300040208146CC1000954D400FFFFFFFF")

	return response, nil
}
