// asmcheck

//go:build goexperiment.simd && amd64

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "simd/archsimd"

// Both operands are constant, so the IR builder returns a constant directly.
// Requirement metadata must never be attached to that non-instruction value.
// LLVM-AMD64-LABEL: define {{.*}} @codegen.llvmSIMDShiftFoldedConstant(
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: load i8, ptr {{.*}}!goallc.cpu.guard ![[AVX2:[0-9]+]]
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: br i1 {{.*}}, label %[[CONSTANT_ON:[a-zA-Z0-9_.]+]], label %{{[a-zA-Z0-9_.]+}}
// LLVM-AMD64: [[CONSTANT_ON]]:
// LLVM-AMD64-NOT: {{^[a-zA-Z0-9_.]+:}}
// LLVM-AMD64: call void @llvm.sideeffect()
// LLVM-AMD64-DAG: !goallc.cpu.requires ![[AVX2]]
// LLVM-AMD64-DAG: !goallc.cpu.require-anchor ![[ANCHOR:[0-9]+]]
// LLVM-AMD64: {{^[}]}}
// LLVM-AMD64-DAG: ![[AVX2]] = !{!"x86.avx2"}
// LLVM-AMD64-DAG: ![[ANCHOR]] = !{}

// The NOT checks cover the entire optimized module, including the gaps
// between the positive checks.
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-LABEL: define internal {{.*}} @"codegen.llvmSIMDShiftFoldedConstant<goallc.fmv.avx2>"
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-NOT: {{^[}]}}
// LLVM-OPT-AMD64: ret <8 x i32> zeroinitializer
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor

// Folded shifts generate no shifts or runtime anchor calls.
// LLVM-ASM-AMD64-NOT: llvm.sideeffect
// LLVM-ASM-AMD64-NOT: require-anchor
// LLVM-ASM-AMD64-NOT: VPS{{LL|RL|RA}}
// LLVM-ASM-AMD64: TEXT codegen.llvmSIMDShiftFoldedConstant(SB)
// LLVM-ASM-AMD64-NOT: llvm.sideeffect
// LLVM-ASM-AMD64-NOT: require-anchor
// LLVM-ASM-AMD64-NOT: VPS{{LL|RL|RA}}
// LLVM-NM-AMD64: D codegen.llvmSIMDShiftFoldedConstant.goallc.fmv.slot

//go:noinline
func llvmSIMDShiftFoldedConstant() archsimd.Uint32x8 {
	if archsimd.X86.AVX2() {
		return (archsimd.Uint32x8{}).ShiftAllLeft(1)
	}
	return archsimd.Uint32x8{}
}
