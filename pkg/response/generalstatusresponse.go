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

type PodProgress int8

const (
	PodProgressInitial                = 0
	PodProgressMemoryInitialized      = 1
	PodProgressReminderInitialized    = 2
	PodProgressPairingCompleted       = 3
	PodProgressPriming                = 4
	PodProgressPrimingCompleted       = 5
	PodProgressBasalInitialized       = 6
	PodProgressInsertingCannula       = 7
	PodProgressRunningAbove50U        = 8
	PodProgressRunningBelow50U        = 9
	PodProgressNotUsed10              = 10
	PodProgressNotUsed11              = 11
	PodProgressNotUsed12              = 12
	PodProgressFault                  = 13
	PodProgressActivationTimeExceeded = 14
	PodProgressPodInactive            = 15
)

type GeneralStatusResponse struct {
	Seq                 uint16
	Alerts              uint8
	BolusActive         bool
	TempBasalActive     bool
	BasalActive         bool
	ExtendedBolusActive bool
	PodProgress         PodProgress
	Delivered           uint16
	BolusRemaining      uint16
	IsFaulted           bool
	MinutesActive       uint16
	Reservoir           uint16
	LastProgSeqNum      uint8
}

func (r *GeneralStatusResponse) Marshal() ([]byte, error) {
	response, _ := hex.DecodeString("1D1800A02800000463FF") // Default

	// Delivery bits
	response[1] = response[1] & 0b1111
	if r.ExtendedBolusActive {
		response[1] = response[1] | (1 << 7)
	}
	if r.BolusActive {
		response[1] = response[1] | (1 << 6)
	}
	if r.TempBasalActive {
		response[1] = response[1] | (1 << 5)
	}
	if r.BasalActive {
		response[1] = response[1] | (1 << 4)
	}

	// PodProgress
	response[1] = response[1]&0b11110000 | (byte(r.PodProgress) & 0b1111)

	// Total insulin delivered
	response[2] = byte(r.Delivered >> 9)
	response[3] = byte((r.Delivered >> 1) & 0xff)
	response[4] = response[4]&0b11111110 | uint8((r.Delivered&0b1)<<7)

	// LastProgSeqNum
	response[4] = response[4]&0b10000111 | ((r.LastProgSeqNum & 0xf) << 3)

	// Bolus remaining pulses
	response[4] = response[4]&0b11111000 | uint8((r.BolusRemaining>>8)&0b111)
	response[5] = uint8(r.BolusRemaining & 0xff)

	// Set active alert slot bits
	response[6] = response[6]&0b10000000 | (r.Alerts >> 1)
	response[7] = response[7]&0b01111111 | (r.Alerts << 7)

	// Time Active Minutes
	response[7] = response[7]&0b10000000 | uint8((r.MinutesActive>>6)&0b01111111)
	response[8] = response[8]&0b00000011 | uint8((r.MinutesActive<<2)&0b11111100)

	if r.Reservoir < (50 / 0.05) {
		response[8] = response[8]&0b11111100 | uint8(r.Reservoir>>8)
		response[9] = uint8(r.Reservoir & 0xff)
	} else {
		response[8] = response[8] | 0b11
		response[9] = 0xff
	}

	return response, nil
}
