// asmcheck -gcflags=-d=ssa/check/on

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "math"

// LLVM-DAG: define goabiinternal double @codegen.llvmTrunc64(double %x)
// LLVM-DAG: call double @llvm.trunc.f64(double %x)
// LLVM-OPT-DAG: define goabiinternal double @codegen.llvmTrunc64(double %x)
// LLVM-OPT-DAG: call double @llvm.trunc.f64(double %x)
func llvmTrunc64(x float64) float64 {
	return math.Trunc(x)
}

// LLVM-DAG: define goabiinternal double @codegen.llvmCeil64(double %x)
// LLVM-DAG: call double @llvm.ceil.f64(double %x)
func llvmCeil64(x float64) float64 {
	return math.Ceil(x)
}

// LLVM-DAG: define goabiinternal double @codegen.llvmFloor64(double %x)
// LLVM-DAG: call double @llvm.floor.f64(double %x)
func llvmFloor64(x float64) float64 {
	return math.Floor(x)
}

// LLVM-DAG: define goabiinternal double @codegen.llvmRound64(double %x)
// LLVM-DAG: call double @llvm.round.f64(double %x)
func llvmRound64(x float64) float64 {
	return math.Round(x)
}

// LLVM-DAG: define goabiinternal double @codegen.llvmRoundToEven64(double %x)
// LLVM-DAG: call double @llvm.roundeven.f64(double %x)
func llvmRoundToEven64(x float64) float64 {
	return math.RoundToEven(x)
}
