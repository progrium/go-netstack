// Copyright 2018 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build amd64
// +build amd64

package linux

import (
	"github.com/progrium/go-netstack/gvisor/pkg/abi/linux"
	"github.com/progrium/go-netstack/gvisor/pkg/errors/linuxerr"
	"github.com/progrium/go-netstack/gvisor/pkg/marshal/primitive"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/arch"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/kernel"
)

// ArchPrctl implements linux syscall arch_prctl(2).
// It sets architecture-specific process or thread state for t.
func ArchPrctl(t *kernel.Task, sysno uintptr, args arch.SyscallArguments) (uintptr, *kernel.SyscallControl, error) {
	switch args[0].Int() {
	case linux.ARCH_GET_FS:
		addr := args[1].Pointer()
		fsbase := t.Arch().TLS()
		switch t.Arch().Width() {
		case 8:
			if _, err := primitive.CopyUint64Out(t, addr, uint64(fsbase)); err != nil {
				return 0, nil, err
			}
		default:
			return 0, nil, linuxerr.ENOSYS
		}
	case linux.ARCH_SET_FS:
		fsbase := args[1].Uint64()
		if !t.Arch().SetTLS(uintptr(fsbase)) {
			return 0, nil, linuxerr.EPERM
		}
	case linux.ARCH_GET_GS, linux.ARCH_SET_GS:
		t.Kernel().EmitUnimplementedEvent(t, sysno)
		fallthrough
	default:
		return 0, nil, linuxerr.EINVAL
	}

	return 0, nil, nil
}
