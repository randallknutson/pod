package message

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_PrintfromByteArray(t *testing.T) {
	tests := []string{
		"54,57,00,03,00,00,06,e0,ff,ff,ff,fe,00,00,02,42,00,00,00,00,53,50,53,31,3d,00,30,2f,e5,7d,a3,47,cd,62,43,15,28,da,ac,5f,bb,29,07,30,ff,f6,84,af,c4,cf,c2,ed,90,99,5f,58,cb,3b,74,00,00,00,00,00,00,00,00,00,00,00,00,00,00,00,00",
		"01,26,00,38,17,01,00,00,02,05,00,00,ed,52,e3,f5,c2,87,b9,b9,4d,2d,4f,54,88,d4,b3,94,01,05,00,00,0c,4d,fe,2b,25,4e,d0,df,07,b4,8b,c5,88,38,fb,fd,7e,02,00,00,7c,b1,8d,f9",
	}
	for _, test := range tests {
		h := strings.Replace(test, ",", "", -1)
		h = strings.Replace(h, " ", "", -1)
		in, err := hex.DecodeString(h)
		if err != nil {
			t.Error(err)
		}
		msg, err := Unmarshal(in)
		if err != nil {
			t.Error(err)
		}
		spew.Dump(msg)
	}
}
