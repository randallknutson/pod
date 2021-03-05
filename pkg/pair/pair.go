package pair

import (
	"bytes"
	"errors"
	"fmt"

	"golang.org/x/crypto/curve25519"

	"github.com/avereha/pod/pkg/message"

	"github.com/davecgh/go-spew/spew"
	"github.com/jacobsa/crypto/cmac"
	log "github.com/sirupsen/logrus"
)

const (
	sp1 = "SP1="
	sp2 = ",SP2="

	sps1   = "SPS1="
	sps2   = "SPS2="
	sp0gp0 = "SP0,GP0"
	p0     = "P0="
)

type Pair struct {
	podPublic  []byte
	podPrivate []byte
	podNonce   []byte
	podConf    []byte

	pdmPublic []byte
	pdmNonce  []byte
	pdmConf   []byte

	curve25519LTK []byte
	pdmID         []byte
	podID         []byte

	ltk     []byte
	confKey []byte // key used to sign the "Conf" values
}

func parseStringByte(expectedNames []string, data []byte) (map[string][]byte, error) {
	ret := make(map[string][]byte)
	for _, name := range expectedNames {
		n := len(name)
		if string(data[:n]) != name {
			return nil, fmt.Errorf("Name not found %s in %x", name, data)
		}
		data = data[n:]
		length := int(data[0])<<8 | int(data[1])
		ret[name] = data[2 : 2+length]
		log.Tracef("Read field: %s :: %x :: %d", name, ret[name], len(ret[name]))

		data = data[2+length:]
	}
	return ret, nil
}

func buildStringByte(names []string, values map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	for _, name := range names {
		buf.WriteString(name)
		n := len(values[name])
		buf.WriteByte(byte(n >> 8 & 0xff))
		buf.WriteByte(byte(n & 0xff))
		buf.Write(values[name])
	}
	return buf.Bytes(), nil
}

func (c *Pair) ParseSP1SP2(msg *message.Message) error {
	log.Infof("Received SP1 SP2 payload %x", msg.Payload)

	sp, err := parseStringByte([]string{sp1, sp2}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		return err
	}

	log.Infof("Received SP1 SP2: %x :: %x", sp[sp1], sp[sp2])
	c.podID = msg.Destination
	c.pdmID = msg.Source
	return nil
}

func (c *Pair) ParseSPS1(msg *message.Message) error {
	sp, err := parseStringByte([]string{sps1}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		return err
	}
	log.Infof("Received SPS1  %x", sp[sps1])
	pdmPublic := sp[sps1][:32]
	pdmNonce := sp[sps1][32:]

	c.pdmPublic = make([]byte, 32)
	c.pdmNonce = make([]byte, 16)
	copy(c.pdmNonce, pdmNonce)
	copy(c.pdmPublic, pdmPublic)
	err = c.computeMyData()
	if err != nil {
		return err
	}
	c.curve25519LTK, err = curve25519.X25519(c.podPrivate, c.pdmPublic)
	if err != nil {
		return err
	}
	return nil
}

func (c *Pair) GenerateSPS1() (*message.Message, error) {
	var err error
	var buf bytes.Buffer

	buf.Write(c.podPublic)
	buf.Write(c.podNonce)

	sp := make(map[string][]byte)
	sp[sps1] = buf.Bytes()

	msg := message.NewMessage(message.MessageTypePairing, c.podID, c.pdmID)
	msg.Payload, err = buildStringByte([]string{sps1}, sp)
	if err != nil {
		return nil, err
	}
	err = c.computePairData()
	if err != nil {
		return nil, err
	}
	log.Debugf("Pod public %x :: %d", c.podPublic, len(c.podPublic))
	log.Debugf("Pod nonce %x :: %d", c.podNonce, len(c.podNonce))
	log.Debugf("Generated SPS1: %x", msg.Payload)
	return msg, nil
}

func (c *Pair) ParseSPS2(msg *message.Message) error {
	sp, err := parseStringByte([]string{sps2}, msg.Payload)
	if err != nil {
		log.Debugf("Message :%s", spew.Sdump(msg))
		return err
	}

	if !bytes.Equal(c.pdmConf, sp[sps2]) {
		return fmt.Errorf("Invalid conf value. Expected: %x. Got %x", c.pdmConf, sp[sps2])
	}
	log.Debugf("Validated PDM SPS2: %x", sp[sps2])
	return nil
}

func (c *Pair) GenerateSPS2() (*message.Message, error) {
	var err error
	sp := make(map[string][]byte)
	sp[sps2] = c.podConf

	msg := message.NewMessage(message.MessageTypePairing, c.podID, c.pdmID)
	msg.Payload, err = buildStringByte([]string{sps2}, sp)
	if err != nil {
		return nil, err
	}
	log.Debugf("Generated SPS2: %x", msg.Payload)
	return msg, nil
}

func (c *Pair) ParseSP0GP0(msg *message.Message) error {
	if string(msg.Payload) != sp0gp0 {
		log.Debugf("Message :%s", spew.Sdump(msg))
		return fmt.Errorf("Expected SP0GP0, got %x", msg.Payload)
	}
	log.Debugf("Parsed SP0GP0")
	return nil
}

func (c *Pair) GenerateP0() (*message.Message, error) {
	var err error
	msg := message.NewMessage(message.MessageTypePairing, c.podID, c.pdmID)
	sp := make(map[string][]byte)
	sp[p0] = []byte{0xa5} // magic constant ???
	msg.Payload, err = buildStringByte([]string{p0}, sp)
	log.Debugf("Generated P0")

	return msg, err
}

func (c *Pair) LTK() ([]byte, error) {
	if c.curve25519LTK != nil {
		return c.ltk, nil
	}
	return nil, errors.New("Missing  enough data to compute LTK")
}

func (c *Pair) computeMyData() error {
	var err error
	c.podPrivate = make([]byte, 32)
	c.podPublic = make([]byte, 32)
	c.podNonce = make([]byte, 16) // 0 for now TODO
	/*
		if _, err := rand.Read(podPrivateKey); err != nil {
			return nil, nil, nil, err
		}
	*/
	c.podPrivate[0] &= 248
	c.podPrivate[31] &= 127
	c.podPrivate[31] |= 64
	c.podPublic, err = curve25519.X25519(c.podPrivate, curve25519.Basepoint)
	return err

}
func (c *Pair) computePairData() error {
	var err error
	// fill in: lrtk, podConf, pdmConf, intermediarKey
	c.curve25519LTK, err = curve25519.X25519(c.podPrivate, c.pdmPublic)
	if err != nil {
		return err
	}
	log.Debugf("Donna LTK: %x", c.curve25519LTK)
	//first_key = data.pod_public[-4:] + data.pdm_public[-4:] + data.pod_nonce[-4:] + data.pdm_nonce[-4:]
	firstKey := append(c.podPublic[28:], c.pdmPublic[28:]...)
	firstKey = append(firstKey, c.podNonce[12:]...)
	firstKey = append(firstKey, c.pdmNonce[12:]...)
	log.Debugf("First key %x :: %d", firstKey, len(firstKey))

	first, err := cmac.New(firstKey)
	if err != nil {
		return err
	}
	log.Debugf("CMAC: %d", first.Size())
	first.Write(c.curve25519LTK)
	intermediarKey := first.Sum([]byte{})

	log.Debugf("Intermediar key %x :: %d", intermediarKey, len(intermediarKey))

	// bb_data = bytes.fromhex("01") + bytes("TWIt", "ascii") + data.pod_nonce + data.pdm_nonce + bytes.fromhex("0001")
	var bbData bytes.Buffer
	bbData.WriteByte(0x01)
	bbData.WriteString("TWIt")
	bbData.Write(c.podNonce)
	bbData.Write(c.pdmNonce)
	bbData.WriteByte(0x00)
	bbData.WriteByte(0x01)
	bbHash, err := cmac.New(intermediarKey)
	if err != nil {
		return err
	}
	bbHash.Write(bbData.Bytes())
	c.confKey = bbHash.Sum([]byte{})

	// ab_data = bytes.fromhex("02") + bytes("TWIt", "ascii") + data.pod_nonce + data.pdm_nonce + bytes.fromhex("0001")
	var abData bytes.Buffer
	abData.WriteByte(0x02) // this is the only difference
	abData.WriteString("TWIt")
	abData.Write(c.podNonce)
	abData.Write(c.pdmNonce)
	abData.WriteByte(0x00)
	abData.WriteByte(0x01)
	abHash, err := cmac.New(intermediarKey)
	if err != nil {
		return err
	}
	abHash.Write(abData.Bytes())
	c.ltk = abHash.Sum([]byte{})

	//  pdm_conf_data = bytes("KC_2_U", "ascii") + data.pdm_nonce + data.pod_nonce
	var pdmConfData bytes.Buffer
	pdmConfData.WriteString("KC_2_U")
	pdmConfData.Write(c.pdmNonce)
	pdmConfData.Write(c.podNonce)
	hash, err := cmac.New(c.confKey)
	if err != nil {
		return err
	}
	hash.Write(pdmConfData.Bytes())
	c.pdmConf = hash.Sum([]byte{})

	//  pdm_conf_data = bytes("KC_2_U", "ascii") + data.pdm_nonce + data.pod_nonce
	var podConfData bytes.Buffer
	podConfData.WriteString("KC_2_V")
	podConfData.Write(c.podNonce) // ???
	podConfData.Write(c.pdmNonce)
	hash, err = cmac.New(c.confKey)
	if err != nil {
		return err
	}
	hash.Write(podConfData.Bytes())
	c.podConf = hash.Sum([]byte{})

	return nil
}
