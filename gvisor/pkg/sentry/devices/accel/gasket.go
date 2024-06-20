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

package accel

import (
	"fmt"

	"github.com/progrium/go-netstack/gvisor/pkg/abi/gasket"
	"github.com/progrium/go-netstack/gvisor/pkg/abi/linux"
	"github.com/progrium/go-netstack/gvisor/pkg/cleanup"
	"github.com/progrium/go-netstack/gvisor/pkg/context"
	"github.com/progrium/go-netstack/gvisor/pkg/errors/linuxerr"
	"github.com/progrium/go-netstack/gvisor/pkg/hostarch"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/fsimpl/eventfd"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/kernel"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/memmap"
	"github.com/progrium/go-netstack/gvisor/pkg/sentry/mm"
	"golang.org/x/sys/unix"
)

func gasketMapBufferIoctl(ctx context.Context, t *kernel.Task, hostFd int32, fd *accelFD, paramsAddr hostarch.Addr) (uintptr, error) {
	var userIoctlParams gasket.GasketPageTableIoctl
	if _, err := userIoctlParams.CopyIn(t, paramsAddr); err != nil {
		return 0, err
	}

	tmm := t.MemoryManager()
	ar, ok := tmm.CheckIORange(hostarch.Addr(userIoctlParams.HostAddress), int64(userIoctlParams.Size))
	if !ok {
		return 0, linuxerr.EFAULT
	}

	if !ar.IsPageAligned() || (userIoctlParams.Size/hostarch.PageSize) == 0 {
		return 0, linuxerr.EINVAL
	}
	// Reserve a range in our address space.
	m, _, errno := unix.RawSyscall6(unix.SYS_MMAP, 0 /* addr */, uintptr(ar.Length()), unix.PROT_NONE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS, ^uintptr(0) /* fd */, 0 /* offset */)
	if errno != 0 {
		return 0, errno
	}
	cu := cleanup.Make(func() {
		unix.RawSyscall(unix.SYS_MUNMAP, m, uintptr(ar.Length()), 0)
	})
	defer cu.Clean()
	// Mirror application mappings into the reserved range.
	prs, err := t.MemoryManager().Pin(ctx, ar, hostarch.ReadWrite, false /* ignorePermissions */)
	cu.Add(func() {
		mm.Unpin(prs)
	})
	if err != nil {
		return 0, err
	}
	sentryAddr := uintptr(m)
	for _, pr := range prs {
		ims, err := pr.File.MapInternal(memmap.FileRange{pr.Offset, pr.Offset + uint64(pr.Source.Length())}, hostarch.ReadWrite)
		if err != nil {
			return 0, err
		}
		for !ims.IsEmpty() {
			im := ims.Head()
			if _, _, errno := unix.RawSyscall6(unix.SYS_MREMAP, im.Addr(), 0 /* old_size */, uintptr(im.Len()), linux.MREMAP_MAYMOVE|linux.MREMAP_FIXED, sentryAddr, 0); errno != 0 {
				return 0, errno
			}
			sentryAddr += uintptr(im.Len())
			ims = ims.Tail()
		}
	}
	sentryIoctlParams := userIoctlParams
	sentryIoctlParams.HostAddress = uint64(m)
	n, err := ioctlInvokePtrArg(hostFd, gasket.GASKET_IOCTL_MAP_BUFFER, &sentryIoctlParams)
	if err != nil {
		return n, err
	}
	cu.Release()
	// Unmap the reserved range, which is no longer required.
	unix.RawSyscall(unix.SYS_MUNMAP, m, uintptr(ar.Length()), 0)

	fd.device.mu.Lock()
	defer fd.device.mu.Unlock()
	devAddr := userIoctlParams.DeviceAddress
	for _, pr := range prs {
		rlen := uint64(pr.Source.Length())
		if !fd.device.devAddrSet.Add(DevAddrRange{
			devAddr,
			devAddr + rlen,
		}, pinnedAccelMem{pinnedRange: pr, pageTableIndex: userIoctlParams.PageTableIndex}) {
			panic(fmt.Sprintf("unexpected overlap of devaddr range [%#x-%#x)", devAddr, devAddr+rlen))
		}
		devAddr += rlen
	}
	return n, nil
}

func gasketUnmapBufferIoctl(ctx context.Context, t *kernel.Task, hostFd int32, fd *accelFD, paramsAddr hostarch.Addr) (uintptr, error) {
	var userIoctlParams gasket.GasketPageTableIoctl
	if _, err := userIoctlParams.CopyIn(t, paramsAddr); err != nil {
		return 0, err
	}
	sentryIoctlParams := userIoctlParams
	sentryIoctlParams.HostAddress = 0 // clobber this value, it's unused.
	n, err := ioctlInvokePtrArg(hostFd, gasket.GASKET_IOCTL_UNMAP_BUFFER, &sentryIoctlParams)
	if err != nil {
		return n, err
	}
	fd.device.mu.Lock()
	defer fd.device.mu.Unlock()
	s := &fd.device.devAddrSet
	r := DevAddrRange{userIoctlParams.DeviceAddress, userIoctlParams.DeviceAddress + userIoctlParams.Size}
	seg := s.LowerBoundSegment(r.Start)
	for seg.Ok() && seg.Start() < r.End {
		seg = s.Isolate(seg, r)
		v := seg.Value()
		mm.Unpin([]mm.PinnedRange{v.pinnedRange})
		gap := s.Remove(seg)
		seg = gap.NextSegment()
	}
	return n, nil
}

func gasketInterruptMappingIoctl(ctx context.Context, t *kernel.Task, hostFd int32, paramsAddr hostarch.Addr) (uintptr, error) {
	var userIoctlParams gasket.GasketInterruptMapping
	if _, err := userIoctlParams.CopyIn(t, paramsAddr); err != nil {
		return 0, err
	}

	// Check that 'userEventFD.Eventfd' is an eventfd.
	eventFileGeneric, _ := t.FDTable().Get(int32(userIoctlParams.EventFD))
	if eventFileGeneric == nil {
		return 0, linuxerr.EBADF
	}
	defer eventFileGeneric.DecRef(ctx)
	eventFile, ok := eventFileGeneric.Impl().(*eventfd.EventFileDescription)
	if !ok {
		return 0, linuxerr.EINVAL
	}

	eventfd, err := eventFile.HostFD()
	if err != nil {
		return 0, err
	}

	sentryIoctlParams := userIoctlParams
	sentryIoctlParams.EventFD = uint64(eventfd)
	n, err := ioctlInvokePtrArg(hostFd, gasket.GASKET_IOCTL_REGISTER_INTERRUPT, &sentryIoctlParams)
	if err != nil {
		return n, err
	}

	outIoctlParams := sentryIoctlParams
	outIoctlParams.EventFD = userIoctlParams.EventFD
	if _, err := outIoctlParams.CopyOut(t, paramsAddr); err != nil {
		return n, err
	}
	return n, nil
}
