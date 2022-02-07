package command

import (
	"fmt"

	"github.com/avereha/pod/pkg/response"

	log "github.com/sirupsen/logrus"
)

type Type byte

type Payload []byte

const (
	SET_UNIQUE_ID      Type = 0x03
	GET_VERSION        Type = 0x07
	GET_STATUS         Type = 0x0e
	SILENCE_ALERTS     Type = 0x11
	PROGRAM_BASAL      Type = 0x13 // Always preceded by 0x1a
	PROGRAM_TEMP_BASAL Type = 0x16 // Always preceded by 0x1a
	PROGRAM_BOLUS      Type = 0x17 // Always preceded by 0x1a
	PROGRAM_ALERTS     Type = 0x19
	PROGRAM_INSULIN    Type = 0x1a // Always followed by one of: 0x13, 0x16, 0x17
	DEACTIVATE         Type = 0x1c
	PROGRAM_BEEPS      Type = 0x1e
	STOP_DELIVERY      Type = 0x1f
	CNFG_DELIV_FLAG    Type = 0x08 // Loop uses configure delivery flag
	NACK               Type = 0x00 // Loop uses configure delivery flag
)

var (
	CommandName = map[Type]string{
		SET_UNIQUE_ID:      "SET_UNIQUE_ID",
		GET_VERSION:        "GET_VERSION",
		GET_STATUS:         "GET_STATUS",
		SILENCE_ALERTS:     "SILENCE_ALERTS",
		PROGRAM_BASAL:      "PROGRAM_BASAL",
		PROGRAM_TEMP_BASAL: "PROGRAM_TEMP_BASAL",
		PROGRAM_BOLUS:      "PROGRAM_BOLUS",
		PROGRAM_ALERTS:     "PROGRAM_ALERTS",
		PROGRAM_INSULIN:    "PROGRAM_INSULIN",
		DEACTIVATE:         "DEACTIVATE",
		PROGRAM_BEEPS:      "PROGRAM_BEEPS",
		STOP_DELIVERY:      "STOP_DELIVERY",
		CNFG_DELIV_FLAG:    "CNFG_DELIV_FLAG",
	}
)

type CommandResponseType int64
const (
	ShortStatus CommandResponseType = iota // 0x02
	//LongStatus
	//PodInfo
	Hardcoded // Return this and pod.go will use command.GetResponse
)

type Command interface {
	GetResponseType() (CommandResponseType)
	GetResponse() (response.Response, error)
	SetHeaderData(uint8, []byte) error
	GetHeaderData() (cmdSeq uint8, requestID []byte, err error)
	GetPayload() Payload
	GetType() Type
}

type CommandReader struct {
	Data []byte // keep it simple for now
}

// PodProgress used to select appropriate the 0x1d response
// initialize to 8 to support restart of pod simulator
// when started with -fresh flag, modified by the 0x07: GET_VERSION, etc.
var PodProgress = 8

func Unmarshal(data []byte) (Command, error) {
	var err error
	if len(data) < 10 {
		return nil, fmt.Errorf("pkg command; command is too short: %x", data)
	}
	if string(data[:5]) != "S0.0=" {
		return nil, fmt.Errorf("pkg command; command should start with S0.0= %x", data)
	}
	n := len(data)
	if string(data[n-5:]) != ",G0.0" {
		return nil, fmt.Errorf("pkg command; command should end with ,G0.0 %x", data)
	}
	l := int(data[5])<<8 | int(data[6])
	if l != n-7-5 {
		return nil, fmt.Errorf("pkg command; invalid data length: %d :: %d :: %x", l, n-7-5, data)

	}
	data = data[5+2 : n-5] // remove unused strings&length
	n = len(data)
	if n < 6 {
		return nil, fmt.Errorf("pkg command; command too short: %x", data)
	}

	log.Tracef("pkg command; command data: %x", data)
	id := data[:4]
	var lsf uint16 = uint16(data[4])<<8 | uint16(data[5])
	length := int(lsf & 1023)
	seq := uint8((lsf >> 10) & 0x0F)
	if length+6+2 != n {
		return nil, fmt.Errorf("pkg command; invalid command length %d :: %d. %x", n, length+6+2, data)
	}
	crc := data[n-2:]
	log.Tracef("pkg command; CRC = %x", crc)
	t := Type(data[6])
	log.Infof("pkg command; 0x%2.2x; %s; HEX, %x", t, CommandName[t], data)

	data = data[7 : n-2]
	var ret Command

	switch t {
	case GET_VERSION:
		ret, err = UnmarshalGetVersion(data)
		PodProgress = 2 // set with -fresh
	case SET_UNIQUE_ID:
		ret, err = UnmarshalSetUniqueID(data)
		PodProgress = 3 // set with -fresh
	case PROGRAM_ALERTS:
		if PodProgress < 4 {
			ret, err = UnmarshalProgramAlertsBeforePrime(data)
		} else if PodProgress < 7 {
			ret, err = UnmarshalProgramAlertsBeforeInsert(data)
		} else {
			ret, err = UnmarshalProgramAlerts(data)
		}
	case PROGRAM_INSULIN:
		if PodProgress < 4 {
			// this must be the prime command
			PodProgress = 4
			ret, err = UnmarshalProgramInsulinPrime(data)
		} else if PodProgress < 6 {
			// this must be the program scheduled basal command
			PodProgress = 6
			ret, err = UnmarshalProgramInsulinSchedule(data)
		} else if PodProgress < 7 {
			// this must be the insert cannula command
			PodProgress = 7
			ret, err = UnmarshalProgramInsulinInsert(data)
		} else {
			if PodProgress < 8 {
				PodProgress = 8
			}
			ret, err = UnmarshalProgramInsulin(data)
		}
	case GET_STATUS:
		// type 7 returns page0, dash specific type
		if data[1] == 0 || data[1] == 7 {
			if PodProgress <= 4 {
				// PDM uses a type 7 get status after prime
				PodProgress = 5
				ret, err = UnmarshalProgramPrimeComplete(data)
			} else {
				if PodProgress < 8 {
					PodProgress = 8
				}
				ret, err = UnmarshalGetStatus(data)
			}
		} else if data[1] == 0x2 {
			ret, err = UnmarshalType2Status(data)
		} else if data[1] == 0x46 {
			ret, err = UnmarshalType46Status(data)
		} else if data[1] == 0x50 {
			ret, err = UnmarshalType50Status(data)
		} else if data[1] == 0x51 {
			ret, err = UnmarshalType51Status(data)
		} else {
			ret, err = UnmarshalNack(data)
		}
	case SILENCE_ALERTS:
		ret, err = UnmarshalSilenceAlerts(data)
	case DEACTIVATE:
		ret, err = UnmarshalDeactivate(data)
	case PROGRAM_BEEPS:
		ret, err = UnmarshalProgramBeeps(data)
	case STOP_DELIVERY:
		ret, err = UnmarshalStopDelivery(data)
	case CNFG_DELIV_FLAG:
		ret, err = UnmarshalCnfgDelivFlag(data)
	default:
		ret, err = UnmarshalNack(data)
	}

	if err != nil {
		return nil, err
	}
	if err := ret.SetHeaderData(seq, id); err != nil {
		return nil, err
	}

	log.Debugf("pkg command; PodProgress = %d", PodProgress)
	return ret, nil
}
