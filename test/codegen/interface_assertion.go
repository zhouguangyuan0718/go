// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

type llvmAssertionInterface interface {
	Value(int) int
}

type llvmAssertionValue int

//go:noinline
func (v llvmAssertionValue) Value(delta int) int {
	return int(v) + delta
}

// LLVM: @codegen..typeAssert.0 = global <{ ptr, ptr, [8 x i8] }>
// LLVM-SAME: !goobj.gotype
// LLVM-LABEL: define goabiinternal { { ptr, ptr }, i8 } @codegen.assertLLVMInterface(
// LLVM: load atomic ptr, ptr @codegen..typeAssert.0 seq_cst
// LLVM: call goabiinternal ptr @runtime.typeAssert(ptr @codegen..typeAssert.0, ptr
// LLVM-DAG: declare goabiinternal ptr @runtime.typeAssert(ptr, ptr)
// LLVM-LABEL: define goabiinternal { i64, i8 } @codegen.assertLLVMConcrete(
// LLVM: extractvalue { ptr, ptr }
// LLVM: icmp eq ptr
// LLVM-SAME: @"type:codegen.llvmAssertionValue"
// LLVM: load i64, ptr
func assertLLVMConcrete(v any) (llvmAssertionValue, bool) {
	x, ok := v.(llvmAssertionValue)
	return x, ok
}

func assertLLVMInterface(v any) (llvmAssertionInterface, bool) {
	x, ok := v.(llvmAssertionInterface)
	return x, ok
}
