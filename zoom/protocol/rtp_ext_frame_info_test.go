package protocol

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func FuzzUnmarshalRtpExtensionFrameInfo(f *testing.F) {
	var (
		invalidPayload  = []byte{0x00}
		validPayload, _ = hex.DecodeString("02c007030702055f")
	)
	f.Add(invalidPayload)
	f.Add(validPayload)
	f.Fuzz(func(t *testing.T, encryptedPayload []byte) {
		extFrameInfo := &RtpExtFrameInfo{}
		extFrameInfo.Unmarshal(encryptedPayload)
	})
}

func TestUnmarshalRtpExtensionFrameInfo(t *testing.T) {
	var (
		validPayload, _ = hex.DecodeString("02c007030702055f")
	)
	extFrameInfo := &RtpExtFrameInfo{}
	err := extFrameInfo.Unmarshal(validPayload)
	if err != nil {
		t.Error(err)
		return
	}

	expectedExtFrameInfo := &RtpExtFrameInfo{
		Version:       uint8(2),
		Start:         true,
		End:           true,
		Independent:   false,
		Required:      false,
		Base:          false,
		TemporalID:    uint8(0),
		CurrentFrame:  uint16(1795),
		PreviousFrame: uint16(1794),
		BaseFrame:     uint16(1375),
	}

	if *extFrameInfo != *expectedExtFrameInfo {
		t.Error("Does not equal")
		return
	}

}

func TestMarshalRtpExtensionFrameInfo(t *testing.T) {
	var (
		validPayload, _ = hex.DecodeString("02c007030702055f")
	)

	extFrameInfo := &RtpExtFrameInfo{
		Version:       uint8(2),
		Start:         true,
		End:           true,
		Independent:   false,
		Required:      false,
		Base:          false,
		TemporalID:    uint8(0),
		CurrentFrame:  uint16(1795),
		PreviousFrame: uint16(1794),
		BaseFrame:     uint16(1375),
	}

	payload, err := extFrameInfo.Marshal()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(payload, validPayload) {
		t.Error("payload did not match expected")
	}
}
