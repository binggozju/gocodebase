package message

import (
    "bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"time"
)


type payloadFormat uint

const (
    compressionNone payloadFormat = iota // no compression
	compressionMade
)

// Compressor defines the interface used to compress a message.
type Compressor interface {
	Do(w io.Writer, p []byte) error		// Do compresses p into w.
	Type() string		// Type returns the compression algorithm the Compressor uses.
}

// NewGZIPCompressor creates a Compressor based on GZIP.
func NewGZIPCompressor() Compressor {
	return &gzipCompressor{}
}

type gzipCompressor struct {
}

func (c *gzipCompressor) Do(w io.Writer, p []byte) error {
	z := gzip.NewWriter(w)
	if _, err := z.Write(p); err != nil {
		return err
	}
	return z.Close()
}

func (c *gzipCompressor) Type() string {
	return "gzip"
}

// Decompressor defines the interface used to decompress a message.
type Decompressor interface {
	// Do reads the data from r and uncompress them.
	Do(r io.Reader) ([]byte, error)
	// Type returns the compression algorithm the Decompressor uses.
	Type() string
}

type gzipDecompressor struct {
}

// NewGZIPDecompressor creates a Decompressor based on GZIP.
func NewGZIPDecompressor() Decompressor {
	return &gzipDecompressor{}
}

func (d *gzipDecompressor) Do(r io.Reader) ([]byte, error) {
	z, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	return ioutil.ReadAll(z)
}

func (d *gzipDecompressor) Type() string {
	return "gzip"
}
