// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal i64 @"codegen.(*llvmTailInter).dynamic"
// LLVM-NOT: tail call
// LLVM: %[[INTER_CALL:[0-9]+]] = call goabiinternal i64 %{{[0-9]+}}
// LLVM-NEXT: ret i64 %[[INTER_CALL]]
// LLVM-LABEL: define goabiinternal i64 @"codegen.(*llvmTailStatic).static"
// LLVM-NOT: tail call
// LLVM: %[[STATIC_CALL:[0-9]+]] = call goabiinternal i64 @"codegen.(*llvmTailBase).static"
// LLVM-NEXT: ret i64 %[[STATIC_CALL]]

type llvmTailBase struct{}

//go:noinline
func (*llvmTailBase) static(value int) int {
	return value + 1
}

type llvmTailStatic struct {
	*llvmTailBase
}

// Force the compiler-generated static method wrapper.
var llvmTailStaticMethod = (*llvmTailStatic).static

type llvmTailInterface interface {
	dynamic(int) int
}

type llvmTailInter struct {
	llvmTailInterface
}

// Force the compiler-generated interface method wrapper.
var llvmTailInterMethod = (*llvmTailInter).dynamic
