package rtp

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/protocol"
	"github.com/pion/rtp"
)

type ZoomRtpDecoder struct {
	streamType        StreamType
	naluPacketizers   map[ /*ssrc*/ uint32]*protocol.NaluPacketizer
	ParticipantRoster *ZoomParticipantRoster
}

func NewZoomRtpDecoder(streamType StreamType) *ZoomRtpDecoder {
	return &ZoomRtpDecoder{
		naluPacketizers:   make(map[uint32]*protocol.NaluPacketizer),
		ParticipantRoster: NewParticipantRoster(),
		streamType:        streamType,
	}
}

func (parser *ZoomRtpDecoder) getNaluPacketizerFor(ssrc uint32) *protocol.NaluPacketizer {
	if parser.naluPacketizers[ssrc] == nil {
		parser.naluPacketizers[ssrc] = protocol.NewNaluPacketizer()
	}
	return parser.naluPacketizers[ssrc]
}

func (parser *ZoomRtpDecoder) Decode(rawPkt []byte) ([]byte, error) {
	// 1. Decode the RTP packet
	rtpPacket := &rtp.Packet{}
	err := rtpPacket.Unmarshal(rawPkt)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Printf("rtp header [M = %v] [PT type=%v] [SN seq=%v] [TS timestamp=%v] [P padding=%v size=%v] [ssrc=%v csrc=%v]", rtpPacket.Marker, rtpPacket.PayloadType, rtpPacket.SequenceNumber, rtpPacket.Timestamp, rtpPacket.Padding, rtpPacket.PaddingSize, rtpPacket.SSRC, rtpPacket.CSRC)
	log.Printf("rtp payload [PYLD size=%v]", len(rtpPacket.Payload))

	// 2. Fetch the secretNonce for the ssrc
	secretNonce, err := parser.ParticipantRoster.GetSecretNonceForSSRC((int)(rtpPacket.SSRC))
	if err != nil {
		return nil, fmt.Errorf("cannot decode packet for ssrc=%v", rtpPacket.SSRC)
	}

	// 3. Fetch the shared meeting key
	sharedMeetingKey, err := parser.ParticipantRoster.GetSharedMeetingKey()
	if err != nil {
		return nil, err
	}

	// 4. Build decryptor per key type
	var keyType protocol.AesKeyType
	switch parser.streamType {
	case STREAM_TYPE_SCREENSHARE:
		keyType = protocol.KEY_TYPE_SCREENSHARE
	case STREAM_TYPE_VIDEO:
		keyType = protocol.KEY_TYPE_VIDEO
	case STREAM_TYPE_AUDIO:
		keyType = protocol.KEY_TYPE_AUDIO
	}

	log.Printf("key=%v sn=%v", hex.EncodeToString(sharedMeetingKey), hex.EncodeToString(secretNonce))
	decryptor, err := protocol.NewAesGcmCrypto(sharedMeetingKey, secretNonce, keyType)
	if err != nil {
		return nil, err
	}

	// 5. Call the decoder
	switch parser.streamType {
	case STREAM_TYPE_SCREENSHARE:
		return DecodeScreenShare(rtpPacket, parser.getNaluPacketizerFor(rtpPacket.SSRC), decryptor)
	case STREAM_TYPE_VIDEO:
		return DecodeVideo(rtpPacket, parser.getNaluPacketizerFor(rtpPacket.SSRC), decryptor)
	case STREAM_TYPE_AUDIO:
		panic("unimplemented")
	}
	return nil, fmt.Errorf("unexpected end")
}
