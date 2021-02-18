package encrypt

import (
	"crypto/aes"
	"fmt"

	"github.com/avereha/pod/pkg/bluetooth"

	aesccm "github.com/pschlump/AesCCM"
	log "github.com/sirupsen/logrus"
)

func buildNonce(noncePrefix []byte, seq uint64, podReceiving bool) []byte {
	log.Infof("Seq: %d", seq)
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

func DecryptMessage(ck, noncePrefix []byte, seq uint64, msg *bluetooth.Message) (*bluetooth.Message, error) {
	log.Debugf("Using CK:    %x", ck)
	nonce := buildNonce(noncePrefix, seq, true)
	log.Debugf("Using Nonce: %x :: %d", nonce, len(nonce))
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

	log.Debugf("MAC: %x", tag, len(tag))
	log.Debugf("Data: %x :: %d", encryptedData, len(encryptedData))
	log.Debugf("Header: %x :: %d", header, len(header))

	decrypted, err := ccm.Open([]byte{}, nonce, msg.Payload, header)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt: %w", err)
	}
	msg.EncryptedPayload = false
	msg.Payload = decrypted
	log.Debugf("Decrypted: %x", decrypted)

	return msg, nil
}

func EncryptMessage(ck, noncePrefix []byte, seq uint64, msg *bluetooth.Message) (*bluetooth.Message, error) {
	if msg.EncryptedPayload {
		return msg, nil
	}

	log.Debugf("Using CK:    %x", ck)
	nonce := buildNonce(noncePrefix, seq, false)
	log.Debugf("Using Nonce: %x :: %d", nonce, len(nonce))
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
	log.Debugf("Encrypted: %x", msg.Raw)

	return msg, nil
}
