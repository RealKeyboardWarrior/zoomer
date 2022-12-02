package rtp

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/codecs/h264"
	"github.com/RealKeyboardWarrior/zoomer/zoom/crypto"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

type ZoomRtpDecoder struct {
	streamType        StreamType
	sampleBuilders    map[ /*ssrc*/ uint32]*samplebuilder.SampleBuilder
	ParticipantRoster *ZoomParticipantRoster
}

func NewZoomRtpDecoder(streamType StreamType) *ZoomRtpDecoder {
	return &ZoomRtpDecoder{
		sampleBuilders:    make(map[uint32]*samplebuilder.SampleBuilder),
		ParticipantRoster: NewParticipantRoster(),
		streamType:        streamType,
	}
}

func (parser *ZoomRtpDecoder) getDecryptorFor(ssrc uint32) (*crypto.AesGcmCrypto, error) {
	// 1. Fetch the secretNonce for the ssrc
	secretNonce, err := parser.ParticipantRoster.GetSecretNonceForSSRC((int)(ssrc))
	if err != nil {
		return nil, fmt.Errorf("cannot decode packet for ssrc=%v", ssrc)
	}

	// 2. Fetch the shared meeting key
	sharedMeetingKey, err := parser.ParticipantRoster.GetSharedMeetingKey()
	if err != nil {
		return nil, err
	}

	// 3. Build decryptor per key type
	var keyType crypto.AesKeyType
	switch parser.streamType {
	case STREAM_TYPE_SCREENSHARE:
		keyType = crypto.KEY_TYPE_SCREENSHARE
	case STREAM_TYPE_VIDEO:
		keyType = crypto.KEY_TYPE_VIDEO
	case STREAM_TYPE_AUDIO:
		keyType = crypto.KEY_TYPE_AUDIO
	}

	log.Printf("key=%v sn=%v", hex.EncodeToString(sharedMeetingKey), hex.EncodeToString(secretNonce))
	decryptor, err := crypto.NewAesGcmCrypto(sharedMeetingKey, secretNonce, keyType)
	if err != nil {
		return nil, err
	}
	return decryptor, nil
}

func (parser *ZoomRtpDecoder) getSampleBuilderFor(ssrc uint32) (*samplebuilder.SampleBuilder, error) {
	if parser.sampleBuilders[ssrc] == nil {
		// TODO: think about reasonable max late
		maxLate := uint16(400)
		decryptor, err := parser.getDecryptorFor(ssrc)
		if err != nil {
			return nil, err
		}

		// TODO: add audio support
		var depacketizer rtp.Depacketizer
		switch parser.streamType {
		case STREAM_TYPE_SCREENSHARE, STREAM_TYPE_VIDEO:
			depacketizer = h264.NewVideoDepacketizer(decryptor)
		case STREAM_TYPE_AUDIO:
			panic("unimplemented")
		}

		// TODO: fix sample rate
		sampleRate := uint32(1)
		parser.sampleBuilders[ssrc] = samplebuilder.New(maxLate, depacketizer, sampleRate)
	}
	return parser.sampleBuilders[ssrc], nil
}

func (parser *ZoomRtpDecoder) Decode(rawPkt []byte) (*media.Sample, error) {
	// 1. Decode the RTP packet
	rtpPacket := &rtp.Packet{}
	err := rtpPacket.Unmarshal(rawPkt)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Printf("rtp header [M = %v] [PT type=%v] [SN seq=%v] [TS timestamp=%v] [P padding=%v size=%v] [ssrc=%v csrc=%v]", rtpPacket.Marker, rtpPacket.PayloadType, rtpPacket.SequenceNumber, rtpPacket.Timestamp, rtpPacket.Padding, rtpPacket.PaddingSize, rtpPacket.SSRC, rtpPacket.CSRC)
	log.Printf("rtp payload [PYLD size=%v data=%v]", len(rtpPacket.Payload), hex.EncodeToString(rtpPacket.Payload))

	// 2.
	sampleBuilder, err := parser.getSampleBuilderFor(rtpPacket.SSRC)
	if err != nil {
		return nil, err
	}

	// 3. Call the decoder
	switch parser.streamType {
	case STREAM_TYPE_SCREENSHARE:
		return DecodeScreenShare(rtpPacket, sampleBuilder)
	case STREAM_TYPE_VIDEO:
		return DecodeVideo(rtpPacket, sampleBuilder)
	case STREAM_TYPE_AUDIO:
		panic("unimplemented")
	}
	return nil, fmt.Errorf("unexpected end")
}
