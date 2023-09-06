package h264

import (
	"encoding/hex"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/crypto"
)

type VideoDepacketizer struct {
	naluDepacketizer *NaluPacketizer
	decryptor        *crypto.AesGcmCrypto
}

func NewVideoDepacketizer(decryptor *crypto.AesGcmCrypto) *VideoDepacketizer {
	return &VideoDepacketizer{
		naluDepacketizer: NewNaluPacketizer(),
		decryptor:        decryptor,
	}
}

func (depacketizer *VideoDepacketizer) Unmarshal(packet []byte) ([]byte, error) {
	// 1. Depacketize the NALU packets
	naluStream, err := depacketizer.naluDepacketizer.Unmarshal(packet)
	if err != nil {
		return nil, err
	}

	// 2. We received a complete encrypted payload, attempt to decrypt it.
	if len(naluStream) != 0 {
		// 3. Decode the inner encrypted payload
		decodedPayload := &crypto.RtpEncryptedPayload{}
		err = decodedPayload.Unmarshal(naluStream)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		// 4. Decrypt the ciphertext
		log.Printf("crypto: body iv=%v body=%v tag=%v", hex.EncodeToString(decodedPayload.IV), hex.EncodeToString(decodedPayload.Ciphertext), hex.EncodeToString(decodedPayload.Tag))
		ciphertextWithTag := append(decodedPayload.Ciphertext, decodedPayload.Tag...)
		plaintext, err := depacketizer.decryptor.Decrypt(decodedPayload.IV, ciphertextWithTag)
		if err != nil {
			return nil, err
		}
		log.Printf("crypto: decrypted=%v", hex.EncodeToString(plaintext))

		return plaintext, nil
	} else {
		// 5. The nalu depacketizer hasn't received the last packet of its fragmented unit yet.
		return []byte{}, nil
	}
}

// IsPartitionHead checks if this is the head of a packetized nalu stream.
func (*VideoDepacketizer) IsPartitionHead(payload []byte) bool {
	if len(payload) < 2 {
		return false
	}

	if isSingle(payload[0]) {
		return true
	} else if isFragmented(payload[0]) {
		return isFragmentedStart(payload[1])
	} else {
		log.Fatalf("H264Depacketizer IsPartitionHead received invalid payload = %v", hex.EncodeToString(payload))
	}
	return false
}

// Checks if the packet is at the end of a partition.  This should
// return false if the result could not be determined.
func (*VideoDepacketizer) IsPartitionTail(marker bool, payload []byte) bool {
	if len(payload) < 2 {
		return false
	}

	if isSingle(payload[0]) {
		return true
	} else if isFragmented(payload[0]) {
		fragmentedEnd := isFragmentedEnd(payload[1])
		if marker != fragmentedEnd {
			// This is a bit of defensive code structure, checks whether the marker bit can be used.
			// may only work on fragmented units - need to check singles.
			log.Printf("H264Depacketizer IsPartitionTail detected that the marker %v != fragmentedEnd %v", marker, fragmentedEnd)
		}
		return fragmentedEnd
	} else {
		log.Fatalf("H264Depacketizer IsPartitionTail received invalid payload = %v", hex.EncodeToString(payload))
	}
	return false
}
