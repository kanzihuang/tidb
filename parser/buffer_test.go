package parser

import (
	testify_require "github.com/stretchr/testify/require"
	goio "io"
	"strings"
	"testing"
)

func Test_Buffer_ReadString(t *testing.T) {
	tests := []struct {
		name     string
		reader   goio.Reader
		buf      []byte
		size     int
		offset   int
		want     string
		errReset error
		errRead  error
	}{
		{
			name:   "first reading",
			reader: strings.NewReader("01234567890123456789"),
			buf:    []byte("abcdefghij"),
			want:   "0123456789",
		},
		{
			name:   "repeated reading",
			reader: strings.NewReader("01234567890123456789"),
			buf:    []byte("abcdefghij"),
			size:   10,
			want:   "abcdefghij",
		},
		{
			name:   "with offset less than size",
			reader: strings.NewReader("01234567890123456789"),
			buf:    []byte("abcdefghij"),
			size:   8,
			offset: 5,
			want:   "fgh0123456",
		},
		{
			name:   "with offset equals size",
			reader: strings.NewReader("01234567890123456789"),
			buf:    []byte("abcdefghij"),
			size:   8,
			offset: 8,
			want:   "0123456789",
		},
		{
			name:     "with offset great than size",
			reader:   strings.NewReader("01234567890123456789"),
			buf:      []byte("abcdefghij"),
			size:     10,
			offset:   11,
			errReset: ErrInvalidBufferOffset,
		},
		{
			name:    "with some data and EOF",
			reader:  strings.NewReader("012345"),
			buf:     []byte("abcdefghij"),
			size:    10,
			offset:  10,
			want:    "012345",
			errRead: goio.EOF,
		},
		{
			name:    "with empty and EOF",
			reader:  strings.NewReader(""),
			buf:     []byte("abcdefghij"),
			size:    10,
			offset:  10,
			errRead: goio.EOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := buffer{
				reader: tt.reader,
				buf:    tt.buf,
				size:   tt.size,
			}
			err := buf.Reset(tt.offset)
			testify_require.Equal(t, tt.errReset, err)
			if err != nil {
				return
			}
			testify_require.NoError(t, err)
			got, err := buf.ReadString()
			testify_require.Equal(t, tt.errRead, err)
			if err != nil && err != goio.EOF {
				return
			}
			testify_require.Equal(t, tt.want, got)
		})
	}
}

func Test_withBufferSize(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		size int
		want []byte
	}{
		{
			name: "with size less than buf.size",
			buf:  []byte("abcdefghij"),
			size: 5,
			want: []byte("abcdefghij"),
		},
		{
			name: "with size equals buf.size",
			buf:  []byte("abcdefghij"),
			size: 10,
			want: []byte("abcdefghij"),
		},
		{
			name: "with size great than buf.size",
			buf:  []byte("abcdefghij"),
			size: 11,
			want: []byte("abcdefghij\000"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &buffer{
				buf: tt.buf,
			}
			withBufferSize(tt.size)(buf)
			testify_require.Equal(t, tt.want, buf.buf)
		})
	}
}
