// asmcheck

//go:build goexperiment.simd && amd64

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "simd/archsimd"

// Ordinary shifts use full-width unsigned counts and standard vector IR.
// In particular, the scalar count must be checked before narrowing it to
// the data lane width, and signed right shifts must preserve the sign.
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllLeft128(
// LLVM-AMD64-DAG: icmp ult i64 {{.*}}, 16
// LLVM-AMD64-DAG: trunc i64 {{.*}} to i16
// LLVM-AMD64-DAG: shl <8 x i16>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllRightSigned128(
// LLVM-AMD64-DAG: icmp ult i64 {{.*}}, 64
// LLVM-AMD64-DAG: ashr <2 x i64>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllRightUnsigned256(
// LLVM-AMD64-DAG: lshr <8 x i32>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftLeftWords128(
// LLVM-AMD64-DAG: icmp ult <8 x i16>
// LLVM-AMD64-DAG: select <8 x i1>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftLeft256(
// LLVM-AMD64-DAG: icmp ult <8 x i32>
// LLVM-AMD64-DAG: shl <8 x i32>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftRightSigned256(
// LLVM-AMD64-DAG: ashr <8 x i32>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftRightUnsigned512(
// LLVM-AMD64-DAG: icmp ult <8 x i64>
// LLVM-AMD64-DAG: lshr <8 x i64>
// LLVM-AMD64-DAG: define {{.*}} @codegen.llvmSIMDShiftRightSigned128(

// Word variable shifts and signed qword right shifts require AVX512 even
// with 128-bit values. The 256-bit dword cases require AVX2 above their AVX
// ABI floor. Standard IR must retain those source capability requirements.
// LLVM-AMD64-DAG: !goallc.cpu.guard !{{[0-9]+}}
// LLVM-AMD64-DAG: !goallc.cpu.requires !{{[0-9]+}}
// LLVM-AMD64-DAG: "goallc.cpu.multiversion"="x86.avx2"
// LLVM-AMD64-DAG: "goallc.cpu.multiversion"="x86.avx512"
// LLVM-AMD64-DAG: !{!"x86.avx2"}
// LLVM-AMD64-DAG: !{!"x86.avx512"}
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftLeftWords128<goallc.fmv.baseline>"
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftLeftWords128<goallc.fmv.avx512>"
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftLeftWords128<goallc.fmv.resolve>"
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftAllRightSigned128<goallc.fmv.avx512>"
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftRightSigned128<goallc.fmv.avx512>"
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftLeft256<goallc.fmv.avx2>"
// LLVM-OPT-AMD64-DAG: define internal {{.*}} @"codegen.llvmSIMDShiftRightSigned256<goallc.fmv.avx2>"
// LLVM-NM-AMD64-DAG: D codegen.llvmSIMDShiftLeftWords128.goallc.fmv.slot
// LLVM-NM-AMD64-DAG: D codegen.llvmSIMDShiftAllRightSigned128.goallc.fmv.slot
// LLVM-NM-AMD64-DAG: D codegen.llvmSIMDShiftRightSigned128.goallc.fmv.slot
// LLVM-NM-AMD64-DAG: D codegen.llvmSIMDShiftLeft256.goallc.fmv.slot
// LLVM-ASM-AMD64-DAG: VPSLLW
// LLVM-ASM-AMD64-DAG: VPSLLVW
// LLVM-ASM-AMD64-DAG: VPSLLVD
// LLVM-ASM-AMD64-DAG: VPSRLD
// LLVM-ASM-AMD64-DAG: VPSRLVQ
// LLVM-ASM-AMD64-DAG: VPSRAVD
// LLVM-ASM-AMD64-DAG: VPSRAQ
// LLVM-ASM-AMD64-DAG: VPSRAVQ

//go:noinline
func llvmSIMDShiftAllLeft128(x archsimd.Uint16x8, count uint64) archsimd.Uint16x8 {
	if !archsimd.X86.AVX() {
		return x
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func llvmSIMDShiftAllRightSigned128(x archsimd.Int64x2, count uint64) archsimd.Int64x2 {
	if !archsimd.X86.AVX512() {
		return x
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func llvmSIMDShiftAllRightUnsigned256(x archsimd.Uint32x8, count uint64) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return x
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func llvmSIMDShiftLeftWords128(x, count archsimd.Uint16x8) archsimd.Uint16x8 {
	if !archsimd.X86.AVX512() {
		return x
	}
	return x.ShiftLeft(count)
}

//go:noinline
func llvmSIMDShiftLeft256(x, count archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return x
	}
	return x.ShiftLeft(count)
}

//go:noinline
func llvmSIMDShiftRightSigned256(x archsimd.Int32x8, count archsimd.Uint32x8) archsimd.Int32x8 {
	if !archsimd.X86.AVX2() {
		return x
	}
	return x.ShiftRight(count)
}

//go:noinline
func llvmSIMDShiftRightUnsigned512(x, count archsimd.Uint64x8) archsimd.Uint64x8 {
	if !archsimd.X86.AVX512() {
		return x
	}
	return x.ShiftRight(count)
}

//go:noinline
func llvmSIMDShiftRightSigned128(x archsimd.Int64x2, count archsimd.Uint64x2) archsimd.Int64x2 {
	if !archsimd.X86.AVX512() {
		return x
	}
	return x.ShiftRight(count)
}
