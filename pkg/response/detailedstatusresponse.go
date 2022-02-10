package response

import (
	"encoding/hex"
)

type DetailedStatusResponse struct {
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
	FaultEvent          uint8
	FaultEventTime      uint16
}

func (r *DetailedStatusResponse) Marshal() ([]byte, error) {

	// OFF 1  2  3  4  5 6  7  8 9 10 1112 1314 1516 17 18 19 20 21 2223
	// 02 16 02 0J 0K LLLL MM NNNN PP QQQQ RRRR SSSS TT UU VV WW XX YYYY
	// 02 16 02 08 02 0000 00 01b2 00 0000 03ff 01cc 00 00 00 1f ff 030d

	response, _ := hex.DecodeString("021602000000000001b200000003ff01cc0000001fff030d")

	// PodProgress
	response[3] = byte(r.PodProgress)

	// Delivery bits
	if r.BasalActive {
		response[4] = response[4] | (1 << 0)
	}
	if r.TempBasalActive {
		response[4] = response[4] | (1 << 1)
	}
	if r.BolusActive {
		response[4] = response[4] | (1 << 2)
	}
	if r.ExtendedBolusActive {
		response[4] = response[4] | (1 << 3)
	}

	// Bolus remaining pulses
	response[5] = byte(r.BolusRemaining >> 8)
	response[6] = byte(r.BolusRemaining & 0xff)

	// LastProgSeqNum
	response[7] = r.LastProgSeqNum

	// Total delivered pulses
	response[8] = byte(r.Delivered >> 8)
	response[9] = byte(r.Delivered & 0xff)

	// Fault event
	response[10] = r.FaultEvent

	// Fault Event Time
	response[11] = byte(r.FaultEventTime >> 8)
	response[12] = byte(r.FaultEventTime & 0xff)

	// Reservoir
	if r.Reservoir < (50 / 0.05) {
		response[13] = byte(r.Reservoir >> 8)
		response[14] = byte(r.Reservoir & 0xff)
	} else {
		response[13] = 0x03
		response[14] = 0xff
	}

	// Minutes since activation
	response[15] = byte(r.MinutesActive >> 8)
	response[16] = byte(r.MinutesActive & 0xff)

	// Set active alert slot bits
	response[6] = r.Alerts

	// TODO: add other fault details

	return response, nil
}
