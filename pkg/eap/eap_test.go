package eap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEapAka_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		eap  *EapAka
	}{
		{
			name: "emtpy",
			eap:  &EapAka{},
		},
		{
			name: "success",
			eap: &EapAka{
				Code: CodeSuccess,
			},
		},
		{
			name: "challenge",
			eap: &EapAka{
				Code:    CodeRequest,
				SubType: SubTypeAkaChallenge,
				Attributes: map[AttributeType]*Attribute{
					AT_RAND: &Attribute{
						Data: make([]byte, 16),
					},
					AT_CUSTOM_IV: &Attribute{
						Data: make([]byte, 4),
					},
					AT_AUTN: &Attribute{
						Data: make([]byte, 16),
					},
				},
			},
		},
		{
			name: "response",
			eap: &EapAka{
				Code:    CodeResponse,
				SubType: SubTypeAkaChallenge,
				Attributes: map[AttributeType]*Attribute{
					AT_RES: &Attribute{
						Data: make([]byte, 8),
					},
					AT_CUSTOM_IV: &Attribute{
						Data: make([]byte, 4),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.eap.Marshal()
			if err != nil {
				t.Errorf("EapAka.Marshal() error = %v", err)
				return
			}

			back, err := Unmarshal(got)
			if diff := cmp.Diff(tt.eap, back); diff != "" {
				t.Errorf("EapAka.Unmarshal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
