package ext

const (
	RTP_EXTENSION_ID_UUID                   = 7
	RTP_EXTENSION_ID_AUDIO_IV               = 6 // warning: extension id 6 is re-used
	RTP_EXTENSION_ID_SCREENSHARE_RESOLUTION = 6 // warning: extension id 6 is re-used
	RTP_EXTENSION_ID_SCREENSHARE_FRAME_INFO = 4 // Screen share
	RTP_EXTENSION_ID_VIDEO_UNKNOWN_1        = 1 // always 4ff770 or 400000
	RTP_EXTENSION_ID_VIDEO_FRAME_INFO       = 3 // Some counters that increase
	RTP_EXTENSION_ID_VIDEO_UNKNOWN_5        = 5 // always 00
	RTP_EXTENSION_ID_VIDEO_UNKNOWN_7        = 7 // always 00
	RTP_EXTENSION_UNKNOWN                   = 1
)
