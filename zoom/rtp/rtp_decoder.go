package rtp

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/codecs/h264"
	"github.com/RealKeyboardWarrior/zoomer/zoom/codecs/opus"
	"github.com/RealKeyboardWarrior/zoomer/zoom/crypto"
	"github.com/RealKeyboardWarrior/zoomer/zoom/rtp/ext"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

type ZoomRtpDecoder struct {
	streamType        StreamType
	sampleBuilders    map[ /*ssrc*/ uint32]*samplebuilder.SampleBuilder
	decryptors        map[ /*ssrc*/ uint32]*crypto.AesGcmCrypto
	ParticipantRoster *ZoomParticipantRoster
}

func NewZoomRtpDecoder(streamType StreamType) *ZoomRtpDecoder {
	return &ZoomRtpDecoder{
		sampleBuilders:    make(map[uint32]*samplebuilder.SampleBuilder),
		decryptors:        make(map[uint32]*crypto.AesGcmCrypto),
		ParticipantRoster: NewParticipantRoster(),
		streamType:        streamType,
	}
}

func (parser *ZoomRtpDecoder) getDecryptorFor(ssrc uint32) (*crypto.AesGcmCrypto, error) {
	if parser.decryptors[ssrc] == nil {
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
		parser.decryptors[ssrc] = decryptor
	}

	return parser.decryptors[ssrc], nil
}

func (parser *ZoomRtpDecoder) getSampleBuilderFor(ssrc uint32) (*samplebuilder.SampleBuilder, error) {
	if parser.sampleBuilders[ssrc] == nil {
		decryptor, err := parser.getDecryptorFor(ssrc)
		if err != nil {
			return nil, err
		}

		// TODO: think about reasonable max late
		maxLate := uint16(400)
		// TODO: add audio support
		var depacketizer rtp.Depacketizer
		switch parser.streamType {
		case STREAM_TYPE_SCREENSHARE, STREAM_TYPE_VIDEO:
			depacketizer = h264.NewVideoDepacketizer(decryptor)
		case STREAM_TYPE_AUDIO:
			depacketizer = opus.NewAudioDepacketizer(decryptor)
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

	// 2. Retrieve the sampler builder for ssrc & codec
	sampleBuilder, err := parser.getSampleBuilderFor(rtpPacket.SSRC)
	if err != nil {
		return nil, err
	}

	// 3. Decode Metadata
	var metadata *ext.RtpMetadata
	switch parser.streamType {
	case STREAM_TYPE_SCREENSHARE:
		metadata, err = ext.DecodeScreenShareMetadata(rtpPacket)
	case STREAM_TYPE_VIDEO:
		metadata, err = ext.DecodeVideoMetadata(rtpPacket)
	case STREAM_TYPE_AUDIO:
		metadata, err = ext.DecodeAudioMetadata(rtpPacket)
	}
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		log.Printf("metadata is nil, returning")
		return nil, nil
	}

	// 3b. We need to rewrite the rtpPacket a bit for audio because
	// because the sample builder is always lagging one packet behind.
	// so the rtpPacket we have here != the sample we receive out.
	// This is required because the IV is stored inside an RTP extension
	// and the rtp.Depacketizer interface only passes the payload of the RTP
	// packet so we just shag it into the payload so our sample builder
	// keeps the IV and payload in sync. We can't do it after sampleBuilder.Pop().
	if len(metadata.AudioHeaderEncryptedPayload) > 0 {
		clonedRtpPacket := rtpPacket.Clone()
		clonedRtpPacket.Payload = crypto.RewriteAudioHack(metadata.AudioHeaderEncryptedPayload, clonedRtpPacket.Payload)
		rtpPacket = clonedRtpPacket
	}

	// 4. Push the RTP packet to the sampleBuilder, this re-orders the packets based
	// on the RTP Sequence, they may arrive out of order aggregates them and calls
	// the correct depacketizer (VideoDepacketizer / AudioDepacketizer).
	sampleBuilder.Push(rtpPacket)

	// 5. Pop a sample, may return nil if no packets are ready yet.
	sample := sampleBuilder.Pop()
	// WARNING: SAMPLE THAT WAS POPPED MAY NOT BE ASSOCIATED WITH THE RTP PACKET!
	// THIS IS BECAUSE THE SAMPLE BUILDER ALWAYS LAGS BEHIND ONE PACKET AND MAY
	// AGGREGATE PACKETS IN THE CASE OF VIDEO STREAMS.

	return sample, nil
}
