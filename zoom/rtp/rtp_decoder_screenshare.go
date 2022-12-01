package rtp

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/rtp/ext"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

const (
	RTP_EXTENSION_ID_UUID       = 7
	RTP_EXTENSION_ID_RESOLUTION = 6
	RTP_EXTENSION_ID_FRAME_INFO = 4
)

func DecodeScreenShare(rtpPacket *rtp.Packet, sampleBuilder *samplebuilder.SampleBuilder) (*media.Sample, error) {
	// 1. Decode RTP extensions
	id := rtpPacket.GetExtension(RTP_EXTENSION_ID_UUID)

	resolutionBytes := rtpPacket.GetExtension(RTP_EXTENSION_ID_RESOLUTION)
	var resolutionMeta *ext.RtpExtResolution
	if len(resolutionBytes) > 0 {
		resolutionMeta := &ext.RtpExtResolution{}
		err := resolutionMeta.Unmarshal(rtpPacket.GetExtension(RTP_EXTENSION_ID_RESOLUTION))
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	svcBytes := rtpPacket.GetExtension(RTP_EXTENSION_ID_RESOLUTION)
	var svcMeta *ext.RtpExtFrameInfo
	if len(resolutionBytes) > 0 {
		svcMeta := &ext.RtpExtFrameInfo{}
		err := svcMeta.Unmarshal(svcBytes)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}
	log.Printf("rtp extensions [RtpId id=%v] [meta=%v] [%v]", id, svcMeta, resolutionMeta)

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
