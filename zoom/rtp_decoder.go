package zoom

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/protocol"
	"github.com/pion/rtp"
)

/*type ZoomHeader struct {
	prefix  uint16
	bodyLen uint16
	ivLen   uint8
	IV      [12]byte
}
*/
const (
	RTP_EXTENSION_ID_UUID       = 7
	RTP_EXTENSION_ID_RESOLUTION = 6
	RTP_EXTENSION_ID_FRAME_INFO = 4
)

type ZoomRtpDecoder struct {
	isScreenShare     bool
	NaluPacketizer    *protocol.NaluPacketizer
	ParticipantRoster *ZoomParticipantRoster
}

func NewZoomRtpDecoder(participantRoster *ZoomParticipantRoster, isScreenShare bool) *ZoomRtpDecoder {
	return &ZoomRtpDecoder{
		// TODO: NALU packetizer thinks its one big stream, so doesnt multiplex fragmented units
		// very well here, maybe needs to map ssrc -> packetizer
		NaluPacketizer:    protocol.NewNaluPacketizer(),
		ParticipantRoster: participantRoster,
		isScreenShare:     isScreenShare,
	}
}

func (parser *ZoomRtpDecoder) SetSharedMeetingKey(k []byte) {
	parser.ParticipantRoster.SetSharedMeetingKey(k)
}

func (parser *ZoomRtpDecoder) Decode(rawPkt []byte) (decoded []byte, err error) {
	// 1. Decode the RTP packet
	rtpPacket := &rtp.Packet{}
	err = rtpPacket.Unmarshal(rawPkt)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	id := rtpPacket.GetExtension(RTP_EXTENSION_ID_UUID)

	resolutionBytes := rtpPacket.GetExtension(RTP_EXTENSION_ID_RESOLUTION)
	var resolutionMeta *protocol.RtpExtResolution
	if len(resolutionBytes) > 0 {
		resolutionMeta := &protocol.RtpExtResolution{}
		err = resolutionMeta.Unmarshal(rtpPacket.GetExtension(RTP_EXTENSION_ID_RESOLUTION))
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	svcBytes := rtpPacket.GetExtension(RTP_EXTENSION_ID_RESOLUTION)
	var svcMeta *protocol.RtpExtFrameInfo
	if len(resolutionBytes) > 0 {
		svcMeta := &protocol.RtpExtFrameInfo{}
		err = svcMeta.Unmarshal(svcBytes)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	log.Printf("rtp header [M = %v] [PT type=%v] [SN seq=%v] [TS timestamp=%v] [P padding=%v size=%v] [ssrc=%v csrc=%v]", rtpPacket.Marker, rtpPacket.PayloadType, rtpPacket.SequenceNumber, rtpPacket.Timestamp, rtpPacket.Padding, rtpPacket.PaddingSize, rtpPacket.SSRC, rtpPacket.CSRC)
	log.Printf("rtp extensions [RtpId id=%v] [meta=%v] [%v]", id, svcMeta, resolutionMeta)
	log.Printf("rtp payload [PYLD size=%v]", len(rtpPacket.Payload))

	if rtpPacket.PayloadType == 110 {
		log.Printf("rtp [PT type=10] payload=%v", hex.EncodeToString(rtpPacket.Payload))
		return
	} else if rtpPacket.PayloadType == 98 {
		// Expected payload format
	} else {
		return nil, fmt.Errorf("payload type has unexpected value %v", rtpPacket.PayloadType)
	}

	payload := rtpPacket.Payload
	// TODO: header length
	if len(payload) < 35 {
		return nil, errors.New("payload does not have enough bytes")
	}

	// 2. Check Fragmented Unit
	complete, err := parser.NaluPacketizer.Unmarshal(payload)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 3. Exit if we don't yet have a complete packet in our NALU packetizer.
	// TODO: ideally this would be more tightly coupled with the loop that it's running in.
	if complete == nil {
		return
	}

	// 4. Decode the inner encrypted payload
	decodedPayload := &protocol.RtpEncryptedPayload{}
	decodedPayload.Unmarshal(complete)

	// 5. Decrypt the ciphertext
	secretNonce, err := parser.ParticipantRoster.GetSecretNonceForSSRC((int)(rtpPacket.SSRC))
	if err != nil {
		return nil, fmt.Errorf("cannot decode packet for ssrc=%v", rtpPacket.SSRC)
	}

	sharedMeetingKey := parser.ParticipantRoster.GetSharedMeetingKey()
	if true {
		log.Printf("body key=%v sn=%v iv=%v body=%v tag=%v", hex.EncodeToString(sharedMeetingKey), hex.EncodeToString(secretNonce), hex.EncodeToString(decodedPayload.IV), hex.EncodeToString(decodedPayload.Ciphertext), hex.EncodeToString(decodedPayload.Tag))
	}

	var keyType protocol.AesKeyType
	if parser.isScreenShare {
		keyType = protocol.KEY_TYPE_SCREENSHARE
	} else {
		keyType = protocol.KEY_TYPE_VIDEO
	}
	decryptor, err := protocol.NewAesGcmCrypto(sharedMeetingKey, secretNonce, keyType)
	if err != nil {
		return nil, err
	}

	ciphertextWithTag := append(decodedPayload.Ciphertext, decodedPayload.Tag...)
	plaintext, err := decryptor.Decrypt(decodedPayload.IV, ciphertextWithTag)
	if err != nil {
		return nil, err
	}
	log.Printf("rtp: decrypted=%v", hex.EncodeToString(plaintext))

	decoded = plaintext
	return
}
