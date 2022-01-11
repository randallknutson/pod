package response

import (
	"encoding/hex"
)

// This is the special case - sent with the 0x011B response to 0x03 message

type SetUniqueID struct {
	Seq uint16
}

func (r *SetUniqueID) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("011B13881008340A50040A00010300040308146DB10006E45100001091")

	return response, nil
}
