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
	for _, id := range extensions {
		switch id {
		default:
			extensionData := rtpPacket.GetExtension(id)
			log.Printf("rtp extensions found unknown ext id=%v data=%v", id, hex.EncodeToString(extensionData))
		}
	}
	if rtpPacket.PayloadType == 110 {
		log.Printf("rtp [PT type=%v] payload=%v", rtpPacket.PayloadType, hex.EncodeToString(rtpPacket.Payload))
		return nil, nil
	} else if rtpPacket.PayloadType == 98 {
		// Expected payload format
	} else {
		return nil, fmt.Errorf("payload type has unexpected value %v", rtpPacket.PayloadType)
	}

	return &RtpMetadata{}, nil
}
