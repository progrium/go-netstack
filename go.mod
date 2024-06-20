module github.com/progrium/go-netstack

go 1.22.0

toolchain go1.22.4

replace github.com/progrium/go-netstack/gvisor => ./gvisor

replace github.com/progrium/go-netstack/uio => ./uio

replace github.com/progrium/go-netstack/dhcp => ./dhcp

replace golang.org/x/sys => github.com/progrium/sys-wasm v0.0.0-20240620001524-43ddd9475fa9

require (
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/google/gopacket v1.1.19
	github.com/inetaf/tcpproxy v0.0.0-20240214030015-3ce58045626c
	github.com/miekg/dns v1.1.58
	github.com/pkg/errors v0.9.1
	github.com/progrium/go-netstack/dhcp v0.0.0-00010101000000-000000000000
	github.com/progrium/go-netstack/gvisor v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/btree v1.1.2 // indirect
	github.com/progrium/go-netstack/uio v0.0.0-20210528114334-82958018845c // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
)
