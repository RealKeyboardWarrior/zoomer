package rtp

import (
	"log"

	"github.com/pion/rtcp"
)

func RtcpProcess(rawPkt []byte) ([]rtcp.Packet, error) {
	p, err := rtcp.Unmarshal(rawPkt)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// log.Printf("%v", p[0])
	return p, nil
}
