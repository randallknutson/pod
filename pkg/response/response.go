package response

import (
	"bytes"
	"fmt"

	"github.com/avereha/pod/pkg/crc"
	"github.com/avereha/pod/pkg/message"

	log "github.com/sirupsen/logrus"
)

type Response interface {
	Marshal() ([]byte, error)
}

type ResponseMetadata struct {
	CmdSeq uint8
	MsgSeq uint8
	AckSeq uint8

	RequestID []byte
	Src       []byte
	Dst       []byte
}

func payloadWithHeaderAndCRC(rsp Response, seq uint8, responseID []byte) ([]byte, error) {
	var buf bytes.Buffer
	var header uint16
	var msgType uint16

	payload, err := rsp.Marshal()
	if err != nil {
		return nil, err
	}

	msgType = uint16(payload[0])

	// write header
	buf.Write(responseID) // 4 bytes
	if buf.Len() != 4 {
		return nil, fmt.Errorf("responseID should be 4 bytes, got:  %x", responseID)
	}
	seq &= 0x0F
	header = uint16(seq) << 10
	payloadLen := uint16(len(payload))
	header |= (payloadLen & 0x03FF) // last 10 bits

	buf.WriteByte(byte(header >> 8))
	buf.WriteByte(byte(header))
	buf.Write(payload)
	buf.Write(crc.CRC16(buf.Bytes()))

	log.Infof("pkg response 0x%x; HEX, %x", msgType, buf.Bytes())

	return buf.Bytes(), nil
}

func Marshal(rsp Response, metadata *ResponseMetadata) (*message.Message, error) {
	var buf bytes.Buffer

	payload, err := payloadWithHeaderAndCRC(rsp, metadata.CmdSeq, metadata.RequestID)
	if err != nil {
		return nil, err
	}

	buf.WriteString("0.0=")  // TODO figure out error cases: 0.2, 0.5, 0.0=00, 1.0=
	totalLen := len(payload) // length including header and CRC
	buf.WriteByte(byte(totalLen >> 8))
	buf.WriteByte(byte(totalLen))
	buf.Write(payload)

	msg := message.NewMessage(message.MessageTypeEncrypted, metadata.Src, metadata.Dst)
	msg.Payload = buf.Bytes()
	msg.SequenceNumber = metadata.MsgSeq
	msg.Ack = true
	msg.AckNumber = metadata.AckSeq
	msg.Eqos = 1 // It seems to be some sort of flag that this is a reply
	// TODO fill other msg fields
	return msg, nil
}
