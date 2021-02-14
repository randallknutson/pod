package pod

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/curve25519"

	log "github.com/sirupsen/logrus"
)

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

func generateSPS1() ([]byte, []byte, []byte, error) {
	podPrivateKey := make([]byte, 32)
	podPublicKey := make([]byte, 32)
	podNonce := make([]byte, 16) // 0 for now
	/*
		if _, err := rand.Read(podPrivateKey); err != nil {
			return nil, nil, nil, err
		}
	*/
	podPrivateKey[0] &= 248
	podPrivateKey[31] &= 127
	podPrivateKey[31] |= 64

	podPublicKey, err := curve25519.X25519(podPrivateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, nil, err
	}

	return podPrivateKey, podPublicKey, podNonce, nil
}

func getLTK(pdmPublicKey, podSecretKey []byte) ([]byte, error) {
	return curve25519.X25519(podSecretKey, pdmPublicKey)
}
