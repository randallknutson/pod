package pod

import (
	"io/ioutil"

	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

type PODState struct {
	LTK       []byte `toml:"ltk"`
	EapAkaSeq uint64 `toml:"eap_aka_seq"`

	Id []byte `toml:"id"` // 4 byte

	MsgSeq   uint8  `toml:"msg_seq"`   // TODO: is this the same as nonceSeq?
	CmdSeq   uint8  `toml:"cmd_seq"`   // TODO: are all those 3 the same number ???
	NonceSeq uint64 `toml:"nonce_seq"` // or 16?

	NoncePrefix []byte `toml:"nonce_prefix"`
	CK          []byte `toml:"ck"`

	ReservoirLevel float32
	Alerts         uint8
	FaultType      uint8

	filename    string
}

func NewState(filename string) (*PODState, error) {
	var ret PODState
	ret.filename = filename
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (p *PODState) Save() error {
	log.Debugf("Saving state to file: %s", p.filename)
	data, err := toml.Marshal(p)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p.filename, data, 0777)
}
