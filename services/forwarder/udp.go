package forwarder

import (
	"fmt"
	"net"
	"sync"

	"github.com/progrium/go-netstack/gvisor/pkg/tcpip"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/adapters/gonet"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/header"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/stack"
	"github.com/progrium/go-netstack/gvisor/pkg/tcpip/transport/udp"
	"github.com/progrium/go-netstack/gvisor/pkg/waiter"
	log "github.com/sirupsen/logrus"
)

func UDP(s *stack.Stack, nat map[tcpip.Address]tcpip.Address, natLock *sync.Mutex) *udp.Forwarder {
	return udp.NewForwarder(s, func(r *udp.ForwarderRequest) {
		localAddress := r.ID().LocalAddress

		if linkLocal().Contains(localAddress) || localAddress == header.IPv4Broadcast {
			return
		}

		natLock.Lock()
		if replaced, ok := nat[localAddress]; ok {
			localAddress = replaced
		}
		natLock.Unlock()

		var wq waiter.Queue
		ep, tcpErr := r.CreateEndpoint(&wq)
		if tcpErr != nil {
			log.Errorf("r.CreateEndpoint() = %v", tcpErr)
			return
		}

		p, _ := NewUDPProxy(&autoStoppingListener{underlying: gonet.NewUDPConn(s, &wq, ep)}, func() (net.Conn, error) {
			return net.Dial("udp", fmt.Sprintf("%s:%d", localAddress, r.ID().LocalPort))
		})
		go p.Run()
	})
}
