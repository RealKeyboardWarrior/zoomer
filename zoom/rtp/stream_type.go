package rtp

type StreamType byte

const (
	STREAM_TYPE_VIDEO       StreamType = 0x00
	STREAM_TYPE_AUDIO       StreamType = 0x01
	STREAM_TYPE_SCREENSHARE StreamType = 0x02
)
