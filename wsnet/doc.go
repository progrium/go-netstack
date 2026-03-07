// Package wsnet provides net.Conn adapters for WebSocket that support batched
// framing. BatchedConn is used on both server and client: writes are buffered
// and flushed on a timer or size limit (single frames still go out promptly),
// and reads present incoming messages as a byte stream. Use Conn for simple
// one-message-per-read/write when batching is not needed.
package wsnet
