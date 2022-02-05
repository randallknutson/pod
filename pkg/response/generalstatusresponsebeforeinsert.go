package response

import (
	"encoding/hex"
)

// This is the special case - sent after config and before prime

type GeneralStatusResponseBeforeInsert struct {
	Seq uint16
}

func (r *GeneralStatusResponseBeforeInsert) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1d160017600000000BFF") // basal active, pod progress 6, insulin=$17<<2=$2E pulses (46 x 0.05U = 2.3U0, sequence = $C, pod alive time = 2 minutes
	return response, nil
}
