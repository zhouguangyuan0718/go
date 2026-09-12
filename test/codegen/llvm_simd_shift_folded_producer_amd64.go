// asmcheck

//go:build goexperiment.simd && amd64

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "simd/archsimd"

// This fixed-vector signature supplies the AVX floor. The independent AVX
// predicate stays a runtime load, and the add dominates a separate AVX512
// guard. The shift's requirement belongs in the nested true block, without
// attaching it to the earlier producer instruction.
// LLVM-AMD64-LABEL: define {{.*}} @codegen.llvmSIMDShiftFoldedProducer(
// LLVM-AMD64-SAME: #[[PRODUCER:[0-9]+]]
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: [[AVX_LOAD:%[a-zA-Z0-9_.]+]] = load i8, ptr
// LLVM-AMD64-NOT: !goallc.cpu.guard
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: [[AVX_COND:%[a-zA-Z0-9_.]+]] = icmp ne i8 [[AVX_LOAD]], 0
// LLVM-AMD64: br i1 [[AVX_COND]], label %[[ADD_ON:[a-zA-Z0-9_.]+]], label %{{[a-zA-Z0-9_.]+}}
// LLVM-AMD64: [[ADD_ON]]:
// LLVM-AMD64-NOT: {{^[a-zA-Z0-9_.]+:}}
// LLVM-AMD64: add <2 x i64>
// LLVM-AMD64-NOT: !goallc.cpu.requires
// LLVM-AMD64-NOT: {{^[a-zA-Z0-9_.]+:}}
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: load i8, ptr {{.*}}!goallc.cpu.guard ![[AVX512:[0-9]+]]
// LLVM-AMD64-NOT: @llvm.sideeffect
// LLVM-AMD64: br i1 {{.*}}, label %[[PRODUCER_ON:[a-zA-Z0-9_.]+]], label %{{[a-zA-Z0-9_.]+}}
// LLVM-AMD64: [[PRODUCER_ON]]:
// LLVM-AMD64-NOT: {{^[a-zA-Z0-9_.]+:}}
// LLVM-AMD64: call void @llvm.sideeffect()
// LLVM-AMD64-DAG: !goallc.cpu.requires ![[AVX512]]
// LLVM-AMD64-DAG: !goallc.cpu.require-anchor ![[ANCHOR:[0-9]+]]
// LLVM-AMD64: {{^[}]}}
// LLVM-AMD64: attributes #[[PRODUCER]] = { {{.*}}"goallc.cpu.feature-floor"="x86.avx"{{.*}}"goallc.cpu.multiversion"="x86.avx512"
// LLVM-AMD64-DAG: ![[AVX512]] = !{!"x86.avx512"}
// LLVM-AMD64-DAG: ![[ANCHOR]] = !{}

// The NOT checks cover the entire optimized module, including the gaps
// between the positive checks. Specializing AVX512 does not force the
// independent AVX predicate true: the clone retains both its load and add.
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-LABEL: define internal {{.*}} @"codegen.llvmSIMDShiftFoldedProducer<goallc.fmv.avx512>"
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-NOT: {{^[}]}}
// LLVM-OPT-AMD64: load i8, ptr
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64-NOT: {{^[}]}}
// LLVM-OPT-AMD64: add <2 x i64>
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor
// LLVM-OPT-AMD64: ret
// LLVM-OPT-AMD64-NOT: call void @llvm.sideeffect
// LLVM-OPT-AMD64-NOT: !goallc.cpu.require-anchor

// Folded shifts generate no shifts or runtime anchor calls.
// LLVM-ASM-AMD64-NOT: llvm.sideeffect
// LLVM-ASM-AMD64-NOT: require-anchor
// LLVM-ASM-AMD64-NOT: VPS{{LL|RL|RA}}
// LLVM-ASM-AMD64: VPADDQ
// LLVM-ASM-AMD64-NOT: llvm.sideeffect
// LLVM-ASM-AMD64-NOT: require-anchor
// LLVM-ASM-AMD64-NOT: VPS{{LL|RL|RA}}
// LLVM-NM-AMD64: D codegen.llvmSIMDShiftFoldedProducer.goallc.fmv.slot

//go:noinline
func llvmSIMDShiftFoldedProducer(x, y archsimd.Int64x2) archsimd.Int64x2 {
	if archsimd.X86.AVX() {
		sum := x.Add(y)
		if archsimd.X86.AVX512() {
			return sum.ShiftAllRight(0)
		}
		return sum
	}
	return x
}
