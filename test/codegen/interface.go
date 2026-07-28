// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-DAG: %go.runtime.ITab = type <{ ptr, ptr, i32, [4 x i8], [1 x ptr] }>
// LLVM-DAG: %go.itab.codegen.llvmInterfaceValue_codegen.llvmInterface = type <{ %go.runtime.ITab, [1 x ptr] }>
// LLVM-DAG: @"go:itab.codegen.llvmInterfaceValue,codegen.llvmInterface" = weak constant %go.itab.codegen.llvmInterfaceValue_codegen.llvmInterface
// LLVM-DAG: ptr @"codegen.(*llvmInterfaceValue).Double"
// LLVM-DAG: ptr @"codegen.(*llvmInterfaceValue).Value"
// LLVM-DAG: !goobj.weak_relocs
// LLVM-LABEL: define goabiinternal i64 @codegen.useLLVMInterface(
// LLVM-SAME: !goobj.marker_relocs
// LLVM-LABEL: define goabiinternal i64 @codegen.llvmInterface.Value(
// LLVM: load ptr, ptr
// LLVM: call goabiinternal i64 %
// LLVM-SAME: (ptr
// LLVM-SAME: i64
// LLVM-LABEL: define goabiinternal i64 @codegen.llvmInterface.Double(
// LLVM: load ptr, ptr
// LLVM: call goabiinternal i64 %
// LLVM-SAME: (ptr
// LLVM: !{i32 23, i64 0, !"type:codegen.llvmInterfaceValue"}
// LLVM-DAG: !{i32 24, i64 {{[0-9]+}}, !"type:codegen.llvmInterface"}

type llvmInterface interface {
	Double() int
	Value(int) int
}

type llvmInterfaceValue struct {
	value int
}

//go:noinline
func (v llvmInterfaceValue) Value(delta int) int {
	return v.value + delta
}

//go:noinline
func (v llvmInterfaceValue) Double() int {
	return v.value * 2
}

func useLLVMInterface(v llvmInterfaceValue) int {
	var i llvmInterface = v
	return i.Value(2) + i.Double()
}
