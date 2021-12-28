package response

import (
	"encoding/hex"
)

// This is the default for most 0x1d response
//   Note - this response cannot have immediate_bolus_active =   True
//          if true, Loop will refuse to bolus and won't see green loop with simulator
// Special cases are in statusresponseXXX modules
//   PodProgress = 4: response, _ := hex.DecodeString("1D4400004034000003FF")
//   PodProgress = 6: response, _ := hex.DecodeString("1D160016D000400023FF")
//   PodProgress = 7: response, _ := hex.DecodeString("1D570016F011000023FF")
//   PodProgress = 8: response, _ := hex.DecodeString("1D18001F7000000023FF")
//   Default        : response, _ := hex.DecodeString("1D1800A02800000463FF")

// From actual pod messages, expected CRC: 0x00FB for seqNumber = 9 for a msgBody of:
//    "1d58001cc014000013ff"

type GeneralStatusResponse struct {
	Seq uint16
}

func (r *GeneralStatusResponse) Marshal() ([]byte, error) {
	// response, _ := hex.DecodeString("1d58001cc014000013ff") // immediate_bolus_active is true

	response, _ := hex.DecodeString("1D1800A02800000463FF") // Default
	return response, nil
}
