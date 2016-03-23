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

	. "bgserver/common"
)


// parser reads complelete messages from the underlying reader.
type parser struct {
	r io.Reader		// r is the underlying reader.
	header [5]byte	// The header of a message.
	// 其中第一个字节用于表示消息体是否被压缩了，后面四个字节标记消息体的长度
}

// recvMsg reads a complete message from the stream.
//
// It returns the message and its payload (compression/encoding)
// format. The caller owns the returned msg memory.
//
// If there is an error, possible values are:
//   * io.EOF, when no messages remain
//   * io.ErrUnexpectedEOF
//   * of type transport.ConnectionError
//   * of type transport.StreamError
// No other error values or types must be returned, which also means
// that the underlying io.Reader must not return an incompatible
// error.
func (p *parser) recvMsg() (pf payloadFormat, msg []byte, err error) {
	if _, err := io.ReadFull(p.r, p.header[:]); err != nil {
		return 0, nil, err
	}

	pf = payloadFormat(p.header[0])
	length := binary.BigEndian.Uint32(p.header[1:])

	if length == 0 {
		return pf, nil, nil
	}
	// TODO: garbage. reuse buffer after proto decoding instead of making it for each message:
	msg = make([]byte, int(length))
	if _, err := io.ReadFull(p.r, msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	return pf, msg, nil
}

func (p *parser) SendMsg(pf payloadFormat, msg []byte) (err error) {
	// TODO
}

// encode serializes msg and prepends the message header. If msg is nil, it
// generates the message header of 0 message length.
// 对需要发送的消息进行编码
func encode(c Codec, msg interface{}, cp Compressor, cbuf *bytes.Buffer) ([]byte, error) {
	var b []byte
	var length uint
	if msg != nil {
		var err error
		// TODO(zhaoq): optimize to reduce memory alloc and copying.
		b, err = c.Marshal(msg)
		if err != nil {
			return nil, err
		}
		if cp != nil {
			if err := cp.Do(cbuf, b); err != nil {
				return nil, err
			}
			b = cbuf.Bytes()
		}
		length = uint(len(b))
	}
	if length > math.MaxUint32 {
		return nil, Errorf(codes.InvalidArgument, "grpc: message too large (%d bytes)", length)
	}

	const (
		payloadLen = 1
		sizeLen    = 4
	)

	var buf = make([]byte, payloadLen+sizeLen+len(b))

	// Write payload format
	if cp == nil {
		buf[0] = byte(compressionNone)
	} else {
		buf[0] = byte(compressionMade)
	}
	// Write length of b into buf
	binary.BigEndian.PutUint32(buf[1:], uint32(length))
	// Copy encoded msg to buf
	copy(buf[5:], b)

	return buf, nil
}

func checkRecvPayload(pf payloadFormat, recvCompress string, dc Decompressor) error {
	switch pf {
	case compressionNone:
	case compressionMade:
		if recvCompress == "" {
			return transport.StreamErrorf(codes.InvalidArgument, "grpc: invalid grpc-encoding %q with compression enabled", recvCompress)
		}
		if dc == nil || recvCompress != dc.Type() {
			return transport.StreamErrorf(codes.InvalidArgument, "grpc: Decompressor is not installed for grpc-encoding %q", recvCompress)
		}
	default:
		return transport.StreamErrorf(codes.InvalidArgument, "grpc: received unexpected payload format %d", pf)
	}
	return nil
}

// 从网络中接收消息，提取消息并解码
func recv(p *parser, c Codec, s *transport.Stream, dc Decompressor, m interface{}) error {
	pf, d, err := p.recvMsg()
	if err != nil {
		return err
	}
	if err := checkRecvPayload(pf, s.RecvCompress(), dc); err != nil {
		return err
	}
	if pf == compressionMade {
		d, err = dc.Do(bytes.NewReader(d))
		if err != nil {
			return transport.StreamErrorf(codes.Internal, "grpc: failed to decompress the received message %v", err)
		}
	}
	if err := c.Unmarshal(d, m); err != nil {
		return transport.StreamErrorf(codes.Internal, "grpc: failed to unmarshal the received message %v", err)
	}
	return nil
}

