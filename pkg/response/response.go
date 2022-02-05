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
	var priorSeq uint8

	payload, err := rsp.Marshal()
	if err != nil {
		return nil, err
	}

	msgType = uint16(payload[0])
	priorSeq = seq - 1
	// Special treatment for 0x1d response
	if (msgType == 0x1d) {
		//     The $1D status response has the following form:
		//        byte# 00 01 02 03 04 05 06070809
		//              1d SS 0P PP SN NN AATTTTRR
		// 0PPPSNNN dword = 0000 pppp pppp pppp psss snnn nnnn nnnn
		//        the s bits of pssssnnn must be priorSeq
		log.Debugf("pkg response: message body (before s bit update) = %x", payload)
		log.Debugf("pkg response: msgType 0x%2.2x; priorSeq %x; seq %x", msgType, priorSeq, seq)
		payload[4] = (payload[4] & 0x87)  | (0x78 & (priorSeq << 3))
	}

	log.Debugf("pkg response: message body to encrypt %x", payload)

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

	//log.Infof("pkg response; priorSeq 0x%x; msgType 0x%2.2x; simResp, %x", priorSeq, msgType, buf.Bytes())
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
