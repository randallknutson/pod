package eap

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/avereha/pod/pkg/message"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/wmnsk/milenage"
)

type Code byte
type SubType byte
type AttributeType byte

const (
	CodeRequest Code = iota + 1
	CodeResponse
	CodeSuccess
	CodeFailure

	SubTypeAkaChallenge = 1

	AT_RAND      AttributeType = 1
	AT_AUTN      AttributeType = 2
	AT_RES       AttributeType = 3
	AT_CUSTOM_IV AttributeType = 126
)

type Attribute struct {
	Data []byte
}

type EapAka struct {
	Len        int
	Code       Code
	Identifier byte
	SubType    SubType
	Attributes map[AttributeType]*Attribute
}

type EapAkaChallenge struct {
	podID []byte
	pdmID []byte

	k    []byte
	ck   []byte // session key
	rand []byte
	autn []byte
	res  []byte
	amf  uint16
	op   []byte

	podIV []byte
	pdmIV []byte
	Sqn   uint64

	identifier byte
}

func Unmarshal(data []byte) (*EapAka, error) {
	var ret = &EapAka{}
	if len(data) < 4 { // TODO
		return nil, fmt.Errorf("data is too short for an EAP packet %x", data)
	}
	ret.Code = Code(data[0])
	if ret.Code > CodeFailure {
		return nil, fmt.Errorf("invalid eap code: %d %x", ret.Code, data)
	}

	ret.Identifier = data[1]
	ret.Len = int(data[2])<<8 + int(data[3])

	if ret.Len <= 4 { // eap success/failure
		return ret, nil
	}
	if ret.Len < len(data) {
		return nil, fmt.Errorf("expected len: %d. Actual: %d", ret.Len, len(data))
	}

	if data[4] != 0x17 {
		return nil, fmt.Errorf("invalid eap packet type. Expected 23. %d %x", data[4], data)
	}
	ret.SubType = SubType(data[5])
	ret.Attributes = make(map[AttributeType]*Attribute)
	tail := data[8:]
	for len(tail) > 0 {
		len := tail[1] * 4
		aType := AttributeType(tail[0])
		data = tail[2:len]
		switch aType {
		case AT_RAND, AT_AUTN:
			if len != 20 {
				return nil, fmt.Errorf("invalid len received for attribute: %d -- %d", aType, len)
			}
			data = data[2:] // skip two reserved bytes
		case AT_RES:
			if len != 12 {
				return nil, fmt.Errorf("invalid len received for attribute: %d -- %d", aType, len)
			}
			resBits := int(data[0])<<8 | int(data[1])
			if resBits != 64 {
				return nil, fmt.Errorf("invalid res bits received for attribute: %d -- %d", resBits, aType)

			}
			data = data[2:10] // 8 bytes
		case AT_CUSTOM_IV:
			if len != 8 {
				return nil, fmt.Errorf("invalid len received for attribute: %d -- %d", len, aType)
			}
			data = data[2:] // skip two reserved bytes
		default:
			return nil, fmt.Errorf("received unknown EAP attribute type: %d", aType)
		}
		a := &Attribute{
			Data: data,
		}
		ret.Attributes[aType] = a
		tail = tail[len:]
	}

	return ret, nil
}

func (e *EapAka) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteByte(byte(e.Code))
	buf.WriteByte(e.Identifier)
	//len, will fill at the end
	buf.Write([]byte{0, 0})
	if len(e.Attributes) == 0 { // short packet: success/failure
		len := uint16(buf.Len()) //?
		e.Len = int(len)
		log.Tracef("short packet buf len: %d", buf.Len())
		ret := buf.Bytes()
		ret[2] = byte(len<<8) & 0xFF
		ret[3] = byte(len)
		return ret, nil
	}
	buf.WriteByte(0x17) // Aka
	buf.WriteByte(byte(e.SubType))
	buf.Write([]byte{0, 0}) // Eap-Aka reserved

	for k, v := range e.Attributes {
		var len byte
		buf.WriteByte(byte(k))
		//len := byte(len(v.Data) / 4)
		dataLen := 0
		switch k {
		case AT_RAND, AT_AUTN:
			len = 5 // 5 * 4 = 20 bytes
			buf.WriteByte(len)
			buf.Write([]byte{0, 0}) // two reserved bytes that are set to 0
			dataLen = 16
		case AT_RES:
			len = 3 // 3 * 4 == 12 bytes
			buf.WriteByte(len)
			buf.WriteByte(0)
			buf.WriteByte(64) // RES len in bits
			dataLen = 8
		case AT_CUSTOM_IV:
			len = 2
			buf.WriteByte(len)
			buf.Write([]byte{0, 0}) // two reserved bytes that are set to 0
			dataLen = 4
		default:
			return nil, fmt.Errorf("don't know how to marshal attribute type %d", k)
		}
		buf.Write(v.Data[:dataLen])
	}
	len := uint16(buf.Len()) //?
	e.Len = int(len)
	ret := buf.Bytes()
	ret[2] = byte(len<<8) & 0xFF
	ret[3] = byte(len)
	return ret, nil
}

func NewEapAkaChallenge(k []byte, sqn uint64) *EapAkaChallenge {
	op, _ := hex.DecodeString("cdc202d5123e20f62b6d676ac72cb318")
	// amf, _ := hex.DecodeString("b9b9")
	log.Debugf("Starting EAP-AKA session with SQN(after incrementing SQN): %d", sqn+1)
	return &EapAkaChallenge{
		k:     k,
		op:    op,
		Sqn:   sqn + 1,
		amf:   47545,                      // b9b9
		podIV: []byte{0xa, 0xa, 0xa, 0xa}, // constant for now, easier to debug. TODO
	}
}

func (e *EapAkaChallenge) ParseChallenge(msg *message.Message) error {
	e.pdmID = msg.Source
	e.podID = msg.Destination
	eapChallenge, err := Unmarshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("error parsing eap message: %s", err)
	}

	log.Debugf("received EAP-AKA challenge: %s", spew.Sdump(eapChallenge))
	e.rand = eapChallenge.Attributes[AT_RAND].Data
	e.autn = eapChallenge.Attributes[AT_AUTN].Data
	e.pdmIV = eapChallenge.Attributes[AT_CUSTOM_IV].Data
	e.identifier = eapChallenge.Identifier

	return nil
}

func (e *EapAkaChallenge) SqnBytes() []byte {
	return nil
}

func (e *EapAkaChallenge) CKNoncePrefix() ([]byte, []byte) {
	nonce := append(e.pdmIV, e.podIV...)

	return e.ck, nonce
}

func (e *EapAkaChallenge) GenerateChallengeResponse() (*message.Message, error) {
	var err error
	ret := message.NewMessage(message.MessageTypeSessionEstablishment, e.podID, e.pdmID)
	mil := milenage.New(
		e.k,
		e.op,
		e.rand,
		e.Sqn,
		e.amf,
	)

	// TODO check AUTN

	// TODO check if IK/AK is used for anything
	e.res, e.ck, _, _, err = mil.F2345()
	if err != nil {
		return nil, err
	}

	eap := &EapAka{Code: CodeResponse,
		Attributes: make(map[AttributeType]*Attribute),
		SubType:    SubTypeAkaChallenge,
		Identifier: e.identifier,
	}
	eap.Attributes[AT_RES] = &Attribute{
		Data: e.res,
	}
	eap.Attributes[AT_CUSTOM_IV] = &Attribute{
		Data: e.podIV,
	}
	ret.Payload, err = eap.Marshal()
	if err != nil {
		return nil, err
	}
	mil.F1()
	log.Tracef("response: %s :: %d", spew.Sdump(eap), len(eap.Attributes))
	log.Debugf("EapAka response payload: %x", ret.Payload)
	log.Debugf("EapAka AUTN %x", e.autn)
	log.Debugf("EapAka RAND %x", e.rand)
	log.Debugf("EapAka RES %x", e.res)
	log.Debugf("EapAka Milenage AK %x", mil.AK)
	log.Debugf("EapAka Milenage MACA %x", mil.MACA)
	log.Debugf("EapAka CK %x", e.ck)
	log.Debugf("EapAka K %x", e.k)
	log.Debugf("EapAka podIV %x", e.podIV)
	log.Debugf("EapAka pdmIV %x", e.pdmIV)

	return ret, nil
}

func (e *EapAkaChallenge) ParseSuccess(msg *message.Message) error {
	eap, err := Unmarshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("error parsing eap message: %s", err)
	}
	if eap.Code != CodeSuccess {
		return fmt.Errorf("eap code is not success: %s", spew.Sdump(eap))
	}
	return nil
}
