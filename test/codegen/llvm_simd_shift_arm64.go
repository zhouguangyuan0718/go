// asmcheck

//go:build goexperiment.simd && arm64

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "simd/archsimd"

// These scalar-count APIs use all 64 count bits. They must not inherit the
// signed-low-byte semantics of the distinct arm64 Shift API.
// LLVM-ARM64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllLeft8(
// LLVM-ARM64-DAG: icmp ult i64 {{.*}}, 8
// LLVM-ARM64-DAG: trunc i64 {{.*}} to i8
// LLVM-ARM64-DAG: shl <16 x i8>
// LLVM-ARM64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllRightSigned8(
// LLVM-ARM64-DAG: ashr <16 x i8>
// LLVM-ARM64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllLeft16(
// LLVM-ARM64-DAG: icmp ult i64 {{.*}}, 16
// LLVM-ARM64-DAG: shl <8 x i16>
// LLVM-ARM64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllRightUnsigned32(
// LLVM-ARM64-DAG: icmp ult i64 {{.*}}, 32
// LLVM-ARM64-DAG: lshr <4 x i32>
// LLVM-ARM64-DAG: define {{.*}} @codegen.llvmSIMDShiftAllRightSigned64(
// LLVM-ARM64-DAG: icmp ult i64 {{.*}}, 64
// LLVM-ARM64-DAG: ashr <2 x i64>
// LLVM-OPT-ARM64-DAG: shl <16 x i8>
// LLVM-OPT-ARM64-DAG: ashr <16 x i8>
// LLVM-OPT-ARM64-DAG: lshr <4 x i32>
// LLVM-OPT-ARM64-DAG: ashr <2 x i64>
// LLVM-ASM-ARM64-DAG: VUSHL
// LLVM-ASM-ARM64-DAG: VSSHL

//go:noinline
func llvmSIMDShiftAllLeft8(x archsimd.Uint8x16, count uint64) archsimd.Uint8x16 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func llvmSIMDShiftAllRightSigned8(x archsimd.Int8x16, count uint64) archsimd.Int8x16 {
	return x.ShiftAllRight(count)
}

//go:noinline
func llvmSIMDShiftAllLeft16(x archsimd.Int16x8, count uint64) archsimd.Int16x8 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func llvmSIMDShiftAllRightUnsigned32(x archsimd.Uint32x4, count uint64) archsimd.Uint32x4 {
	return x.ShiftAllRight(count)
}

//go:noinline
func llvmSIMDShiftAllRightSigned64(x archsimd.Int64x2, count uint64) archsimd.Int64x2 {
	return x.ShiftAllRight(count)
}
