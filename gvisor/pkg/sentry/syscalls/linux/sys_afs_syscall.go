// Copyright 2023 The gVisor Authors.
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

package linux

import (
	"github.com/progrium/go-netstack/gvisor/pkg/atomicbitops"
	"github.com/progrium/go-netstack/gvisor/pkg/errors/linuxerr"
	"github.com/progrium/go-netstack/gvisor/pkg/log"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/arch"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/kernel"
)

var afsSyscallPanic = atomicbitops.Bool{}

// SetAFSSyscallPanic sets the panic behaviour of afs_syscall.
// Should only be called based on the config value of TESTONLY-afs-syscall-panic.
func SetAFSSyscallPanic(v bool) {
	if v {
		log.Warningf("AFSSyscallPanic is set. User workloads may trigger sentry panics.")
	}
	afsSyscallPanic.Store(v)
}

// AFSSyscall is a gVisor specific implementation of afs_syscall:
// - if TESTONLY-afs-syscall-panic flag is set it triggers a panic.
func AFSSyscall(t *kernel.Task, sysno uintptr, args arch.SyscallArguments) (uintptr, *kernel.SyscallControl, error) {
	if afsSyscallPanic.Load() {
		panic("User workload triggered a panic via afs_syscall. This panic is intentional.")
	}

	return 0, nil, linuxerr.ENOSYS
}
