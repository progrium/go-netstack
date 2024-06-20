// Copyright 2018 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64 || arm || arm64 || mipsle || mips64le || ppc64le || riscv || riscv64 || wasm
// +build 386 amd64 arm arm64 mipsle mips64le ppc64le riscv riscv64 wasm

package ubinary

import (
	"encoding/binary"
)

// NativeEndian is $GOARCH's implementation of byte order.
var NativeEndian = binary.LittleEndian
