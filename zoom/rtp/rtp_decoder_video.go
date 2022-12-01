package rtp

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

func DecodeVideo(rtpPacket *rtp.Packet, sampleBuilder *samplebuilder.SampleBuilder) (*media.Sample, error) {
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

	// 2. Push the RTP packet to the sampleBuilder
	sampleBuilder.Push(rtpPacket)

	// 3. Pop a sample, may return nil
	return sampleBuilder.Pop(), nil
}
