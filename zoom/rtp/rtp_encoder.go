package rtp

import (
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/crypto"
	"github.com/RealKeyboardWarrior/zoomer/zoom/rtp/ext"
	"github.com/pion/rtp"
)

type ZoomRtpEncoder struct {
	id             int
	ssrc           int
	resolution     *ext.RtpExtResolution
	messageCounter int
	timestamp      int

	// Frame information
	baseFrameCounter    int
	currentFrameCounter int

	// Encryption information
	ParticipantRoster *ZoomParticipantRoster
}

func NewZoomRtpEncoder(roster *ZoomParticipantRoster, ssrc int, id int, width, height int) *ZoomRtpEncoder {
	return &ZoomRtpEncoder{
		id:   id,
		ssrc: ssrc,
		resolution: &ext.RtpExtResolution{
			Width:  uint16(width),
			Height: uint16(height),
		},

		messageCounter: 0,
		timestamp:      2746202358,

		ParticipantRoster: roster,
	}
}

func (parser *ZoomRtpEncoder) Encode(payload []byte) ([]byte, error) {

	encryptedPayload, err := encryptPayloadWithRoster(parser.ParticipantRoster, parser.ssrc, parser.messageCounter, payload)
	if err != nil {
		return nil, err
	}

	// TODO: properly NALU packetize the payload if too long
	packetizedPayload := append(make([]byte, 1), encryptedPayload...)

	// Wrap encoded payload in RTP packets
	// TODO: multiple packets, currently only single packets
	parser.timestamp = parser.timestamp + (parser.messageCounter * 15000)
	p := &rtp.Packet{
		Header: rtp.Header{
			Version:          2,
			Padding:          false,
			Extension:        false,
			Marker:           false,
			PayloadType:      99,
			SequenceNumber:   uint16(parser.messageCounter),
			Timestamp:        uint32(parser.timestamp),
			SSRC:             uint32(parser.ssrc),
			CSRC:             []uint32{},
			ExtensionProfile: 0,
			Extensions:       []rtp.Extension{},
		},
		Payload:     packetizedPayload,
		PaddingSize: 0,
	}

	// TODO: increment UUID whenever reconnecting, prob big endian but single byte?
	p.Header.SetExtension(ext.RTP_EXTENSION_ID_UUID, []byte{0x01})
	if parser.resolution != nil {
		resolution, err := parser.resolution.Marshal()
		if err != nil {
			return nil, err
		}
		p.Header.SetExtension(ext.RTP_EXTENSION_ID_RESOLUTION, resolution)
	}

	// Set extension screen size
	rtpResolution, err := parser.resolution.Marshal()
	if err != nil {
		return nil, err
	}
	p.Header.SetExtension(ext.RTP_EXTENSION_ID_RESOLUTION, rtpResolution)

	// TODO: extension frame info should probably be more advanced
	rtpFrameInfo := &ext.RtpExtFrameInfo{
		Version:       2,
		Start:         true,  // TODO: NALU packetizer
		End:           true,  // TODO: NALU packetize
		Independent:   false, // TODO: passed along with payload
		Required:      false, // TODO: passed along with payload
		Base:          false, // TODO: passed along with payload
		TemporalID:    0,     // TODO: ????
		CurrentFrame:  uint16(parser.currentFrameCounter + 1),
		PreviousFrame: uint16(parser.currentFrameCounter),
		BaseFrame:     uint16(parser.baseFrameCounter),
	}
	rtpFrameInfoBytes, err := rtpFrameInfo.Marshal()
	if err != nil {
		return nil, err
	}
	p.Header.SetExtension(ext.RTP_EXTENSION_ID_FRAME_INFO, rtpFrameInfoBytes)

	rawPkt, err := p.Marshal()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return rawPkt, nil
}

func encryptPayloadWithRoster(roster *ZoomParticipantRoster, ssrc int, messageCounter int, plaintext []byte) ([]byte, error) {
	secretNonce, err := roster.GetSecretNonceForSSRC(ssrc)
	if err != nil {
		return nil, ErrParticipantExists
	}

	sharedMeetingKey, err := roster.GetSharedMeetingKey()
	if err != nil {
		return nil, ErrParticipantExists
	}

	// Build 12 byte IV
	IV := make([]byte, 12)
	binary.BigEndian.PutUint16(IV, uint16(messageCounter))
	IV = IV[:12]

	// TODO: add keyType, hardcoded to screenshare!
	encryptor, err := crypto.NewAesGcmCrypto(sharedMeetingKey, secretNonce, 0x02)
	if err != nil {
		return nil, err
	}

	ciphertextWithTag, err := encryptor.Encrypt(IV, plaintext)
	if err != nil {
		return nil, err
	}

	if true {
		log.Printf("encrypted body key=%v sn=%v iv=%v ciphertextWithTag=%v", hex.EncodeToString(sharedMeetingKey), hex.EncodeToString(secretNonce), hex.EncodeToString(IV), hex.EncodeToString(ciphertextWithTag))
	}

	encodedPayload := crypto.NewRtpEncryptedPayload(0, IV, ciphertextWithTag)
	encodedPayloadInBytes := encodedPayload.Marshal()
	return encodedPayloadInBytes, nil
}
