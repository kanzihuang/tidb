package parser

import (
	goio "io"
)

const (
	defaultBufSize = 0x1000
	maxBufSize     = 0x100000
)

type buffer struct {
	reader goio.Reader
	buf    []byte
	size   int
}

type bufferOpt func(buf *buffer)

func withBufferSize(size int) bufferOpt {
	return func(buf *buffer) {
		if size <= len(buf.buf) {
			return
		}
		if size > maxBufSize {
			size = maxBufSize
		}
		tmp := make([]byte, size)
		copy(tmp, buf.buf)
		buf.buf = tmp
	}
}

func NewBuffer(reader goio.Reader, opts ...bufferOpt) *buffer {
	buf := &buffer{
		reader: reader,
	}
	for _, opt := range opts {
		opt(buf)
	}
	if len(buf.buf) == 0 {
		withBufferSize(defaultBufSize)(buf)
	}
	return buf
}

func (buf *buffer) Reset(offset int) error {
	if offset < 0 || offset > buf.size {
		return ErrInvalidBufferOffset
	}
	if offset == 0 {
		return nil
	}
	pos := buf.size - offset
	if pos > 0 {
		copy(buf.buf, buf.buf[offset:buf.size])
	}
	buf.size = pos
	return nil
}

func (buf *buffer) ReadString() (string, error) {
	if buf.size < 0 || buf.size > len(buf.buf) {
		return "", ErrInvalidBufferSize
	}
	var err error
	for buf.size < len(buf.buf) {
		var n int
		n, err = buf.reader.Read(buf.buf[buf.size:])
		buf.size += n
		if err == goio.EOF {
			break
		} else if err != nil {
			return "", err
		}
	}
	return string(buf.buf[:buf.size]), err
}
