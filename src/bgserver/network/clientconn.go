package network

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// 定义一些常用的错误
var (
	ErrUnspecTarget			= errors.New("target is unspecified")
	ErrClientConnClosing	= errors.New("the client connection is closing")
	ErrClientConnTimeout	= errors.New("timed out trying to connect")
)

var (
	ConnectTimeout = 20 * time.Second
)

// 包含所有与连接server相关的选项
type ConnectOptions struct {
	Dialer func(string, time.Duration) (net.Conn, error)
	Timeout time.Duration
}

// client发起连接时可指定的选项
type dialOptions struct {
	codec    Codec			// 编码解码
	cp       Compressor		// 压缩
	dc       Decompressor	// 解压缩
	copts    ConnectOptions // 用于连接相关的设置，比如超时、鉴权、拨号函数选择等
}

// 用于设置dialOptions中的字段
type DialOption func(*dialOptions)

// 设置用于编码和解码的codec
func WithCodec(c Codec) DialOption {
	return func(o *dialOptions) {
		o.codec = c
	}
}

func WithCompressor(cp Compressor) DialOption {
	return func(o *dialOptions) {
		o.cp = cp
	}
}

func WithDecompressor(dc Decompressor) DialOption {
	return func(o *dialOptions) {
		o.dc = dc
	}
}

// WithTimeout returns a DialOption that configures a timeout for dialing a client connection.
func WithTimeout(d time.Duration) DialOption {
	return func(o *dialOptions) {
		o.copts.Timeout = d
	}
}

// WithDialer returns a DialOption that specifies a function to use for dialing network addresses.
func WithDialer(f func(addr string, timeout time.Duration) (net.Conn, error)) DialOption {
	return func(o *dialOptions) {
		o.copts.Dialer = f
	}
}
// Dial creates a client connection the given target.
func Dial(target string, opts ...DialOption) (*ClientConn, error) {
	cc := &ClientConn{	// 创建一个ClientConn对象
		target: target,
	}
	for _, opt := range opts { // 设置ClientConn对象的拨号选项
		opt(&cc.dopts)
	}
	if cc.dopts.codec == nil { // 使用proto作为默认的编码解码器
		// Set the default codec.
		cc.dopts.codec = protoCodec{}
	}
	cc.conn, err = NewConn()
	if err != nil {
		return
	}

	colonPos := strings.LastIndex(target, ":")
	if colonPos == -1 {
		colonPos = len(target)
	}
	cc.authority = target[:colonPos]
	return cc, nil
}

// TCP连接的状态类型
type ConnectivityState int

// 连接状态的所有枚举值
const (
	// Idle indicates the ClientConn is idle.
	Idle ConnectivityState = iota
	// Connecting indicates the ClienConn is connecting.
	Connecting
	// Ready indicates the ClientConn is ready for work.
	Ready
	// TransientFailure indicates the ClientConn has seen a failure but expects to recover.
	TransientFailure
	// Shutdown indicates the ClientConn has started shutting down.
	Shutdown
)

func (s ConnectivityState) String() string {
	switch s {
	case Idle:
		return "IDLE"
	case Connecting:
		return "CONNECTING"
	case Ready:
		return "READY"
	case TransientFailure:
		return "TRANSIENT_FAILURE"
	case Shutdown:
		return "SHUTDOWN"
	default:
		panic(fmt.Sprintf("unknown connectivity state: %d", s))
	}
}

type ClientConn struct {
	target		string
	conn		Conn
	dopts		dialOptions
}

// TODO

type Conn struct {
	// TODO
}



