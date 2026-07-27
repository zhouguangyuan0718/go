// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

type llvmSwitchInterface interface {
	Value(int) int
}

type llvmSwitchValue int

//go:noinline
func (v llvmSwitchValue) Value(delta int) int {
	return int(v) + delta
}

// LLVM-DAG: @codegen..interfaceSwitch.0 = global <{ ptr, [8 x i8], ptr }> {{.*}}!goobj.gotype
// LLVM-DAG: load atomic ptr, ptr @codegen..interfaceSwitch.0 seq_cst
// LLVM-DAG: call goabiinternal { i64, ptr } @runtime.interfaceSwitch(ptr @codegen..interfaceSwitch.0, ptr
// LLVM-DAG: declare goabiinternal { i64, ptr } @runtime.interfaceSwitch(ptr, ptr)
// LLVM-DAG: icmp eq ptr
// LLVM-DAG: @"type:codegen.llvmSwitchValue"
// LLVM-DAG: @"type:string"
// LLVM-DAG: ret i64 -1
func classifyLLVMInterface(v any) int {
	switch x := v.(type) {
	case nil:
		return 0
	case llvmSwitchValue:
		return int(x)
	case string:
		return len(x)
	case llvmSwitchInterface:
		return x.Value(1)
	default:
		return -1
	}
}
