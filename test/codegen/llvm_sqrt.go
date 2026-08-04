// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "math"

// LLVM-DAG: define goabiinternal double @codegen.llvmSqrt64(double %x)
// LLVM-DAG: call double @llvm.sqrt.f64(double %x)
// LLVM-OPT-DAG: define goabiinternal double @codegen.llvmSqrt64(double %x)
// LLVM-OPT-DAG: call double @llvm.sqrt.f64(double %x)
func llvmSqrt64(x float64) float64 {
	return math.Sqrt(x)
}

// LLVM-DAG: define goabiinternal float @codegen.llvmSqrt32(float %x)
// LLVM-DAG: call float @llvm.sqrt.f32(float %x)
// LLVM-OPT-DAG: define goabiinternal float @codegen.llvmSqrt32(float %x)
// LLVM-OPT-DAG: call float @llvm.sqrt.f32(float %x)
func llvmSqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}
