// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "unsafe"

//go:linkname llvmAllocate runtime.mallocgc
func llvmAllocate(size uintptr, typ unsafe.Pointer, zero bool) unsafe.Pointer

func llvmAllocZeroed() unsafe.Pointer                  { return llvmAllocate(8, nil, true) }
func llvmAllocUnknownContents() unsafe.Pointer         { return llvmAllocate(8, nil, false) }
func llvmAllocDynamicZero(zero bool) unsafe.Pointer    { return llvmAllocate(8, nil, zero) }
func llvmAllocZeroSize() unsafe.Pointer                { return llvmAllocate(0, nil, true) }
func llvmAllocDynamicSize(size uintptr) unsafe.Pointer { return llvmAllocate(size, nil, true) }

func llvmAllocUnused()          { _ = llvmAllocate(8, nil, true) }
func llvmAllocReadZero() uint64 { return *(*uint64)(llvmAllocate(8, nil, true)) }

// LLVM-DAG: call goabiinternal noalias ptr @"runtime.mallocgc<builtin.{{[0-9]+}}>"(i64 8, ptr null, i8 1) #[[ZERO:[0-9]+]]
// LLVM-DAG: call goabiinternal noalias ptr @"runtime.mallocgc<builtin.{{[0-9]+}}>"(i64 8, ptr null, i8 0) #[[ALLOC:[0-9]+]]
// LLVM-DAG: call goabiinternal noalias ptr @"runtime.mallocgc<builtin.{{[0-9]+}}>"(i64 8, ptr null, i8 %{{[^)]+}}) #[[ALLOC]]
// LLVM-DAG: call goabiinternal ptr @"runtime.mallocgc<builtin.{{[0-9]+}}>"(i64 0, ptr null, i8 1), !dbg
// LLVM-DAG: call goabiinternal ptr @"runtime.mallocgc<builtin.{{[0-9]+}}>"(i64 %size, ptr null, i8 1), !dbg
// LLVM-DAG: attributes #[[ZERO]] = { allockind("alloc,zeroed") "alloc-family"="runtime.mallocgc" }
// LLVM-DAG: attributes #[[ALLOC]] = { allockind("alloc") "alloc-family"="runtime.mallocgc" }

// LLVM-OPT-LABEL: define goabiinternal {{.*}}i64 @codegen.llvmAllocReadZero()
// LLVM-OPT-NEXT: b1:
// LLVM-OPT-NEXT: ret i64 0,
// LLVM-OPT-NEXT: }
