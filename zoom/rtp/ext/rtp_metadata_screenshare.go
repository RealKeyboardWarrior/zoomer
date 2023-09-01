package ext

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/pion/rtp"
)

func DecodeScreenShareMetadata(rtpPacket *rtp.Packet) (*RtpMetadata, error) {
	// 1. Decode RTP extensions
	extensions := rtpPacket.GetExtensionIDs()
	for _, id := range extensions {
		switch id {
		case RTP_EXTENSION_ID_UUID, RTP_EXTENSION_ID_SCREENSHARE_RESOLUTION, RTP_EXTENSION_ID_SCREENSHARE_FRAME_INFO:
		default:
			extensionData := rtpPacket.GetExtension(id)
			log.Printf("rtp extensions found unknown ext id=%v data=%v", id, hex.EncodeToString(extensionData))
		}
	}

	id := rtpPacket.GetExtension(RTP_EXTENSION_ID_UUID)

	resolutionBytes := rtpPacket.GetExtension(RTP_EXTENSION_ID_SCREENSHARE_RESOLUTION)
	var resolutionMeta *RtpExtResolution
	if len(resolutionBytes) > 0 {
		resolutionMeta = &RtpExtResolution{}
		err := resolutionMeta.Unmarshal(resolutionBytes)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}
	svcBytes := rtpPacket.GetExtension(RTP_EXTENSION_ID_SCREENSHARE_FRAME_INFO)
	var svcMeta *RtpExtFrameInfo
	if len(svcBytes) > 0 {
		svcMeta = &RtpExtFrameInfo{}
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
	} else if rtpPacket.PayloadType == 99 {
		// Expected payload format
	} else {
		return nil, fmt.Errorf("payload type has unexpected value %v", rtpPacket.PayloadType)
	}

	return &RtpMetadata{
		StreamId:              id,
		ScreenShareResolution: resolutionMeta,
		ScreenShareFrameInfo:  svcMeta,
	}, nil
}
