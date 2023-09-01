package ext

type RtpMetadata struct {
	StreamId                    []byte
	ScreenShareResolution       *RtpExtResolution
	ScreenShareFrameInfo        *RtpExtFrameInfo
	AudioHeaderEncryptedPayload []byte
}
