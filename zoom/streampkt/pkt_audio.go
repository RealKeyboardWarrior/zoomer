package streampkt

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type ZoomAudioPkt struct {
	lenRtp         uint16
	Rtp            []byte
	AdditionalData []byte
}

func (pkt *ZoomAudioPkt) Unmarshal(data []byte) error {
	// TODO check data not nil
	// TODO: check min header length
	if data[0] != 0x6B {
		return errors.New("ZoomAudioPkt expects 0x6B as starting byte")
	}

	rtpPktSize := binary.BigEndian.Uint16(data[21:23])
	rtpPkt := data[23 : 23+rtpPktSize]

	pkt.lenRtp = rtpPktSize
	pkt.Rtp = rtpPkt

	if len(data) > int(45+rtpPktSize) {
		pkt.AdditionalData = data[45+rtpPktSize:]
	}
	return nil
}

func (pkt *ZoomAudioPkt) String() string {
	return fmt.Sprintf("[ZoomAudioPkt lenRtp=%v rtp=%v additionalData=%v]", pkt.lenRtp, hex.EncodeToString(pkt.Rtp), hex.EncodeToString(pkt.AdditionalData))
}
