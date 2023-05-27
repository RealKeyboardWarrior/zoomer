package ext

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/pion/rtp"
)

func DecodeVideoMetadata(rtpPacket *rtp.Packet) (*RtpMetadata, error) {
	// 1. Decode RTP extensions
	extensions := rtpPacket.GetExtensionIDs()
	if len(extensions) != 0 {
		log.Printf("warn: DecodeVideo expected 0 extensions to be provided, instead got extension ids = %v", extensions)
	}

	if rtpPacket.PayloadType == 110 {
		log.Printf("rtp [PT type=10] payload=%v", hex.EncodeToString(rtpPacket.Payload))
		return nil, nil
	} else if rtpPacket.PayloadType == 98 {
		// Expected payload format
	} else {
		return nil, fmt.Errorf("payload type has unexpected value %v", rtpPacket.PayloadType)
	}

	return &RtpMetadata{}, nil
}
