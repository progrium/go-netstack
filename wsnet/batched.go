package wsnet

import (
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// BatchedOptions configures batched write and stream read behavior (used on both ends).
type BatchedOptions struct {
	// FlushInterval is the maximum time to hold data before sending a WebSocket
	// message. A single frame is sent within this long even if the batch never fills.
	// Zero uses DefaultFlushInterval (2ms).
	FlushInterval time.Duration
	// MaxBatchBytes is the maximum buffered bytes before flushing. Larger values
	// improve throughput under load; smaller values reduce latency.
	// Zero uses DefaultMaxBatchBytes (32 * 1500, i.e. ~32 MTU-sized packets).
	MaxBatchBytes int
}

const (
	DefaultFlushInterval = 2 * time.Millisecond
	DefaultMaxBatchBytes = 32 * 1500 // 32 * typical MTU
)

// BatchedConn wraps a WebSocket connection for use on both server and client:
// - Write: buffered; flushes when buffer reaches MaxBatchBytes or after FlushInterval (single frames still go out promptly).
// - Read: stream; each WebSocket message may contain many bytes; Read() drains the current message and fetches the next when empty.
// Use the same wrapper on both ends so both sides batch sends and consume batched receives as a stream.
func BatchedConn(conn *websocket.Conn, opts BatchedOptions) net.Conn {
	if opts.FlushInterval == 0 {
		opts.FlushInterval = DefaultFlushInterval
	}
	if opts.MaxBatchBytes == 0 {
		opts.MaxBatchBytes = DefaultMaxBatchBytes
	}
	c := &batchedConn{
		conn: conn,
		opts: opts,
		buf:  make([]byte, 0, opts.MaxBatchBytes*2), // avoid realloc for typical batch
	}
	return c
}

type batchedConn struct {
	conn    *websocket.Conn
	opts    BatchedOptions
	mu      sync.Mutex
	buf     []byte
	timer   *time.Timer
	closed  bool
	readMu  sync.Mutex
	readBuf []byte
	readOff int
}

func (c *batchedConn) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, net.ErrClosed
	}
	c.buf = append(c.buf, p...)
	n = len(p)
	if len(c.buf) >= c.opts.MaxBatchBytes {
		return n, c.flushLocked()
	}
	c.armTimerLocked()
	return n, nil
}

func (c *batchedConn) armTimerLocked() {
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.timer = time.AfterFunc(c.opts.FlushInterval, c.onFlush)
}

func (c *batchedConn) onFlush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timer = nil
	if len(c.buf) > 0 {
		_ = c.flushLocked()
	}
}

func (c *batchedConn) flushLocked() error {
	if len(c.buf) == 0 {
		return nil
	}
	err := c.conn.WriteMessage(websocket.BinaryMessage, c.buf)
	c.buf = c.buf[:0]
	return err
}

func (c *batchedConn) Read(p []byte) (n int, err error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()
	for len(p) > 0 {
		if c.readOff >= len(c.readBuf) {
			_, c.readBuf, err = c.conn.ReadMessage()
			if err != nil {
				return n, err
			}
			c.readOff = 0
		}
		nc := copy(p, c.readBuf[c.readOff:])
		c.readOff += nc
		n += nc
		p = p[nc:]
	}
	return n, nil
}

func (c *batchedConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	_ = c.flushLocked()
	return c.conn.Close()
}

func (c *batchedConn) LocalAddr() net.Addr  { return c.conn.LocalAddr() }
func (c *batchedConn) RemoteAddr() net.Addr { return c.conn.RemoteAddr() }
func (c *batchedConn) SetDeadline(t time.Time) error {
	_ = c.conn.SetReadDeadline(t)
	_ = c.conn.SetWriteDeadline(t)
	return nil
}
func (c *batchedConn) SetReadDeadline(t time.Time) error  { return c.conn.SetReadDeadline(t) }
func (c *batchedConn) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }

var _ net.Conn = (*batchedConn)(nil)
