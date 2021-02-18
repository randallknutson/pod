package response

import (
	"encoding/hex"
)

type SetUniqueID struct {
	Seq uint16
}

func (r *SetUniqueID) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("011B13881008340A50040A00010300040308146CC1000954D402420001")

	return response, nil
}
