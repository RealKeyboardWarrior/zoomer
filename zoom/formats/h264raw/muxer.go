package h264raw

import (
	"io"
	"os"
	"time"
)

func Recorder() (io.WriteCloser, error) {
	f, err := os.Create(time.Now().Format("2006-01-02-15-04-05") + ".h264")
	if err != nil {
		return nil, err
	}
	return f, nil
	// defer f.Close()
	// n2, err := f.Write(d2)
}
