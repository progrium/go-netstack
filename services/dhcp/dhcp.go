package dhcp

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/progrium/go-netstack/dhcp/dhcpv4"
	"github.com/progrium/go-netstack/dhcp/dhcpv4/server4"
	"github.com/progrium/go-netstack/dhcp/rfc1035label"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/adapters/gonet"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/network/ipv4"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/stack"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/transport/udp"
	"github.com/progrium/go-netstack/gvisor/pkg/waiter"
	"github.com/progrium/go-netstack/tap"
	"github.com/progrium/go-netstack/types"
	log "github.com/sirupsen/logrus"
)

const serverPort = 67

func handler(configuration *types.Configuration, ipPool *tap.IPPool) server4.Handler {
	return func(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
		reply, err := dhcpv4.NewReplyFromRequest(m)
		if err != nil {
			log.Errorf("dhcp: cannot build reply from request: %v", err)
			return
		}

		ip, err := ipPool.GetOrAssign(m.ClientHWAddr.String())
		if err != nil {
			log.Errorf("dhcp: cannot assign ip: %v", err)
			return
		}

		_, parsedSubnet, err := net.ParseCIDR(configuration.Subnet)
		if err != nil {
			log.Errorf("dhcp: invalid subnet %v", err)
			return
		}

		reply.YourIPAddr = ip
		reply.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP(configuration.GatewayIP)))
		reply.UpdateOption(dhcpv4.OptIPAddressLeaseTime(time.Hour))

		reply.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionSubnetMask, Value: dhcpv4.IP(parsedSubnet.Mask)})
		reply.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionRouter, Value: dhcpv4.IP(net.ParseIP(configuration.GatewayIP))})
		reply.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionDomainNameServer, Value: dhcpv4.IPs([]net.IP{net.ParseIP(configuration.GatewayIP)})})
		reply.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionInterfaceMTU, Value: dhcpv4.Uint16(configuration.MTU)})
		reply.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionDNSDomainSearchList, Value: &rfc1035label.Labels{
			Labels: configuration.DNSSearchDomains,
		}})

		switch mt := m.MessageType(); mt {
		case dhcpv4.MessageTypeDiscover:
			reply.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
		case dhcpv4.MessageTypeRequest:
			reply.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		default:
			log.Errorf("dhcp: unhandled message type: %v", mt)
			return
		}

		if _, err := conn.WriteTo(reply.ToBytes(), peer); err != nil {
			log.Errorf("dhcp: cannot reply to client: %v", err)
		}
	}
}

func dial(s *stack.Stack, nic int) (*gonet.UDPConn, error) {
	var wq waiter.Queue
	ep, err := s.NewEndpoint(udp.ProtocolNumber, ipv4.ProtocolNumber, &wq)
	if err != nil {
		return nil, errors.New(err.String())
	}

	ep.SocketOptions().SetBroadcast(true)

	if err := ep.Bind(tcpip.FullAddress{
		NIC:  tcpip.NICID(nic),
		Addr: tcpip.Address{},
		Port: uint16(serverPort),
	}); err != nil {
		ep.Close()
		return nil, errors.New(err.String())
	}

	return gonet.NewUDPConn(s, &wq, ep), nil
}

type Server struct {
	Underlying *server4.Server
	IPPool     *tap.IPPool
}

func New(configuration *types.Configuration, stack *stack.Stack, ipPool *tap.IPPool) (*Server, error) {
	ln, err := dial(stack, 1)
	if err != nil {
		return nil, err
	}

	s, err := server4.NewServer("", nil, handler(configuration, ipPool), server4.WithConn(ln))
	if err != nil {
		return nil, err
	}

	return &Server{
		Underlying: s,
		IPPool:     ipPool,
	}, nil
}

func (s *Server) Serve() error {
	return s.Underlying.Serve()
}

func (s *Server) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/leases", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(s.IPPool.Leases())
	})
	return mux
}
