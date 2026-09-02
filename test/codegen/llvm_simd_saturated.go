// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && (amd64 || arm64)

package codegen

import "simd/archsimd"

// LLVM-AMD64-DAG: define goabiinternal { <16 x i8>, <16 x i8> } @codegen.llvmSIMDSaturatedSigned(
// LLVM-AMD64-DAG: call <16 x i8> @llvm.sadd.sat.v16i8(
// LLVM-AMD64-DAG: call <16 x i8> @llvm.ssub.sat.v16i8(
// LLVM-ARM64-DAG: define goabiinternal { <16 x i8>, <16 x i8> } @codegen.llvmSIMDSaturatedSigned(
// LLVM-ARM64-DAG: call <16 x i8> @llvm.sadd.sat.v16i8(
// LLVM-ARM64-DAG: call <16 x i8> @llvm.ssub.sat.v16i8(
//
//go:noinline
func llvmSIMDSaturatedSigned(x, y archsimd.Int8x16) (archsimd.Int8x16, archsimd.Int8x16) {
	return x.AddSaturated(y), x.SubSaturated(y)
}

// LLVM-AMD64-DAG: define goabiinternal { <8 x i16>, <8 x i16> } @codegen.llvmSIMDSaturatedUnsigned(
// LLVM-AMD64-DAG: call <8 x i16> @llvm.uadd.sat.v8i16(
// LLVM-AMD64-DAG: call <8 x i16> @llvm.usub.sat.v8i16(
// LLVM-ARM64-DAG: define goabiinternal { <8 x i16>, <8 x i16> } @codegen.llvmSIMDSaturatedUnsigned(
// LLVM-ARM64-DAG: call <8 x i16> @llvm.uadd.sat.v8i16(
// LLVM-ARM64-DAG: call <8 x i16> @llvm.usub.sat.v8i16(
//
//go:noinline
func llvmSIMDSaturatedUnsigned(x, y archsimd.Uint16x8) (archsimd.Uint16x8, archsimd.Uint16x8) {
	return x.AddSaturated(y), x.SubSaturated(y)
}
