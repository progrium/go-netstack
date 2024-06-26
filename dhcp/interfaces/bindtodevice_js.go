//go:build js
// +build js

package interfaces

import "golang.org/x/sys/unix"

func BindToInterface(fd int, ifname string) error {
	return unix.BindToDevice(fd, ifname)
}
