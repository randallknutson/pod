package message

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

type MessageType byte

const (
	MessageTypeClear MessageType = iota
	MessageTypeEncrypted
	MessageTypeSessionEstablishment
	MessageTypePairing
	MagicPattern = "TW"
)

// Message is one CTwiPacket
type Message struct {
	Eqos             uint16
	Ack              bool
	Priority         bool
	LastMessage      bool
	Gateway          bool
	Type             MessageType
	EncryptedPayload bool
	Source           []byte
	Destination      []byte
	Payload          []byte
	Raw              []byte
	Sas              bool
	Tfs              bool
	SequenceNumber   uint8
	AckNumber        uint8
	Version          uint16
}

type flag byte

func (f *flag) set(index byte, val bool) {
	var mask flag = 1 << (7 - index)
	if !val {
		return
	}
	*f |= mask
}

func (f flag) get(index byte) byte {
	var mask flag = 1 << (7 - index)
	if f&mask == 0 {
		return 0
	}
	return 1
}

func NewMessage(t MessageType, source, destination []byte) *Message {
	msg := &Message{
		Source:      make([]byte, 4),
		Destination: make([]byte, 4),
		Type:        t,
	}

	copy(msg.Source, source)
	copy(msg.Destination, destination)
	return msg
}

func (m *Message) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	if m.Type == MessageTypeEncrypted && m.EncryptedPayload { // Already encrypted
		return m.Raw, nil
	}

	f := new(flag)

	buf.WriteString(MagicPattern)

	f.set(0, (m.Version&4) != 0)
	f.set(1, (m.Version&2) != 0)
	f.set(2, (m.Version&1) != 0)
	f.set(3, m.Sas)
	f.set(4, m.Tfs)
	f.set(5, (m.Eqos&4) != 0)
	f.set(6, (m.Eqos&2) != 0)
	f.set(7, (m.Eqos&1) != 0)
	buf.WriteByte(byte(*f))

	f.set(0, m.Ack)
	f.set(1, m.Priority)
	f.set(2, m.LastMessage)
	f.set(3, m.Gateway)
	f.set(4, ((m.Type&8)>>3) == 1)
	f.set(5, ((m.Type&4)>>2) == 1)
	f.set(6, ((m.Type&2)>>1) == 1)
	f.set(7, (m.Type&1) == 1)
	buf.WriteByte(byte(*f))

	buf.WriteByte(byte(m.SequenceNumber))
	buf.WriteByte(byte(m.AckNumber))
	var l uint16
	if m.Payload != nil {
		l = uint16(len(m.Payload))
	}
	buf.WriteByte(byte(l >> 3))
	buf.WriteByte(byte(l << 5))

	buf.Write(m.Source)
	buf.Write(m.Destination)
	if m.Payload != nil {
		buf.Write(m.Payload)
	}

	ret := make([]byte, buf.Len())
	copy(ret, buf.Bytes())
	m.Raw = ret

	return ret, nil
}

func Unmarshal(data []byte) (*Message, error) {
	ret := &Message{}
	ret.Raw = data

	if len(data) < 16 {
		return nil, fmt.Errorf("data %x is too short to parse as a Message", data)
	}
	if string(data[:2]) != MagicPattern {
		return nil, fmt.Errorf("magic pattern not found in %x", data)
	}

	f := flag(data[2])
	ret.Sas = f.get(3) != 0
	ret.Tfs = f.get(4) != 0
	ret.Version = uint16(f.get(2) | f.get(1)<<1 | f.get(0)<<2)
	ret.Eqos = uint16(f.get(7) | f.get(6)<<1 | f.get(5)<<2)

	f = flag(data[3])
	ret.Ack = f.get(0) == 1
	ret.Priority = f.get(1) == 1
	ret.LastMessage = f.get(2) == 1
	ret.Gateway = f.get(3) == 1
	ret.Type = MessageType(f.get(7) | f.get(6)<<1 | f.get(5)<<2 | f.get(4)<<3)

	if ret.Type > MessageTypePairing {
		return nil, fmt.Errorf("invalid message type found in %x", data)
	}
	if ret.Version != 0 {
		return nil, fmt.Errorf("invalid version received in %x", data)
	}
	ret.SequenceNumber = data[4]
	ret.AckNumber = data[5]
	var n = data[6]<<3 | data[7]>>5
	if int(n) > len(data)-16 {
		spew.Dump(ret)
		return nil, fmt.Errorf("received length is too big in %x. Length:%d . remaining: %d", data, n, len(data)-16)
	}
	ret.Source = data[8:12]
	ret.Destination = data[12:16]
	if ret.Type == MessageTypeEncrypted {
		ret.Payload = data[16 : 16+n+8]
		ret.EncryptedPayload = true
	} else {
		ret.Payload = data[16 : 16+n]
	}
	return ret, nil
}
