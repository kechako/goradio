package rtlfm

import (
	"encoding/binary"
	"fmt"
	"io"
)

type FrameReader interface {
	Read(frame []int16) error
}

type frameReader struct {
	r   io.Reader
	buf []byte
}

func NewFrameReader(r io.Reader) FrameReader {
	return &frameReader{
		r: r,
	}
}

func (r *frameReader) Read(frame []int16) error {
	if len(frame) == 0 {
		return nil
	}

	size := 2 * len(frame)
	if r.buf == nil || len(r.buf) != size {
		r.buf = make([]byte, size)
	}

	err := binary.Read(r.r, binary.LittleEndian, frame)
	if err != nil {
		return fmt.Errorf("failed to read audio frame: %w", err)
	}

	return nil
}
