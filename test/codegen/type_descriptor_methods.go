// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-DAG: %go.runtime.Method = type <{ i32, i32, i32, i32 }>
// LLVM-DAG: %go.descriptor.codegen.llvmMethodType = type <{ %go.runtime.StructType, %go.runtime.UncommonType, [1 x %go.runtime.StructField], [2 x %go.runtime.Method] }>
// LLVM-DAG: %go.descriptor._codegen.llvmMethodType = type <{ %go.runtime.PtrType, %go.runtime.UncommonType, [3 x %go.runtime.Method] }>
// LLVM-DAG: @"type:codegen.llvmMethodType" = constant %go.descriptor.codegen.llvmMethodType
// LLVM-DAG: @"type:*codegen.llvmMethodType" = constant %go.descriptor._codegen.llvmMethodType
// LLVM-DAG: !{{[0-9]+}} = !{i32 {{[0-9]+}}, i32 26}

type llvmMethodType struct {
	value int
}

func (v llvmMethodType) Value(delta int) int {
	return v.value + delta
}

func (v llvmMethodType) hidden() int {
	return v.value
}

func (v *llvmMethodType) Pointer(delta int) int {
	return v.value + delta
}

func useLLVMMethodType(v llvmMethodType) int {
	return v.Value(2) + (&v).Pointer(3)
}
