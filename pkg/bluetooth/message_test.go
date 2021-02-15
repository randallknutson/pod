package bluetooth

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_PrintfromByteArray(t *testing.T) {
	tests := []string{
		"54,57,00,03,00,00,06,e0,ff,ff,ff,fe,00,00,02,42,00,00,00,00,53,50,53,31,3d,00,30,2f,e5,7d,a3,47,cd,62,43,15,28,da,ac,5f,bb,29,07,30,ff,f6,84,af,c4,cf,c2,ed,90,99,5f,58,cb,3b,74,00,00,00,00,00,00,00,00,00,00,00,00,00,00,00,00",
	}
	for _, test := range tests {
		h := strings.Replace(test, ",", "", -1)
		in, err := hex.DecodeString(h)
		if err != nil {
			t.Error(err)
		}
		msg, err := fromByteArray(in)
		if err != nil {
			t.Error(err)
		}
		spew.Dump(msg)
	}
}
