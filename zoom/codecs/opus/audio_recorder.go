package opus

import (
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

type OggRecorder struct {
	oggFile *oggwriter.OggWriter
}

func CreateNewOggRecorder() (*OggRecorder, error) {
	fileName := time.Now().Format("2006-01-02-15-04-05") + ".ogg"
	oggFile, err := oggwriter.New(fileName, 16000, 2)
	if err != nil {
		return nil, err
	}
	return &OggRecorder{
		oggFile: oggFile,
	}, nil
}

func (recorder *OggRecorder) Record(sample *media.Sample) error {
	rtpPacket := &rtp.Packet{
		Header: rtp.Header{
			Timestamp: sample.PacketTimestamp,
		},
		Payload: sample.Data,
	}

	err := recorder.oggFile.WriteRTP(rtpPacket)
	if err != nil {
		return err
	}
	return nil
}

func (recorder *OggRecorder) Close() error {
	return recorder.oggFile.Close()
}
