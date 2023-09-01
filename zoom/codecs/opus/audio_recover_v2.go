package opus

import (
	"io"
	"os"
	"time"

	"github.com/pion/webrtc/v3/pkg/media"
)

type PCMRecorder struct {
	rawFile io.WriteCloser
}

func CreateNewPCMRecorder() (*PCMRecorder, error) {
	fileName := time.Now().Format("2006-01-02-15-04-05") + ".raw"
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return &PCMRecorder{
		rawFile: f,
	}, nil
}

func (recorder *PCMRecorder) Record(sample *media.Sample) error {
	_, err := recorder.rawFile.Write(sample.Data)
	return err
}

func (recorder *PCMRecorder) Close() error {
	return recorder.rawFile.Close()
}
