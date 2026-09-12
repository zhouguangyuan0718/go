// asmcheck

//go:build goexperiment.simd && amd64

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "simd/archsimd"

// AVX512 supplies AVX2 instruction capability without forcing the separately
// controllable AVX2 runtime predicate to true. The independent load remains
// observable in the AVX512 clone after the requirement anchor is removed.
// LLVM-AMD64-LABEL: define {{.*}} @codegen.llvmSIMDShiftFoldedStrongIndependent(
// LLVM-AMD64-SAME: #[[STRONG:[0-9]+]]
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: load i8, ptr {{.*}}!goallc.cpu.guard ![[AVX512:[0-9]+]]
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: br i1 {{.*}}, label %[[STRONG_ON:[a-zA-Z0-9_.]+]], label %{{[a-zA-Z0-9_.]+}}
// LLVM-AMD64: [[STRONG_ON]]:
// LLVM-AMD64-NOT: {{^[a-zA-Z0-9_.]+:}}
// LLVM-AMD64: call void @llvm.sideeffect()
// LLVM-AMD64-DAG: !goallc.cpu.requires ![[AVX2:[0-9]+]]
// LLVM-AMD64-DAG: !goallc.cpu.require-anchor ![[ANCHOR:[0-9]+]]
// LLVM-AMD64: {{^[}]}}
// LLVM-AMD64-NOT: "goallc.cpu.multiversion"="x86.avx2
// LLVM-AMD64: attributes #[[STRONG]] = { {{.*}}"goallc.cpu.multiversion"="x86.avx512"
// LLVM-AMD64-DAG: ![[AVX2]] = !{!"x86.avx2"}
// LLVM-AMD64-DAG: ![[AVX512]] = !{!"x86.avx512"}
// LLVM-AMD64-DAG: ![[ANCHOR]] = !{}

// The NOT checks cover the entire optimized module, including the gaps
// between the positive checks.
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-LABEL: define internal {{.*}} @"codegen.llvmSIMDShiftFoldedStrongIndependent<goallc.fmv.avx512>"
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-NOT: {{^[}]}}
// LLVM-OPT-AMD64: load i8, ptr
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64: ret
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor

// Folded shifts generate no shifts or runtime anchor calls.
// LLVM-ASM-AMD64-NOT: llvm.sideeffect
// LLVM-ASM-AMD64-NOT: require-anchor
// LLVM-ASM-AMD64-NOT: VPS{{LL|RL|RA}}
// LLVM-ASM-AMD64: TEXT codegen.llvmSIMDShiftFoldedStrongIndependent(SB)
// LLVM-ASM-AMD64-NOT: llvm.sideeffect
// LLVM-ASM-AMD64-NOT: require-anchor
// LLVM-ASM-AMD64-NOT: VPS{{LL|RL|RA}}
// LLVM-NM-AMD64: D codegen.llvmSIMDShiftFoldedStrongIndependent.goallc.fmv.slot

//go:noinline
func llvmSIMDShiftFoldedStrongIndependent(x archsimd.Uint32x8) (archsimd.Uint32x8, bool) {
	if archsimd.X86.AVX512() {
		return x.ShiftAllLeft(0), archsimd.X86.AVX2()
	}
	return x, false
}
