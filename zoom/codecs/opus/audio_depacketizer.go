package opus

import (
	"encoding/hex"
	"log"

	"github.com/RealKeyboardWarrior/zoomer/zoom/crypto"
)

type AudioDepacketizer struct {
	decryptor *crypto.AesGcmCrypto
}

func NewAudioDepacketizer(decryptor *crypto.AesGcmCrypto) *AudioDepacketizer {
	return &AudioDepacketizer{
		decryptor: decryptor,
	}
}

func (depacketizer *AudioDepacketizer) Unmarshal(packet []byte) ([]byte, error) {
	// 1. Decode the inner encrypted payload
	decodedPayload := &crypto.RtpEncryptedPayload{}
	err := decodedPayload.Unmarshal(packet)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 2. Decrypt the ciphertext
	log.Printf("crypto: body iv=%v body=%v tag=%v", hex.EncodeToString(decodedPayload.IV), hex.EncodeToString(decodedPayload.Ciphertext), hex.EncodeToString(decodedPayload.Tag))
	ciphertextWithTag := append(decodedPayload.Ciphertext, decodedPayload.Tag...)
	plaintext, err := depacketizer.decryptor.Decrypt(decodedPayload.IV, ciphertextWithTag)
	if err != nil {
		return nil, err
	}
	log.Printf("crypto: decrypted=%v", hex.EncodeToString(plaintext))

	return plaintext, nil
}

// IsPartitionHead checks if this is the head of a packetized nalu stream.
func (*AudioDepacketizer) IsPartitionHead(payload []byte) bool {
	return true
}

// Checks if the packet is at the end of a partition.  This should
// return false if the result could not be determined.
func (*AudioDepacketizer) IsPartitionTail(marker bool, payload []byte) bool {
	return true
}
