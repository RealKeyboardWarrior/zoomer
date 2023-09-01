package ext

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/pion/rtp"
)

func DecodeAudioMetadata(rtpPacket *rtp.Packet) (*RtpMetadata, error) {
	// 1. Decode RTP extensions
	extensions := rtpPacket.GetExtensionIDs()
	for _, id := range extensions {
		switch id {
		default:
			extensionData := rtpPacket.GetExtension(id)
			log.Printf("rtp extensions found unknown ext id=%v data=%v", id, hex.EncodeToString(extensionData))
		}
	}

	if rtpPacket.PayloadType == 99 || rtpPacket.PayloadType == 112 {
		// Expected payload format
	} else {
		return nil, fmt.Errorf("payload type has unexpected value %v", rtpPacket.PayloadType)
	}

	audioHeader := rtpPacket.GetExtension(RTP_EXTENSION_ID_AUDIO_IV)
	if len(audioHeader) != 9 {
		return nil, fmt.Errorf("rtp extension audio iv expected length %v received %v", 9, len(audioHeader))
	}

	return &RtpMetadata{
		AudioHeaderEncryptedPayload: audioHeader,
	}, nil
}
