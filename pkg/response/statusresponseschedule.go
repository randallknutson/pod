package response

import (
	"encoding/hex"
)

// This is the special case - sent after the initial basal schedule is programmed

type StatusResponseSchedule struct {
	Seq uint16
}

func (r *StatusResponseSchedule) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D160016D000000023FF") // Schedule
	return response, nil
}
