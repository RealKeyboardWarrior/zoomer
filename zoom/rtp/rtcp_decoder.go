package rtp

import (
	"fmt"
	"log"

	"github.com/pion/rtcp"
)

func RtcpProcess(rawPkt []byte) ([]rtcp.Packet, error) {
	rtcpPackets, err := rtcp.Unmarshal(rawPkt)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	for _, rtcpPacket := range rtcpPackets {
		if stringer, canString := rtcpPacket.(fmt.Stringer); canString {
			fmt.Printf("rtcp : %v", stringer.String())
		}
	}
	return rtcpPackets, nil
}
