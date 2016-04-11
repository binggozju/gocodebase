/*
transport defines and implements message oriented communication channel
to complete various transactions (e.g., an RPC).
*/

package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// recvMsg表示从传输层接收到的消息. 所有的传输协议信息已被移除.
type recvMsg struct {
	data []byte
	// nil: received some data
	// io.EOF: stream is completed. data is nil.
	// other non-nil error: transport failure. data is nil.
	err error
}

func (recvMsg) isItem() bool {
	return true
}

// All items in an out of a recvBuffer should be the same type.
type item interface {
	isItem() bool
}

// recvBuffer is an unbounded channel of item.
type recvBuffer struct {
	c       chan item
	mu      sync.Mutex
	backlog []item
}

func newRecvBuffer() *recvBuffer {
	b := &recvBuffer{
		c: make(chan item, 1),
	}
	return b
}

// 向接收缓存中添加一条消息, 并将缓存中的第一条消息写入channel中等待被读取
func (b *recvBuffer) put(r item) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.backlog = append(b.backlog, r)
	select {
	case b.c <- b.backlog[0]:
		b.backlog = b.backlog[1:]
	default:
	}
}

func (b *recvBuffer) load() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.backlog) > 0 {
		select {
		case b.c <- b.backlog[0]:
			b.backlog = b.backlog[1:]
		default:
		}
	}
}

// 从接收缓存中读取一条消息（即读取channel中存储的消息）
func (b *recvBuffer) get() <-chan item {
	return b.c
}

// recvBufferReader实现io.Reader接口，从接收缓存recvBuffer中读取数据
type recvBufferReader struct {
	recv *recvBuffer
	last *bytes.Reader // Stores the remaining data in the previous calls.
	err  error
}

// Read reads the next len(p) bytes from last. If last is drained, it tries to
// read additional data from recv. It blocks if there no additional data available
// in recv. If Read returns any non-nil error, it will continue to return that error.
func (r *recvBufferReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	defer func() { r.err = err }()

	if r.last != nil && r.last.Len() > 0 {
		// Read remaining data left in last call.
		return r.last.Read(p)
	}
	select {
	case <-r.ctx.Done():
		return 0, ContextErr(r.ctx.Err())
	case i := <-r.recv.get():
		r.recv.load()
		m := i.(*recvMsg)
		if m.err != nil {
			return 0, m.err
		}
		r.last = bytes.NewReader(m.data)
		return r.last.Read(p)
	}
}
