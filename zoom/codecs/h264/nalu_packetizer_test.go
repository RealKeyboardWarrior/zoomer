package h264

import (
	"bytes"
	"log"
	"math/rand"
	"testing"
)

func TestNaluPacketizer(t *testing.T) {

	// Generate a random payload
	payload := make([]byte, 1800)
	rand.Read(payload)

	packetizer := NewNaluPacketizer()

	// Packetize the payload
	pkts, err := packetizer.Marshal(payload)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Depacketize the payload
	var decoded []byte
	for _, pkt := range pkts {
		decoded, err = packetizer.Unmarshal(pkt)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if !bytes.Equal(payload, decoded) {
		t.Errorf("Decoded payload do not match expected payload")
	}
}
