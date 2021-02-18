package encrypt

import (
	"crypto/aes"
	"fmt"

	"github.com/avereha/pod/pkg/message"
	aesccm "github.com/pschlump/AesCCM"
	log "github.com/sirupsen/logrus"
)

func buildNonce(noncePrefix []byte, seq uint64, podReceiving bool) []byte {
	seq &= 549755813887
	seqBytes := []byte{
		(byte)(seq >> 32),
		(byte)(seq >> 24),
		(byte)(seq >> 16),
		(byte)(seq >> 8),
		(byte)(seq),
	}
	if podReceiving {
		seqBytes[0] &= 127
	} else {
		seqBytes[0] |= 128
	}
	return append(noncePrefix, seqBytes...)
}

func DecryptMessage(ck, noncePrefix []byte, seq uint64, msg *message.Message) (*message.Message, error) {
	log.Tracef("using CK:    %x", ck)
	nonce := buildNonce(noncePrefix, seq, true)
	log.Tracef("decrypt: using nonce: %x :: %d", nonce, len(nonce))
	aes, err := aes.NewCipher(ck)
	if err != nil {
		return nil, fmt.Errorf("could not create aes: %w", err)
	}
	ccm, err := aesccm.NewCCM(aes, 8, len(nonce))
	if err != nil {
		return nil, fmt.Errorf("could not create aes-ccm: %w", err)
	}

	header := msg.Raw[:16]
	n := len(msg.Payload)
	tag := msg.Payload[n-8:]
	encryptedData := msg.Payload[:n-8]

	log.Tracef("tag: %x :: %d", tag, len(tag))
	log.Tracef("data: %x :: %d", encryptedData, len(encryptedData))
	log.Tracef("header: %x :: %d", header, len(header))

	decrypted, err := ccm.Open([]byte{}, nonce, msg.Payload, header)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt: %w", err)
	}
	msg.EncryptedPayload = false
	msg.Payload = decrypted
	log.Tracef("decrypted: %x", decrypted)

	return msg, nil
}

func EncryptMessage(ck, noncePrefix []byte, seq uint64, msg *message.Message) (*message.Message, error) {
	if msg.EncryptedPayload {
		return msg, nil
	}

	log.Tracef("using CK:    %x", ck)
	nonce := buildNonce(noncePrefix, seq, false)
	log.Tracef("encrypt: using nonce: %x :: %d", nonce, len(nonce))
	aes, err := aes.NewCipher(ck)
	if err != nil {
		return nil, fmt.Errorf("could not create aes: %w", err)
	}
	ccm, err := aesccm.NewCCM(aes, 8, len(nonce))
	if err != nil {
		return nil, fmt.Errorf("could not create aes-ccm: %w", err)
	}
	if msg.Raw == nil {
		if _, err := msg.Marshal(); err != nil {
			return nil, err
		}
	}

	header := msg.Raw[:16]
	toEncrypt := msg.Raw[16:]

	encrypted := ccm.Seal(nil, nonce, toEncrypt, header)
	msg.Raw = append(header, encrypted...)
	msg.EncryptedPayload = true
	log.Tracef("encrypted: %x", msg.Raw)

	return msg, nil
}
