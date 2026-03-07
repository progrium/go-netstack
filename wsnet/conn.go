package wsnet

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// Conn adapts *websocket.Conn to net.Conn with one-to-one message semantics:
// each Read reads one binary message; each Write sends one binary message.
// Use this when you do not need batching (e.g. for comparison or low traffic).
func Conn(conn *websocket.Conn) net.Conn {
	return &msgConn{conn: conn}
}

type msgConn struct {
	conn *websocket.Conn
}

func (c *msgConn) Read(b []byte) (n int, err error) {
	_, p, err := c.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(b, p), nil
}

func (c *msgConn) Write(b []byte) (n int, err error) {
	err = c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *msgConn) Close() error {
	return c.conn.Close()
}

func (c *msgConn) LocalAddr() net.Addr  { return c.conn.LocalAddr() }
func (c *msgConn) RemoteAddr() net.Addr { return c.conn.RemoteAddr() }
func (c *msgConn) SetDeadline(t time.Time) error {
	_ = c.conn.SetReadDeadline(t)
	_ = c.conn.SetWriteDeadline(t)
	return nil
}
func (c *msgConn) SetReadDeadline(t time.Time) error  { return c.conn.SetReadDeadline(t) }
func (c *msgConn) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }

var _ net.Conn = (*msgConn)(nil)
