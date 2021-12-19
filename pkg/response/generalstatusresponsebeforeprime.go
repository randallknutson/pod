package response

import (
	"encoding/hex"
)

// This is the default for most 0x1d response
// Special cases are in statusresponseXXX modules
//   PodProgress = 4: response, _ := hex.DecodeString("1D4400004034000003FF")
//   PodProgress = 6: response, _ := hex.DecodeString("1D160016D000400023FF")
//   PodProgress = 7: response, _ := hex.DecodeString("1D570016F011000023FF")
//   PodProgress = 8: response, _ := hex.DecodeString("1D18001F7000000023FF")
//   Default        : response, _ := hex.DecodeString("1D1800A02800000463FF")

type GeneralStatusResponseBeforePrime struct {
	Seq uint16
}

func (r *GeneralStatusResponseBeforePrime) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D0300003000000003FF") // Default
	return response, nil
}
