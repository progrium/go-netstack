package virtualnetwork

import (
	"context"
	"net"

	"github.com/progrium/go-netstack/types"
)

func (n *VirtualNetwork) AcceptStdio(ctx context.Context, conn net.Conn) error {
	return n.networkSwitch.Accept(ctx, conn, types.StdioProtocol)
}
