package vnet

import (
	"context"
	"net"

	"github.com/progrium/go-netstack/types"
)

func (n *VirtualNetwork) AcceptQemu(ctx context.Context, conn net.Conn) error {
	return n.networkSwitch.Accept(ctx, conn, types.QemuProtocol)
}
