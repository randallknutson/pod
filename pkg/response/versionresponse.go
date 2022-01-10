package response

import (
	"encoding/hex"
)

// This is the special case - sent with the 0x0115 response to 0x07 message

type VersionResponse struct {
	Seq uint16
}

func (r *VersionResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("0115040A00010300040208146DB10006E45100FFFFFFFF")

	return response, nil
}
