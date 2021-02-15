package eap

import (
	"encoding/hex"
	"fmt"

	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/wmnsk/milenage"
)

type Code byte
type SubType byte
type AttributeType byte

const (
	CodeRequest Code = iota + 1
	CodeReponse
	CodeSuccess
	CodeFailure

	SubTypeAkaChallenge = 1

	AT_RAND      AttributeType = 1
	AT_AUTN                    = 2
	AT_RES                     = 3
	AT_CUSTOM_IV               = 126
)

type Attribute struct {
	Type AttributeType
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
}

func Unmarshal(data []byte) (*EapAka, error) {
	var ret = &EapAka{
		Attributes: make(map[AttributeType]*Attribute),
	}
	if len(data) < 4 { // TODO
		return nil, fmt.Errorf("data is too short for an EAP packet %x", data)
	}
	ret.Len = int(data[2])<<8 + int(data[3])
	ret.Code = Code(data[0])
	if ret.Code > CodeFailure {
		return nil, fmt.Errorf("invalid eap code: %d %x", ret.Code, data)
	}
	if len(data) <= 4 { // success/failure
		return ret, nil
	}

	ret.Identifier = data[1]
	if data[4] != 23 {
		return nil, fmt.Errorf("invalid eap packet type. Expected 23. %d %x", data[4], data)
	}
	ret.SubType = SubType(data[5])
	tail := data[8:]
	for len(tail) > 0 {
		len := tail[1] * 4
		a := &Attribute{
			Type: AttributeType(tail[0]),
			Data: tail[2:len],
		}
		ret.Attributes[a.Type] = a
		if a.Type == 1 || a.Type == 2 {
			// TODO check reserverved bytes are 0
			a.Data = a.Data[2:]
		}
		tail = tail[len:]
	}

	return ret, nil
}

func (e *EapAka) Marshal() ([]byte, error) {

	return nil, nil
}

func NewEapAkaChallenge(k []byte) *EapAkaChallenge {
	op, _ := hex.DecodeString("cdc202d5123e20f62b6d676ac72cb318")
	//	amf, _ := hex.DecodeString("b9b9")

	return &EapAkaChallenge{
		k:   k,
		op:  op,
		amf: 47545, // b9b9
	}
}

func (e *EapAkaChallenge) ParseChallenge(msg *bluetooth.Message) error {
	e.pdmID = msg.Source
	e.podID = msg.Destination
	eapChallenge, err := Unmarshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("error parsing eap message: %s", err)
	}
	log.Debugf("challenge: %s", spew.Sdump(eapChallenge))
	e.rand = eapChallenge.Attributes[AT_RAND].Data
	e.autn = eapChallenge.Attributes[AT_AUTN].Data
	e.pdmIV = eapChallenge.Attributes[AT_CUSTOM_IV].Data

	return nil
}

func (e *EapAkaChallenge) SqnBytes() []byte {
	return nil
}

func (e *EapAkaChallenge) CKNonce() ([]byte, []byte) {
	nonce := append(e.podIV, e.pdmIV...)
	nonce = append(nonce, e.SqnBytes()...)

	return e.ck, nonce
}

func (e *EapAkaChallenge) GenerateChallengeResponse() (*bluetooth.Message, error) {
	var err error
	ret := bluetooth.NewMessage(bluetooth.MessageTypeSessionEstablishment, e.podID, e.pdmID)
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

	eap := &EapAka{
		Code: CodeReponse,
	}

	ret.Payload, err = eap.Marshal()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (e *EapAkaChallenge) ParseSuccess(msg *bluetooth.Message) error {
	eap, err := Unmarshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("error parsing eap message: %s", err)
	}
	if eap.Code != CodeSuccess {
		return fmt.Errorf("eap code is not success: %s", spew.Sdump(eap))
	}
	return nil
}
