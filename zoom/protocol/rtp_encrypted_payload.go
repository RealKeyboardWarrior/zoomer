package protocol

import (
	"encoding/binary"
	"fmt"
)

var (
	PREFIX = []byte{0, 0}
	SUFFIX = []byte{0}
)

const (
	LEN_PREFIX = 2 // Should match len(PREFIX)
	LEN_SUFFIX = 1 // Should match len(SUFFIX)
	LEN_BODY   = 2
	LEN_LEN_IV = 1
	LEN_IV     = 12
	LEN_TAG    = 16
)

type RtpEncryptedPayload struct {
	IV         []byte
	Ciphertext []byte
	Tag        []byte
}

func NewRtpEncryptedPayload(IV []byte, CiphertextWithTag []byte) *RtpEncryptedPayload {
	// TODO: assert it has at least length of tag or more
	Ciphertext := CiphertextWithTag[:len(CiphertextWithTag)-LEN_TAG]
	Tag := CiphertextWithTag[len(CiphertextWithTag)-LEN_TAG:]
	return &RtpEncryptedPayload{
		IV,
		Ciphertext,
		Tag,
	}
}

func (encryptedPayload *RtpEncryptedPayload) Unmarshal(payload []byte) error {
	if payload == nil {
		return ErrNoData
	}

	lenActualHeader := LEN_BODY + LEN_LEN_IV + LEN_IV
	lenHeader := LEN_PREFIX /* 00 prefix */ + lenActualHeader /* actual header */ + LEN_SUFFIX /* 0 suffix */

	if len(payload) < lenHeader {
		return fmt.Errorf("payload does not have required header size")
	}
	header := payload[LEN_PREFIX:lenActualHeader]
	lenIV := int(header[2])

	if lenIV != LEN_IV {
		return fmt.Errorf("payload does not have IV size of %v received %v instead", LEN_IV, lenIV)
	}
	IV := header[3 : 3+lenIV]

	lenBody := int(binary.BigEndian.Uint16(header[0:2]))
	if len(payload) < lenHeader+lenBody {
		return fmt.Errorf("payload does not have required body size")
	}

	ciphertext := payload[lenHeader : lenHeader+lenBody]
	tag := payload[lenHeader+lenBody:]
	if len(tag) != LEN_TAG {
		return fmt.Errorf("payload does not have tag size of %v received %v instead", LEN_TAG, len(tag))
	}

	encryptedPayload.IV = IV
	encryptedPayload.Ciphertext = ciphertext
	encryptedPayload.Tag = tag
	return nil
}

func (encryptedPayload *RtpEncryptedPayload) Marshal() []byte {
	// Parse the length of the ciphertext to a big endian uint16 bytes
	lenBody := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBody, uint16(len(encryptedPayload.Ciphertext)))

	// Parse the length of the IV to a big endian uint8 byte
	lenIV := []byte{uint8(len(encryptedPayload.IV))}
	// TODO: might need range checks to ensure validity

	/*
	 Finalize encrypted payload
	 [ 00 00 PREFIX ] [ length body ] [ length iv ] [ iv ] [ 00 SUFFIX ] [ ciphertext ] [ tag ]
	*/
	header := append(PREFIX, lenBody...)
	header = append(header, lenIV...)
	header = append(header, encryptedPayload.IV...)
	header = append(header, SUFFIX...)

	encodedPayload := append(header, encryptedPayload.Ciphertext...)
	encodedPayload = append(encodedPayload, encryptedPayload.Tag...)
	return encodedPayload
}
