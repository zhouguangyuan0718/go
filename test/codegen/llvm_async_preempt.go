// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal i64 @codegen.llvmAsyncPreemptLoop(
// LLVM-SAME: #[[DEFAULT:[0-9]+]] gc "goallc"
// LLVM-LABEL: define goabiinternal i64 @codegen.llvmAsyncPreemptNoSplit(
// LLVM-SAME: #[[NOSPLIT:[0-9]+]] gc "goallc"
// LLVM: attributes #[[DEFAULT]] = { {{.*}}"go-async-unsafe"{{.*}} }
// LLVM: attributes #[[NOSPLIT]] = { {{.*}}"go-async-unsafe"{{.*}}"go-nosplit"{{.*}} }
//
//go:noinline
func llvmAsyncPreemptLoop(limit uint64) uint64 {
	var sum uint64
	for i := uint64(0); i < limit; i++ {
		sum += i*17 + 3
	}
	return sum
}

//go:noinline
//go:nosplit
func llvmAsyncPreemptNoSplit(value uint64) uint64 {
	return value + 1
}
