// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && arm64

package codegen

import "simd/archsimd"

// LLVM-ARM64-DAG: call <4 x i32> @llvm.sadd.sat.v4i32(
// LLVM-ARM64-DAG: call <4 x i32> @llvm.ssub.sat.v4i32(
//
//go:noinline
func llvmSIMDSaturatedInt32(x, y archsimd.Int32x4) (archsimd.Int32x4, archsimd.Int32x4) {
	return x.AddSaturated(y), x.SubSaturated(y)
}

// LLVM-ARM64-DAG: call <2 x i64> @llvm.uadd.sat.v2i64(
// LLVM-ARM64-DAG: call <2 x i64> @llvm.usub.sat.v2i64(
//
//go:noinline
func llvmSIMDSaturatedUint64(x, y archsimd.Uint64x2) (archsimd.Uint64x2, archsimd.Uint64x2) {
	return x.AddSaturated(y), x.SubSaturated(y)
}
