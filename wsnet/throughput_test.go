package wsnet

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/progrium/go-netstack/types"
	"github.com/progrium/go-netstack/vnet"
)

const (
	throughputPayloadSize = 1400
	throughputFrameSize   = 14 + throughputPayloadSize // ethernet header + payload
	transferPayloadMB     = 20
)

// buildQemuFrame writes a QEMU-framed ethernet frame to buf: [4 byte BE length][14 byte eth][payload].
// Returns total bytes (4 + throughputFrameSize). dstMAC and srcMAC are 6 bytes each.
func buildQemuFrame(buf []byte, dstMAC, srcMAC []byte) int {
	binary.BigEndian.PutUint32(buf[0:4], uint32(throughputFrameSize))
	copy(buf[4:10], dstMAC)
	copy(buf[10:16], srcMAC)
	buf[16] = 0x08 // IPv4
	buf[17] = 0x00
	return 4 + throughputFrameSize
}

// TestThroughputComparison runs native pipe, native WebSocket, vnet over WebSocket
// (unbatched), and vnet over WebSocket (batched with time-based flush), then
// reports MB/s and an overhead breakdown.
func TestThroughputComparison(t *testing.T) {
	targetPayload := int64(transferPayloadMB) * 1024 * 1024
	if testing.Short() {
		targetPayload = 5 * 1024 * 1024
	}
	t.Logf("transfer target: %.1f MB payload per path", float64(targetPayload)/(1024*1024))

	gatewayMAC := net.HardwareAddr{0x5a, 0x94, 0xef, 0xe4, 0x0c, 0xdd}
	clientMAC := net.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	frameBuf := make([]byte, 4+throughputFrameSize)
	totalBytes := buildQemuFrame(frameBuf, gatewayMAC, clientMAC)
	packetsToSend := (targetPayload + int64(throughputPayloadSize) - 1) / int64(throughputPayloadSize)

	// 1) Native pipe
	clientConn, serverConn := net.Pipe()
	go func() {
		buf := make([]byte, 4+throughputFrameSize)
		for {
			if _, err := io.ReadFull(serverConn, buf[:4]); err != nil {
				return
			}
			n := int(binary.BigEndian.Uint32(buf[:4]))
			if n > len(buf)-4 {
				n = len(buf) - 4
			}
			if _, err := io.ReadFull(serverConn, buf[4:4+n]); err != nil {
				return
			}
		}
	}()
	startNative := time.Now()
	for i := int64(0); i < packetsToSend; i++ {
		if _, err := clientConn.Write(frameBuf[:totalBytes]); err != nil {
			t.Fatal(err)
		}
	}
	_ = clientConn.Close()
	elapsedNative := time.Since(startNative)
	nativePayloadBytes := packetsToSend * int64(throughputPayloadSize)
	nativeMBps := float64(nativePayloadBytes) / (1024 * 1024) / elapsedNative.Seconds()
	t.Logf("native (pipe):     %.2f MB/s  (%.1f MB in %v)", nativeMBps, float64(nativePayloadBytes)/(1024*1024), elapsedNative.Round(time.Millisecond))

	// 2) Native WebSocket: same wire (WS binary msgs, one per packet), server reads and discards
	listenerWS, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listenerWS.Close()
	portWS := listenerWS.Addr().(*net.TCPAddr).Port
	upgraderWS := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srvWS := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgraderWS.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}),
	}
	go func() { _ = srvWS.Serve(listenerWS) }()
	defer srvWS.Shutdown(context.Background())
	time.Sleep(10 * time.Millisecond)

	wsNativeURL := "ws://127.0.0.1:" + fmt.Sprintf("%d", portWS) + "/"
	wsNativeClient, _, err := websocket.DefaultDialer.Dial(wsNativeURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	startNativeWS := time.Now()
	for i := int64(0); i < packetsToSend; i++ {
		if err := wsNativeClient.WriteMessage(websocket.BinaryMessage, frameBuf[:totalBytes]); err != nil {
			t.Fatal(err)
		}
	}
	_ = wsNativeClient.Close()
	elapsedNativeWS := time.Since(startNativeWS)
	nativeWSMBps := float64(nativePayloadBytes) / (1024 * 1024) / elapsedNativeWS.Seconds()
	t.Logf("native (websocket): %.2f MB/s  (%.1f MB in %v)", nativeWSMBps, float64(nativePayloadBytes)/(1024*1024), elapsedNativeWS.Round(time.Millisecond))

	// 3) Vnet over WebSocket (unbatched)
	config := &types.Configuration{
		Debug:             false,
		MTU:               1500,
		Subnet:            "192.168.127.0/24",
		GatewayIP:         "192.168.127.1",
		GatewayMacAddress: "5a:94:ef:e4:0c:dd",
		GatewayVirtualIPs: []string{"192.168.127.253"},
		Protocol:          types.QemuProtocol,
	}
	vn, err := vnet.New(config)
	if err != nil {
		t.Fatal(err)
	}
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Logf("upgrade: %v", err)
				return
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = vn.AcceptQemu(ctx, Conn(conn))
		}),
	}
	go func() { _ = srv.Serve(listener) }()
	defer srv.Shutdown(context.Background())
	time.Sleep(10 * time.Millisecond)

	wsURL := "ws://127.0.0.1:" + fmt.Sprintf("%d", port) + "/"
	wsConnClient, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer wsConnClient.Close()

	startVnet := time.Now()
	for i := int64(0); i < packetsToSend; i++ {
		if err := wsConnClient.WriteMessage(websocket.BinaryMessage, frameBuf[:totalBytes]); err != nil {
			t.Fatal(err)
		}
	}
	elapsedVnet := time.Since(startVnet)
	vnetMBps := float64(nativePayloadBytes) / (1024 * 1024) / elapsedVnet.Seconds()
	t.Logf("vnet (websocket):  %.2f MB/s  (%.1f MB in %v)", vnetMBps, float64(nativePayloadBytes)/(1024*1024), elapsedVnet.Round(time.Millisecond))

	// 4) Vnet over WebSocket (batched): both sides use BatchedConn (stream read + batched write)
	batchOpts := BatchedOptions{
		FlushInterval: 2 * time.Millisecond,
		MaxBatchBytes: 32 * totalBytes,
	}
	vn2, err := vnet.New(config)
	if err != nil {
		t.Fatal(err)
	}
	listenerBatch, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listenerBatch.Close()
	portBatch := listenerBatch.Addr().(*net.TCPAddr).Port
	upgraderBatch := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srvBatch := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgraderBatch.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = vn2.AcceptQemu(ctx, BatchedConn(conn, batchOpts))
		}),
	}
	go func() { _ = srvBatch.Serve(listenerBatch) }()
	defer srvBatch.Shutdown(context.Background())
	time.Sleep(10 * time.Millisecond)

	wsBatchURL := "ws://127.0.0.1:" + fmt.Sprintf("%d", portBatch) + "/"
	wsBatchClient, _, err := websocket.DefaultDialer.Dial(wsBatchURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	batchedClient := BatchedConn(wsBatchClient, batchOpts)

	startVnetBatch := time.Now()
	for i := int64(0); i < packetsToSend; i++ {
		if _, err := batchedClient.Write(frameBuf[:totalBytes]); err != nil {
			t.Fatal(err)
		}
	}
	if err := batchedClient.Close(); err != nil {
		t.Fatal(err)
	}
	elapsedVnetBatch := time.Since(startVnetBatch)
	vnetBatchMBps := float64(nativePayloadBytes) / (1024 * 1024) / elapsedVnetBatch.Seconds()
	t.Logf("vnet (ws batched): %.2f MB/s  (%.1f MB in %v, flush=2ms)", vnetBatchMBps, float64(nativePayloadBytes)/(1024*1024), elapsedVnetBatch.Round(time.Millisecond))
	if vnetMBps > 0 {
		t.Logf("                   batched is %.2fx unbatched vnet/ws", vnetBatchMBps/vnetMBps)
	}

	// Overhead breakdown
	dropPipeToNativeWS := nativeMBps - nativeWSMBps
	dropNativeWSToVnet := nativeWSMBps - vnetMBps
	totalDrop := nativeMBps - vnetMBps
	t.Logf("overhead:          pipe→vnet/ws total %.0f MB/s drop", totalDrop)
	t.Logf("                   WebSocket (pipe→native/ws): %.0f MB/s lost", dropPipeToNativeWS)
	if dropNativeWSToVnet >= 0 {
		t.Logf("                   vnet (native/ws→vnet/ws):   %.0f MB/s lost", dropNativeWSToVnet)
		if totalDrop > 0 {
			t.Logf("                   → %.0f%% of total drop is WebSocket, %.0f%% is vnet",
				(dropPipeToNativeWS/totalDrop)*100, (dropNativeWSToVnet/totalDrop)*100)
		}
	} else {
		t.Logf("                   vnet (native/ws→vnet/ws):   vnet/ws was faster this run (variance)")
	}
	t.Logf("ratio:             vnet/ws is %.2fx of native pipe", vnetMBps/nativeMBps)
}
