package protocol

import "errors"

var (
	ErrNoData                  = errors.New("no data was passed")
	ErrInvalidLength           = errors.New("invalid length")
	ErrNonEmptyBufferStartBit  = errors.New("buffer was not empty while start bit was received")
	ErrEmptyBufferContinuation = errors.New("buffer was empty while continuation was received")
)
