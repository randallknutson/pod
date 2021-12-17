package response

import (
	"encoding/hex"
)

type StatusResponseSchedule struct {
	Seq uint16
}

func (r *StatusResponseSchedule) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D160016D000400023FF") // Schedule
	return response, nil
}
