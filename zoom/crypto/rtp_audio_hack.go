package crypto

func RewriteAudioHack(header []byte, payload []byte) []byte {
	// This is a dirty hack that rewrites audio rtp packets payloads
	// it rewrite the actual payload and adds:
	// prefix, version, header (len ciphertext, len iv, iv)
	reformattedPayload := PREFIX
	reformattedPayload = append(reformattedPayload, VERSION_0)
	reformattedPayload = append(reformattedPayload, header...)
	reformattedPayload = append(reformattedPayload, payload...)
	return reformattedPayload
}
