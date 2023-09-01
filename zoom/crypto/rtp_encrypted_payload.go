package crypto

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
)

var (
	PREFIX    = []byte{0}
	VERSION_0 = byte(0)
	VERSION_1 = byte(1)
	SUFFIX    = []byte{0}
)

const (
	LEN_PREFIX        = 1
	LEN_VERSION       = 1 // Should match len(PREFIX)
	LEN_SUFFIX        = 1 // Should match len(SUFFIX)
	LEN_BODY          = 2
	LEN_LEN_IV        = 1
	LEN_IV            = 12
	LEN_TAG           = 16
	LEN_ACTUAL_HEADER = LEN_BODY + LEN_LEN_IV + LEN_IV
	LEN_HEADER        = LEN_PREFIX /* 00 prefix */ + LEN_VERSION + LEN_ACTUAL_HEADER /* actual header */ + LEN_SUFFIX /* 0 suffix */
)

type RtpEncryptedPayload struct {
	Version       uint8
	LenCiphertext uint32
	LenIV         uint8
	IV            []byte
	Ciphertext    []byte
	Tag           []byte
}

func NewRtpEncryptedPayload(version uint8, IV []byte, CiphertextWithTag []byte) *RtpEncryptedPayload {
	// TODO: assert it has at least length of tag or more
	Ciphertext := CiphertextWithTag[:len(CiphertextWithTag)-LEN_TAG]
	Tag := CiphertextWithTag[len(CiphertextWithTag)-LEN_TAG:]
	return &RtpEncryptedPayload{
		Version:       version,
		LenCiphertext: uint32(len(Ciphertext)),
		LenIV:         LEN_LEN_IV,
		IV:            IV,
		Ciphertext:    Ciphertext,
		Tag:           Tag,
	}
}

func (encryptedPayload *RtpEncryptedPayload) UnmarshalHeader(header []byte) error {
	if header == nil {
		return ErrNoData
	}

	if len(header) < LEN_ACTUAL_HEADER {
		return fmt.Errorf("payload does not have required header size")
	}

	header = header[:LEN_ACTUAL_HEADER]
	lenCiphertext := uint32(binary.BigEndian.Uint16(header[0:2]))
	lenIV := uint8(header[2])
	IV := header[3:]

	if lenIV != LEN_IV {
		return fmt.Errorf("header does not have IV size of %v received %v instead", LEN_IV, encryptedPayload.LenIV)
	}

	encryptedPayload.LenCiphertext = lenCiphertext
	encryptedPayload.LenIV = lenIV
	encryptedPayload.IV = IV

	return nil
}

func (encryptedPayload *RtpEncryptedPayload) Unmarshal(payload []byte) error {
	if payload == nil {
		return ErrNoData
	}

	if len(payload) < LEN_HEADER {
		return fmt.Errorf("payload does not have required header size")
	}
	header := payload[LEN_PREFIX+LEN_VERSION : LEN_PREFIX+LEN_VERSION+LEN_ACTUAL_HEADER]

	version := header[0]
	if !(version != VERSION_0 || version != VERSION_1) {
		return fmt.Errorf("header failed version check version = %v", version)
	}

	// Unmarshal the header, skip prefix and suffix
	err := encryptedPayload.UnmarshalHeader(header)
	if err != nil {
		return err
	}

	lenCiphertext := int(encryptedPayload.LenCiphertext)
	if len(payload) < LEN_HEADER+lenCiphertext+LEN_TAG {
		return fmt.Errorf("payload does not have required size")
	}
	ciphertextWithTag := payload[LEN_HEADER : LEN_HEADER+lenCiphertext+LEN_TAG]
	ciphertext := ciphertextWithTag[0:lenCiphertext]
	tag := ciphertextWithTag[lenCiphertext : lenCiphertext+LEN_TAG]

	if len(payload) > LEN_HEADER+lenCiphertext+LEN_TAG {
		additionalData := payload[LEN_HEADER+lenCiphertext+LEN_TAG:]
		log.Printf("found additional data = %v", hex.EncodeToString(additionalData))
	}

	encryptedPayload.Version = uint8(version)
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
	 [ PREFIX ] [ VERSION_0 ] [ length body ] [ length iv ] [ iv ] [ SUFFIX ] [ ciphertext ] [ tag ]
	*/
	// TODO: use VERSION_1 for screenshare
	header := append([]byte{}, PREFIX...)
	header = append(header, VERSION_0)
	header = append(header, lenBody...)
	header = append(header, lenIV...)
	header = append(header, encryptedPayload.IV...)
	header = append(header, SUFFIX...)

	encodedPayload := append(header, encryptedPayload.Ciphertext...)
	encodedPayload = append(encodedPayload, encryptedPayload.Tag...)
	return encodedPayload
}
