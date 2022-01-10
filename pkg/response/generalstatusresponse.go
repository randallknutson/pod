package response

import (
	"encoding/hex"
)

// This is the default for most 0x1d response
//   Note - this response cannot have immediate_bolus_active =   True
//          if true, Loop will refuse to bolus and won't see green loop with simulator
// Special cases are in statusresponseXXX modules
//   The PodProgress is updated in the command pkg in command.go
//   PodProgress = 2: respond to 0x07 with 0x0115, versionresponse.go
//   PodProgress = 3: respond to 0x03 with 0x011b, setuniqueidresponse.go
//   PodProgress = 3: respond to 0x19 with 0x1d, generalstatusresponsebeforeprime.go
//   PodProgress = 4: response to prime command, statusresponseprime.og
//   PodProgress = 6: response to basal rate program, statusreponseschedule.go
//   PodProgress = 7: response to insert command, statusresponseinsert.go
//   PodProgress = 8: all other 0x1d responses, except deactivate, type2 and type3
//   Default        : this function, generalstatusresponse.go

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
