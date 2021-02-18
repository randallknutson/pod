package command

import (
	"fmt"

	"github.com/avereha/pod/pkg/response"

	log "github.com/sirupsen/logrus"
)

type Type byte

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
)

type Command interface {
	GetResponse() (response.Response, error)
	SetHeaderData(uint8, []byte) error
	GetHeaderData() (cmdSeq uint8, requestID []byte, err error)
}

type CommandReader struct {
	Data []byte // keep it simple for now
}

func Unmarshal(data []byte) (Command, error) {
	var err error
	if len(data) < 10 {
		return nil, fmt.Errorf("command is too short: %x", data)
	}
	if string(data[:5]) != "S0.0=" {
		return nil, fmt.Errorf("command should start with S0.0= %x", data)
	}
	n := len(data)
	if string(data[n-5:]) != ",G0.0" {
		return nil, fmt.Errorf("command should end with ,G0.0 %x", data)
	}
	l := int(data[5])<<8 | int(data[6])
	if l != n-7-5 {
		return nil, fmt.Errorf("invalid data length: %d :: %d :: %x", l, n-7-5, data)

	}
	data = data[5+2 : n-5] // remove unused strings&length
	n = len(data)
	if n < 6 {
		return nil, fmt.Errorf("command too short: %x", data)
	}

	log.Infof("Command data: %x", data)
	id := data[:4]
	var lsf uint16 = uint16(data[4])<<8 | uint16(data[5])
	length := int(lsf & 1023)
	seq := uint8((lsf >> 10) & 0x0F)
	if length+6+2 != n {
		return nil, fmt.Errorf("invalid command length %d :: %d. %x", n, length+6+2, data)
	}
	crc := data[n-2:]
	t := Type(data[6])
	log.Debugf("CRC: %x. Type: %x", crc, t)
	// TODO verify CRC
	data = data[7 : n-2]
	var ret Command
	switch t {
	case GET_VERSION:
		ret, err = UnmarshalGetVersion(data)
	case SET_UNIQUE_ID:
		ret, err = UnmarshalSetUniqueID(data)
	case PROGRAM_ALERTS, PROGRAM_BASAL, PROGRAM_INSULIN, GET_STATUS:
		ret, err = UnmarshalProgramAlerts(data)
	default:
		ret, err = UnmarshalNack(data)
	}
	if err != nil {
		return nil, err
	}
	if err := ret.SetHeaderData(seq, id); err != nil {
		return nil, err
	}
	return ret, nil
}
