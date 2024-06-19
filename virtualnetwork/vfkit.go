package virtualnetwork

import (
	"context"
	"net"

	"github.com/progrium/go-netstack/types"
)

func (n *VirtualNetwork) AcceptVfkit(ctx context.Context, conn net.Conn) error {
	return n.networkSwitch.Accept(ctx, conn, types.VfkitProtocol)
}
