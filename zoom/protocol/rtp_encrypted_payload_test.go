package protocol

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func FuzzUnmarshalRtpEncryptedPayload(f *testing.F) {
	var (
		invalidPayload  = []byte{0x00}
		validPayload, _ = hex.DecodeString("000000860c18000000000000000000000000916b4946439355dd851c3e8fcb0c9ca360e2fa595aaee6917c22177d66bd11863b6c830eed9a767e15193adba82a0e2295053ad6196daffd70377ca6b9be9301472d2ae9ea0320fc69d73d4a0e486db9665945bb5c09fd8db175a01b6bc000bec59a5bfdfece8a4363a40ef06092061bea0762e8ed19b5118837d599bf503f928c621500c4c6c81b9e57f31b768124eecb85922d0331")
	)
	f.Add(invalidPayload)
	f.Add(validPayload)
	f.Fuzz(func(t *testing.T, encryptedPayload []byte) {
		zoomEncryptedPayload := &RtpEncryptedPayload{}
		zoomEncryptedPayload.Unmarshal(encryptedPayload)
	})
}

func TestUnmarshalRtpEncryptedPayload(t *testing.T) {
	var (
		validPayload, _       = hex.DecodeString("000000860c18000000000000000000000000916b4946439355dd851c3e8fcb0c9ca360e2fa595aaee6917c22177d66bd11863b6c830eed9a767e15193adba82a0e2295053ad6196daffd70377ca6b9be9301472d2ae9ea0320fc69d73d4a0e486db9665945bb5c09fd8db175a01b6bc000bec59a5bfdfece8a4363a40ef06092061bea0762e8ed19b5118837d599bf503f928c621500c4c6c81b9e57f31b768124eecb85922d0331")
		expectedIv, _         = hex.DecodeString("180000000000000000000000")
		expectedCiphertext, _ = hex.DecodeString("916b4946439355dd851c3e8fcb0c9ca360e2fa595aaee6917c22177d66bd11863b6c830eed9a767e15193adba82a0e2295053ad6196daffd70377ca6b9be9301472d2ae9ea0320fc69d73d4a0e486db9665945bb5c09fd8db175a01b6bc000bec59a5bfdfece8a4363a40ef06092061bea0762e8ed19b5118837d599bf503f928c621500c4c6")
		expectedTag, _        = hex.DecodeString("c81b9e57f31b768124eecb85922d0331")
	)

	zoomEncryptedPayload := &RtpEncryptedPayload{}
	zoomEncryptedPayload.Unmarshal(validPayload)

	if !bytes.Equal(zoomEncryptedPayload.IV, expectedIv) {
		t.Error("IV did not match expected")
	}

	if !bytes.Equal(zoomEncryptedPayload.Ciphertext, expectedCiphertext) {
		t.Error("Tag did not match expected")
	}

	if !bytes.Equal(zoomEncryptedPayload.Tag, expectedTag) {
		t.Error("Tag did not match expected")
	}
}

func TestMarshalRtpEncryptedPayload(t *testing.T) {
	var (
		validPayload, _       = hex.DecodeString("000000860c18000000000000000000000000916b4946439355dd851c3e8fcb0c9ca360e2fa595aaee6917c22177d66bd11863b6c830eed9a767e15193adba82a0e2295053ad6196daffd70377ca6b9be9301472d2ae9ea0320fc69d73d4a0e486db9665945bb5c09fd8db175a01b6bc000bec59a5bfdfece8a4363a40ef06092061bea0762e8ed19b5118837d599bf503f928c621500c4c6c81b9e57f31b768124eecb85922d0331")
		expectedIv, _         = hex.DecodeString("180000000000000000000000")
		expectedCiphertext, _ = hex.DecodeString("916b4946439355dd851c3e8fcb0c9ca360e2fa595aaee6917c22177d66bd11863b6c830eed9a767e15193adba82a0e2295053ad6196daffd70377ca6b9be9301472d2ae9ea0320fc69d73d4a0e486db9665945bb5c09fd8db175a01b6bc000bec59a5bfdfece8a4363a40ef06092061bea0762e8ed19b5118837d599bf503f928c621500c4c6")
		expectedTag, _        = hex.DecodeString("c81b9e57f31b768124eecb85922d0331")
	)

	ciphertextWithTag := append(expectedCiphertext, expectedTag...)
	zoomEncryptedPayload := NewRtpEncryptedPayload(expectedIv, ciphertextWithTag)
	encryptedPayload := zoomEncryptedPayload.Marshal()

	if !bytes.Equal(encryptedPayload, validPayload) {
		t.Error("Plaintext did not match expected")
	}
}
