package response

import "github.com/avereha/pod/pkg/bluetooth"

type Response interface {
	Marshal() (*bluetooth.Message, error)
}

func getResponseMessage() *bluetooth.Message {
	msg := bluetooth.Message{
		Type: bluetooth.MessageTypeEncrypted,
	}
	return &msg
}

func addHeaderAndCRC(data []byte) []byte {
	return data
}
