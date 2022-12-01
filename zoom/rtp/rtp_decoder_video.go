package rtp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/protocol"
	"github.com/pion/rtp"
)

func DecodeVideo(rtpPacket *rtp.Packet, depacketizer *protocol.NaluPacketizer, decryptor *protocol.AesGcmCrypto) (decoded []byte, err error) {
	// 1. Decode RTP extensions
	extensions := rtpPacket.GetExtensionIDs()
	if len(extensions) != 0 {
		log.Printf("warn: DecodeVideo expected 0 extensions to be provided, instead got extension ids = %v", extensions)
	}

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
	complete, err := depacketizer.Unmarshal(payload)
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
	err = decodedPayload.Unmarshal(complete)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 5. Decrypt the ciphertext
	log.Printf("body iv=%v body=%v tag=%v", hex.EncodeToString(decodedPayload.IV), hex.EncodeToString(decodedPayload.Ciphertext), hex.EncodeToString(decodedPayload.Tag))
	ciphertextWithTag := append(decodedPayload.Ciphertext, decodedPayload.Tag...)
	plaintext, err := decryptor.Decrypt(decodedPayload.IV, ciphertextWithTag)
	if err != nil {
		return nil, err
	}
	log.Printf("rtp: decrypted=%v", hex.EncodeToString(plaintext))

	decoded = plaintext
	return
}
