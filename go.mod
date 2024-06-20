module github.com/progrium/go-netstack

go 1.22.0

toolchain go1.22.4

replace golang.org/x/sys => github.com/progrium/sys-wasm v0.0.0-20240620001524-43ddd9475fa9

require (
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/bazelbuild/rules_go v0.48.1
	github.com/fanliao/go-promise v0.0.0-20141029170127-1890db352a72
	github.com/google/btree v1.1.2
	github.com/google/gopacket v1.1.19
	github.com/inetaf/tcpproxy v0.0.0-20240214030015-3ce58045626c
	github.com/mdlayher/ethernet v0.0.0-20220221185849-529eae5b6118
	github.com/mdlayher/raw v0.1.0
	github.com/miekg/dns v1.1.58
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8
	golang.org/x/mod v0.18.0
	golang.org/x/net v0.26.0
	golang.org/x/sys v0.21.0
	golang.org/x/time v0.5.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/josharian/native v1.0.0 // indirect
	github.com/mdlayher/packet v1.0.0 // indirect
	github.com/mdlayher/socket v0.2.1 // indirect
	github.com/smartystreets/goconvey v1.8.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
)
