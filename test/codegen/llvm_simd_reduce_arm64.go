// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && arm64

package codegen

import "simd/archsimd"

// LLVM-ARM64-DAG: call i8 @llvm.vector.reduce.add.v16i8
// LLVM-ASM-ARM64-DAG: VADDV
//
//go:noinline
func llvmSIMDReduceSumInt8(x archsimd.Int8x16) int8 {
	return x.ReduceSum()
}

// LLVM-ARM64-DAG: call i16 @llvm.vector.reduce.smax.v8i16
// LLVM-ASM-ARM64-DAG: VSMAXV
//
//go:noinline
func llvmSIMDReduceMaxInt16(x archsimd.Int16x8) int16 {
	return x.ReduceMax()
}

// LLVM-ARM64-DAG: call i32 @llvm.vector.reduce.umax.v4i32
// LLVM-ASM-ARM64-DAG: VUMAXV
//
//go:noinline
func llvmSIMDReduceMaxUint32(x archsimd.Uint32x4) uint32 {
	return x.ReduceMax()
}

// LLVM-ARM64-DAG: call i32 @llvm.vector.reduce.smin.v4i32
// LLVM-ASM-ARM64-DAG: VSMINV
//
//go:noinline
func llvmSIMDReduceMinInt32(x archsimd.Int32x4) int32 {
	return x.ReduceMin()
}

// LLVM-ARM64-DAG: call i16 @llvm.vector.reduce.umin.v8i16
// LLVM-ASM-ARM64-DAG: VUMINV
//
//go:noinline
func llvmSIMDReduceMinUint16(x archsimd.Uint16x8) uint16 {
	return x.ReduceMin()
}

// LLVM-ARM64-DAG: call float @llvm.vector.reduce.fmaximum.v4f32
// LLVM-ASM-ARM64-DAG: FMAXV
//
//go:noinline
func llvmSIMDReduceMaxFloat32(x archsimd.Float32x4) float32 {
	return x.ReduceMax()
}

// LLVM-ARM64-DAG: call float @llvm.vector.reduce.fminimum.v4f32
// LLVM-ASM-ARM64-DAG: FMINV
//
//go:noinline
func llvmSIMDReduceMinFloat32(x archsimd.Float32x4) float32 {
	return x.ReduceMin()
}
