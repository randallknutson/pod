package main

import (
	"crypto/aes"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/hex"

	"github.com/davecgh/go-spew/spew"
	aesccm "github.com/pschlump/AesCCM"
	log "github.com/sirupsen/logrus"
)

var podPrivate []byte
var podPublic []byte
var podNonce []byte
var pdmPublic []byte
var pdmNonce []byte
var receivedSPS2 []byte
var firmwareId []byte

func main() {
	podPrivate, _ = hex.DecodeString("facaa985b49cfa3f49add707b0fc0e9e964e09e8cf97fb12186cd9d5827f6099")
	podPublic, _ = hex.DecodeString("b796fade1b1024f71d3441df4cce0dd3378e56819c5b0a357edef460020288c3c598d09c5b406ea121af02d8b912b17f24ef01775419eddfb580ad7ae96d46b9")
	podNonce, _ = hex.DecodeString("81ae98dfc420acdb2fc683ffa51d6627")
	pdmPublic, _ = hex.DecodeString("ab08e5f51a2b12b31ea712a6c72827870d0af6a935a81741840c6da0d8872d70a04731c111c068ddf472cfdeab8bc16059428e7ee3499c12dd1a7e64ad5ecea0")
	pdmNonce, _ = hex.DecodeString("a0487adf1a5012934973cc4f12abc5cc")
	receivedSPS2, _ = hex.DecodeString("0282e4ce38911533632b9273ed554a1cdd730971a982b73207ee6a4be5336f35a734328d0a6dceb50c5f04db9345a29798ae02dfc4da7861e05c75911ddf9af717474ad2e56d5f86927e3512bb191d2a4a4e89ae6f408ed4e4b0c56ef2b6a2576c047a58dc4188841411acb139fb2139abeaf4fd8c8d17119fbf5706c344af123fe28fba271de512ba1c0288271ab1141047a610b9b0e6044f105ce25aa8c3944908b27b1df41c3407ac8a95b148f72ea00e6f0698c2a2e30567e160fd747604ad12632147b6b305c81a60c49ac1e9354f4d9c99a6c020b230d9f9029c3b54")
	firmwareId, _ = hex.DecodeString("9b0ab96a76f4") // Hard coded
	// log.Infof("receivedSPS2: %x :: %d", receivedSPS2, len(receivedSPS2))
	// 151 bytes ASN.1 DER encoded :: 64 bytes certificate :: 8 bytes CCM Checksum

	// var ans1der = receivedSPS2[:215]
	// log.Infof("ans1der: %x :: %d", ans1der, len(ans1der))
	// type resultType struct {
	// x int
	// }
	// var tmp resultType
	// var cert, err = asn1.Unmarshal(ans1der, tmp)
	// cert, err := x509.ParseCertificate(receivedSPS2[0:215])
	// if err != nil {
	// log.Infof("Error :%s", spew.Sdump(err))
	// }
	// log.Infof("result: %x :: %d", result, len(result))
	// log.Infof("cert: %x", spew.Sdump(cert))

	privateKey, err := ecdh.P256().NewPrivateKey(podPrivate)
	if err != nil {
		log.Infof("Error :%s", spew.Sdump(err))
	}

	publicKey, err := ecdh.P256().NewPublicKey(append([]byte{0x04}, pdmPublic...))
	if err != nil {
		log.Infof("Error :%s", spew.Sdump(err))
	}

	sharedSecret, err := privateKey.ECDH(publicKey)

	// log.Infof("Shared Secret: %x :: %d", sharedSecret, len(sharedSecret))

	controllerId1, _ := hex.DecodeString("00004ca4") // (4ca4) - Set by PDM
	controllerId2, _ := hex.DecodeString("00004ca6") // (4ca4) - Set by PDM
	controllerId3, _ := hex.DecodeString("fffffffe") // (4ca4) - Set by PDM

	testControllerId(controllerId1, sharedSecret)
	testControllerId(controllerId2, sharedSecret)
	testControllerId(controllerId3, sharedSecret)
}

func testControllerId(controllerId []byte, sharedSecret []byte) {
	// testKeyDerivation(controllerId, podPublic, podPublic, sharedSecret)
	testKeyDerivation(controllerId, podPublic, pdmPublic, sharedSecret)
	// testKeyDerivation(controllerId, podPublic, pdmNonce, sharedSecret)
	// testKeyDerivation(controllerId, podPublic, podNonce, sharedSecret)
	// testKeyDerivation(controllerId, podPublic, firmwareId, sharedSecret)
	// testKeyDerivation(controllerId, pdmPublic, pdmPublic, sharedSecret)
	testKeyDerivation(controllerId, pdmPublic, podPublic, sharedSecret)
	// testKeyDerivation(controllerId, pdmPublic, pdmNonce, sharedSecret)
	// testKeyDerivation(controllerId, pdmPublic, podNonce, sharedSecret)
	// testKeyDerivation(controllerId, pdmPublic, firmwareId, sharedSecret)
	// testKeyDerivation(controllerId, pdmNonce, pdmNonce, sharedSecret)
	// testKeyDerivation(controllerId, pdmNonce, podNonce, sharedSecret)
	// testKeyDerivation(controllerId, pdmNonce, podPublic, sharedSecret)
	// testKeyDerivation(controllerId, pdmNonce, pdmPublic, sharedSecret)
	// testKeyDerivation(controllerId, pdmNonce, firmwareId, sharedSecret)
	// testKeyDerivation(controllerId, podNonce, podNonce, sharedSecret)
	// testKeyDerivation(controllerId, podNonce, pdmNonce, sharedSecret)
	// testKeyDerivation(controllerId, podNonce, pdmPublic, sharedSecret)
	// testKeyDerivation(controllerId, podNonce, podPublic, sharedSecret)
	// testKeyDerivation(controllerId, podNonce, firmwareId, sharedSecret)
	// testKeyDerivation(controllerId, firmwareId, firmwareId, sharedSecret)
	// testKeyDerivation(controllerId, firmwareId, podNonce, sharedSecret)
	// testKeyDerivation(controllerId, firmwareId, pdmNonce, sharedSecret)
	// testKeyDerivation(controllerId, firmwareId, pdmPublic, sharedSecret)
	// testKeyDerivation(controllerId, firmwareId, podPublic, sharedSecret)
}

func testKeyDerivation(controllerId []byte, key1 []byte, key2 []byte, sharedSecret []byte) {
	hash := sha256.New()
	lengthBytes := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	lengthBytes[7] = byte(len(firmwareId))
	hash.Write(lengthBytes) // length 6 at pos 8
	hash.Write(firmwareId)  // firmwareId (9b0ab96a76f4)
	lengthBytes[7] = byte(len(controllerId))
	hash.Write(lengthBytes)  // length 6 at pos 8
	hash.Write(controllerId) // controllerId
	lengthBytes[7] = byte(len(key1))
	hash.Write(lengthBytes) // length 6 at pos 8
	hash.Write(key1)        // key 1
	lengthBytes[7] = byte(len(key2))
	hash.Write(lengthBytes) // length 6 at pos 8
	hash.Write(key2)        // key 2
	lengthBytes[7] = byte(len(sharedSecret))
	hash.Write(lengthBytes) // length 6 at pos 8
	hash.Write(sharedSecret)
	derivedKey := hash.Sum(nil)
	// log.Infof("DerivedKey: %x :: %d", derivedKey, len(derivedKey))
	confKey := derivedKey[:16]
	ltk := derivedKey[16:]
	// log.Infof("ConfKey: %x :: %d", confKey, len(confKey))
	// log.Infof("LTK:     %x :: %d", ltk, len(ltk))

	testConfKey(confKey)
	testConfKey(ltk)
}

func testConfKey(key []byte) {
	// First byte could be 1 or 2. Don't know which nonce is first.
	nonce0 := make([]byte, 0)
	nonce0 = append(nonce0, 0x00)
	nonce0 = append(nonce0, podNonce[:6]...)
	nonce0 = append(nonce0, pdmNonce[:6]...)

	nonce1 := make([]byte, 0)
	nonce1 = append(nonce1, 0x01)
	nonce1 = append(nonce1, podNonce[:6]...)
	nonce1 = append(nonce1, pdmNonce[:6]...)

	nonce3 := make([]byte, 0)
	nonce3 = append(nonce3, 0x02)
	nonce3 = append(nonce3, podNonce[:6]...)
	nonce3 = append(nonce3, pdmNonce[:6]...)

	nonce5 := make([]byte, 0)
	nonce5 = append(nonce5, 0x00)
	nonce5 = append(nonce5, pdmNonce[:6]...)
	nonce5 = append(nonce5, podNonce[:6]...)

	nonce2 := make([]byte, 0)
	nonce2 = append(nonce2, 0x01)
	nonce2 = append(nonce2, pdmNonce[:6]...)
	nonce2 = append(nonce2, podNonce[:6]...)

	//Most likely correct
	nonce4 := make([]byte, 0)
	nonce4 = append(nonce4, 0x02)
	nonce4 = append(nonce4, pdmNonce[:6]...)
	nonce4 = append(nonce4, podNonce[:6]...)

	nonce6 := make([]byte, 0)
	nonce6 = append(nonce6, 0x00)
	nonce6 = append(nonce6, podNonce[10:]...)
	nonce6 = append(nonce6, pdmNonce[10:]...)

	nonce7 := make([]byte, 0)
	nonce7 = append(nonce7, 0x01)
	nonce7 = append(nonce7, podNonce[10:]...)
	nonce7 = append(nonce7, pdmNonce[10:]...)

	nonce8 := make([]byte, 0)
	nonce8 = append(nonce8, 0x02)
	nonce8 = append(nonce8, podNonce[10:]...)
	nonce8 = append(nonce8, pdmNonce[10:]...)

	nonce9 := make([]byte, 0)
	nonce9 = append(nonce9, 0x00)
	nonce9 = append(nonce9, pdmNonce[10:]...)
	nonce9 = append(nonce9, podNonce[10:]...)

	nonce10 := make([]byte, 0)
	nonce10 = append(nonce10, 0x01)
	nonce10 = append(nonce10, pdmNonce[10:]...)
	nonce10 = append(nonce10, podNonce[10:]...)

	//Most likely correct
	nonce11 := make([]byte, 0)
	nonce11 = append(nonce11, 0x02)
	nonce11 = append(nonce11, pdmNonce[10:]...)
	nonce11 = append(nonce11, podNonce[10:]...)

	for i := 8; i <= 8; i++ {
		testCCMOpen(key, i, nonce0)
		testCCMOpen(key, i, nonce1)
		testCCMOpen(key, i, nonce2)
		testCCMOpen(key, i, nonce3)
		testCCMOpen(key, i, nonce4)
		testCCMOpen(key, i, nonce6)
		testCCMOpen(key, i, nonce7)
		testCCMOpen(key, i, nonce8)
		testCCMOpen(key, i, nonce9)
		testCCMOpen(key, i, nonce10)
		testCCMOpen(key, i, nonce11)
	}
}

func testCCMOpen(key []byte, tagSize int, nonce []byte) {
	aes, _ := aes.NewCipher(key)
	accm, _ := aesccm.NewCCM(aes, tagSize, 13)

	// content := receivedSPS2[:len(receivedSPS2)-8]
	// tag := receivedSPS2[len(receivedSPS2)-8:]
	// cert := receivedSPS2[151:215]
	// asn1der := receivedSPS2[:151]
	// log.Infof("content: %x :: %d", content, len(content))
	// log.Infof("asn1der: %x :: %d", asn1der, len(asn1der))
	// log.Infof("cert: %x :: %d", cert, len(cert))
	// log.Infof("tag: %x :: %d", tag, len(tag))
	var dst []byte
	r, err := accm.Open(nil, nonce, receivedSPS2, nil)
	// log.Infof("nonce: %x :: %d", nonce, len(nonce))
	// log.Infof("r: %x :: %d", r, len(r))
	// log.Infof("dst: %x :: %d", dst, len(dst))
	// log.Infof("Error :%s", spew.Sdump(err))
	if err == nil {
		log.Infof("SUCCESS!!! r: %x :: %d", r, len(r))
		log.Infof("dst: %x :: %d", dst, len(dst))
	}

	// for x := 223; x >= 0; x-- {
	// 	a := receivedSPS2[:x]
	// 	b := receivedSPS2[x:]
	// 	result1, err := accm.Open(nil, nonce, b, a)
	// 	if err != nil {
	// 		// log.Infof("Error :%s", spew.Sdump(err))
	// 	} else {
	// 		// If this works we've got everything correct.
	// 		log.Infof("SUCCESS!! tagsize: %d, key: %x, nonce: %x", tagSize, key, nonce)
	// 		log.Infof("result: %x :: %d", result1, len(result1))
	// 	}
	// 	result2, err := accm.Open(nil, nonce, a, b)
	// 	if err != nil {
	// 		// log.Infof("Error :%s", spew.Sdump(err))
	// 	} else {
	// 		// If this works we've got everything correct.
	// 		log.Infof("SUCCESS!! tagsize: %d, key: %x, nonce: %x", tagSize, key, nonce)
	// 		log.Infof("result: %x :: %d", result2, len(result2))
	// 	}
	// 	result3, err := accm.Open(nil, nonce, receivedSPS2, b)
	// 	if err != nil {
	// 		// log.Infof("Error :%s", spew.Sdump(err))
	// 	} else {
	// 		// If this works we've got everything correct.
	// 		log.Infof("SUCCESS!! tagsize: %d, key: %x, nonce: %x", tagSize, key, nonce)
	// 		log.Infof("result: %x :: %d", result3, len(result3))
	// 	}
	// }
}
