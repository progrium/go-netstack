package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"

	gvntypes "github.com/progrium/go-netstack/types"
	gvnvirtualnetwork "github.com/progrium/go-netstack/virtualnetwork"
)

func main() {
	// just two paired sockets so we can
	// give the virtual network something
	vsock, _ := net.Pipe()

	config := &gvntypes.Configuration{
		Debug:             false,
		MTU:               1500,
		Subnet:            "192.168.127.0/24",
		GatewayIP:         "192.168.127.1",
		GatewayMacAddress: "5a:94:ef:e4:0c:dd",
		GatewayVirtualIPs: []string{"192.168.127.253"},
		Protocol:          gvntypes.QemuProtocol,
	}
	vn, err := gvnvirtualnetwork.New(config)
	if err != nil {
		panic(err)
	}

	go func() {
		s := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "Hello world")
			}),
		}
		l, err := vn.Listen("tcp", "192.168.127.253:80")
		if err != nil {
			panic(err)
		}
		log.Fatal(s.Serve(l))
	}()

	if err := vn.AcceptQemu(context.TODO(), vsock); err != nil {
		panic(err)
	}
}
