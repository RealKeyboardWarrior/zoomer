package protocol

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func FuzzUnmarshalRtpExtResolution(f *testing.F) {
	var (
		invalidPayload  = []byte{0x00}
		validPayload, _ = hex.DecodeString("040002d0")
	)
	f.Add(invalidPayload)
	f.Add(validPayload)
	f.Fuzz(func(t *testing.T, resolutionPayload []byte) {
		extResolution := &RtpExtResolution{}
		extResolution.Unmarshal(resolutionPayload)
	})
}

func TestUnmarshalRtpExtResolution(t *testing.T) {
	var (
		validPayload, _ = hex.DecodeString("040002d0")
	)
	extResolution := &RtpExtResolution{}
	err := extResolution.Unmarshal(validPayload)
	if err != nil {
		t.Error(err)
		return
	}

	expectedExtResolution := &RtpExtResolution{
		Width:  1024,
		Height: 720,
	}

	if *extResolution != *expectedExtResolution {
		t.Error("Does not equal")
		return
	}

}

func TestMarshalRtpExtResolution(t *testing.T) {
	var (
		validPayload, _ = hex.DecodeString("040002d0")
	)

	extResolution := &RtpExtResolution{
		Width:  1024,
		Height: 720,
	}
	extResolutionBytes, err := extResolution.Marshal()
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(extResolutionBytes, validPayload) {
		t.Error("Plaintext did not match expected")
	}
}
