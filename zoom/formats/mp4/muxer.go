package mp4

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/RealKeyboardWarrior/joy4/av"
	"github.com/RealKeyboardWarrior/joy4/av/avutil"
	"github.com/RealKeyboardWarrior/joy4/codec/h264parser"
	"github.com/RealKeyboardWarrior/joy4/format"
)

func init() {
	format.RegisterAll()
}

type Mp4Recorder struct {
	muxer        av.MuxCloser
	initialized  bool
	prevDuration time.Duration
}

func NewRecorder() (*Mp4Recorder, error) {
	muxer, err := avutil.Create(time.Now().Format("2006-01-02-15-04-05") + ".fmp4")
	if err != nil {
		return nil, err
	}

	log.Printf("Mp4Recorder: starting")
	return &Mp4Recorder{
		muxer:        muxer,
		initialized:  false,
		prevDuration: time.Duration(0),
	}, nil
}

func (recorder *Mp4Recorder) tryInitialize(nalus []byte) error {
	codec, err := h264parser.PktToCodecData(av.Packet{
		IsKeyFrame: true,
		Data:       nalus,
	})
	if err != nil {
		return err
	}
	if codec == nil {
		return fmt.Errorf("failed to initialize muxer because no codec data was extracted from nalus")
	}

	codecs := []av.CodecData{codec}
	if err = recorder.muxer.WriteHeader(codecs); err != nil {
		return err
	}

	recorder.initialized = true
	return nil
}

func IsKeyFrame(naluz []byte) bool {
	nalus, _ := h264parser.SplitNALUs(naluz)
	log.Printf("Mp4Recorder: processing %v nalus", len(nalus))
	for _, nalu := range nalus {
		if len(nalu) > 0 {
			naltype := nalu[0] & 0x1f
			if naltype == 5 {
				return true
			}
		}
	}
	return false
}

func (recorder *Mp4Recorder) WritePacket(nalus []byte, duration time.Duration) error {
	if !recorder.initialized {
		if err := recorder.tryInitialize(nalus); err != nil {
			log.Printf("Mp4Recorder: failed to initialize with packet")
			return nil
		} else {
			log.Printf("Mp4Recorder: initialized")
		}
	}

	log.Printf("Mp4Recorder: writing packet")
	isKeyFrame := IsKeyFrame(nalus)
	avccBytes, err := h264parser.AnnexBToAVCC(nalus)
	if err != nil {
		return err
	}

	if avccBytes == nil {
		log.Printf("skipping writing packet, conversion to AVCC yielded no VCL units in %v", hex.EncodeToString(nalus))
		return nil
	}

	recorder.prevDuration = recorder.prevDuration + duration
	err = recorder.muxer.WritePacket(av.Packet{
		IsKeyFrame:      isKeyFrame,
		Idx:             int8(0),
		CompositionTime: time.Duration(0),
		Time:            recorder.prevDuration,
		Data:            avccBytes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (recorder *Mp4Recorder) Close() error {
	if !recorder.initialized {
		log.Printf("warn: Mp4Recorder was never initialized")
		return nil
	}

	log.Printf("Mp4Recorder: closing")
	if err := recorder.muxer.WriteTrailer(); err != nil {
		return err
	}

	return recorder.muxer.Close()
}
