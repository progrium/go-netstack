package virtualnetwork

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/progrium/go-netstack/gvisor/pkg/tcpip"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/adapters/gonet"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/network/ipv4"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/stack"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/transport/tcp"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/transport/udp"
	"github.com/progrium/go-netstack/services/dhcp"
	"github.com/progrium/go-netstack/services/dns"
	"github.com/progrium/go-netstack/services/forwarder"
	"github.com/progrium/go-netstack/tap"
	"github.com/progrium/go-netstack/types"
	log "github.com/sirupsen/logrus"
)

func addServices(configuration *types.Configuration, s *stack.Stack, ipPool *tap.IPPool) (http.Handler, error) {
	var natLock sync.Mutex
	translation := parseNATTable(configuration)

	tcpForwarder := forwarder.TCP(s, translation, &natLock)
	s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
	udpForwarder := forwarder.UDP(s, translation, &natLock)
	s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)

	dnsMux, err := dnsServer(configuration, s)
	if err != nil {
		return nil, err
	}

	dhcpMux, err := dhcpServer(configuration, s, ipPool)
	if err != nil {
		return nil, err
	}

	forwarderMux, err := forwardHostVM(configuration, s)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/forwarder/", http.StripPrefix("/forwarder", forwarderMux))
	mux.Handle("/dhcp/", http.StripPrefix("/dhcp", dhcpMux))
	mux.Handle("/dns/", http.StripPrefix("/dns", dnsMux))
	return mux, nil
}

func parseNATTable(configuration *types.Configuration) map[tcpip.Address]tcpip.Address {
	translation := make(map[tcpip.Address]tcpip.Address)
	for source, destination := range configuration.NAT {
		translation[tcpip.AddrFrom4Slice(net.ParseIP(source).To4())] = tcpip.AddrFrom4Slice(net.ParseIP(destination).To4())
	}
	return translation
}

func dnsServer(configuration *types.Configuration, s *stack.Stack) (http.Handler, error) {
	udpConn, err := gonet.DialUDP(s, &tcpip.FullAddress{
		NIC:  1,
		Addr: tcpip.AddrFrom4Slice(net.ParseIP(configuration.GatewayIP).To4()),
		Port: uint16(53),
	}, nil, ipv4.ProtocolNumber)
	if err != nil {
		return nil, err
	}

	tcpLn, err := gonet.ListenTCP(s, tcpip.FullAddress{
		NIC:  1,
		Addr: tcpip.AddrFrom4Slice(net.ParseIP(configuration.GatewayIP).To4()),
		Port: uint16(53),
	}, ipv4.ProtocolNumber)
	if err != nil {
		return nil, err
	}

	server, err := dns.New(udpConn, tcpLn, configuration.DNS)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := server.Serve(); err != nil {
			log.Error(err)
		}
	}()
	go func() {
		if err := server.ServeTCP(); err != nil {
			log.Error(err)
		}
	}()
	return server.Mux(), nil
}

func dhcpServer(configuration *types.Configuration, s *stack.Stack, ipPool *tap.IPPool) (http.Handler, error) {
	server, err := dhcp.New(configuration, s, ipPool)
	if err != nil {
		return nil, err
	}
	go func() {
		log.Error(server.Serve())
	}()
	return server.Mux(), nil
}

func forwardHostVM(configuration *types.Configuration, s *stack.Stack) (http.Handler, error) {
	fw := forwarder.NewPortsForwarder(s)
	for local, remote := range configuration.Forwards {
		if strings.HasPrefix(local, "udp:") {
			if err := fw.Expose(types.UDP, strings.TrimPrefix(local, "udp:"), remote); err != nil {
				return nil, err
			}
		} else {
			if err := fw.Expose(types.TCP, local, remote); err != nil {
				return nil, err
			}
		}
	}
	return fw.Mux(), nil
}
