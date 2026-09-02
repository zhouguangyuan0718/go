// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && (amd64 || arm64)

package codegen

import "simd/archsimd"

// LLVM-AMD64-DAG: extractelement <4 x i32> {{.*}}, i32 2
// LLVM-ARM64-DAG: extractelement <4 x i32> {{.*}}, i32 2
// LLVM-ASM-AMD64-LABEL: TEXT codegen.llvmSIMDGetInt32(SB)
// LLVM-ASM-AMD64: V{{(EXTRACTPS|PEXTRD)}}
// LLVM-ASM-ARM64-LABEL: TEXT codegen.llvmSIMDGetInt32(SB)
// LLVM-ASM-ARM64: VMOV V0.S[2], R0
//
//go:noinline
func llvmSIMDGetInt32(x archsimd.Int32x4) int32 {
	return x.GetElem(2)
}

// LLVM-AMD64-DAG: insertelement <4 x i32> {{.*}}, i32 {{.*}}, i32 1
// LLVM-ARM64-DAG: insertelement <4 x i32> {{.*}}, i32 {{.*}}, i32 1
// LLVM-ASM-AMD64-LABEL: TEXT codegen.llvmSIMDSetInt32(SB)
// LLVM-ASM-AMD64: VPINSRD
// LLVM-ASM-ARM64-LABEL: TEXT codegen.llvmSIMDSetInt32(SB)
// LLVM-ASM-ARM64: VMOV R0, V0.S[1]
//
//go:noinline
func llvmSIMDSetInt32(x archsimd.Int32x4, value int32) archsimd.Int32x4 {
	return x.SetElem(1, value)
}

// LLVM-AMD64-DAG: extractelement <2 x double> {{.*}}, i32 1
// LLVM-ARM64-DAG: extractelement <2 x double> {{.*}}, i32 1
//
//go:noinline
func llvmSIMDGetFloat64(x archsimd.Float64x2) float64 {
	return x.GetElem(1)
}

// LLVM-AMD64-DAG: insertelement <2 x double> {{.*}}, double {{.*}}, i32 0
// LLVM-ARM64-DAG: insertelement <2 x double> {{.*}}, double {{.*}}, i32 0
//
//go:noinline
func llvmSIMDSetFloat64(x archsimd.Float64x2, value float64) archsimd.Float64x2 {
	return x.SetElem(0, value)
}
