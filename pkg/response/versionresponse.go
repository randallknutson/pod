package response

import (
	"encoding/hex"

	"github.com/avereha/pod/pkg/bluetooth"
)

type VersionResponse struct {
	Seq uint16
}

func Marshal() (*bluetooth.Message, error) {
	msg := getResponseMessage()
	response, _ := hex.DecodeString("0115040A00010300040208146CC1000954D400FFFFFFFF")

	withCrc := addHeaderAndCRC(response)
	msg.Payload = withCrc

	return msg, nil
}
