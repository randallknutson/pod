
package response

import (
	"encoding/hex"
)

type StatusResponsePrimeComplete struct {
	Seq uint16
}

func (r *StatusResponsePrimeComplete) Marshal() ([]byte, error) {
	// no active delivery, pod progress 5, insulin deliveried = ($17<<1)=$2E (46 x 0.05U = 2.3U), ssss=(4<<1)=8
	response, _ := hex.DecodeString("1D0500174000000007FF")
	return response, nil
}
