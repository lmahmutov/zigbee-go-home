package ncp

import (
	"bytes"
	"testing"
)

func TestHDLCEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"simple", []byte{0x01, 0x02, 0x03}},
		{"with flag byte", []byte{0x7E, 0x01}},
		{"with escape byte", []byte{0x7D, 0x02}},
		{"mixed special", []byte{0x00, 0x7E, 0x7D, 0xFF}},
		{"empty payload", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := hdlcEncode(tt.data)

			// Verify framing: starts and ends with flag
			if encoded[0] != hdlcFlag || encoded[len(encoded)-1] != hdlcFlag {
				t.Errorf("missing flags: %X", encoded)
			}

			// Strip flags for decode
			inner := encoded[1 : len(encoded)-1]
			decoded, err := hdlcDecode(inner)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if !bytes.Equal(decoded, tt.data) {
				t.Errorf("round trip failed: got %X, want %X", decoded, tt.data)
			}
		})
	}
}

func TestHDLCDecodeBadFCS(t *testing.T) {
	encoded := hdlcEncode([]byte{0x01, 0x02})
	inner := encoded[1 : len(encoded)-1]
	// Corrupt last byte (part of FCS)
	if len(inner) > 0 {
		inner[len(inner)-1] ^= 0xFF
	}
	_, err := hdlcDecode(inner)
	if err == nil {
		t.Error("expected FCS error")
	}
}

