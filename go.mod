module github.com/progrium/go-netstack

go 1.22.0

toolchain go1.22.4

replace gvisor.dev/gvisor => ./gvisor
replace golang.org/x/sys => ./sys

require (
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/google/gopacket v1.1.19
	github.com/inetaf/tcpproxy v0.0.0-20240214030015-3ce58045626c
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	gvisor.dev/gvisor v0.0.0-20240618223457-75c9597d8ec8
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/containers/gvisor-tap-vsock v0.7.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/insomniacslk/dhcp v0.0.0-20220504074936-1ca156eafb9f // indirect
	github.com/miekg/dns v1.1.58 // indirect
	github.com/u-root/uio v0.0.0-20210528114334-82958018845c // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
)
