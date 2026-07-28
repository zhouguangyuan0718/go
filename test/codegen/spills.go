// asmcheck

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-DAG: define goabiinternal float @codegen.f32(float %a, float %b)
// LLVM-DAG: fadd float %a, %b
// LLVM-DAG: define goabiinternal double @codegen.f64(double %a, double %b)
// LLVM-DAG: fadd double %a, %b
// LLVM-DAG: define goabiinternal i32 @codegen.i32(i32 %a, i32 %b)
// LLVM-DAG: add i32 %a, %b
// LLVM-DAG: define goabiinternal i64 @codegen.i64(i64 %a, i64 %b)
// LLVM-DAG: add i64 %a, %b

func i64(a, b int64) int64 { // arm64:`STP ` `LDP `
	g()
	return a + b
}

func i32(a, b int32) int32 { // arm64:`STPW` `LDPW`
	g()
	return a + b
}

func f64(a, b float64) float64 { // arm64:`FSTPD` `FLDPD`
	g()
	return a + b
}

func f32(a, b float32) float32 { // arm64:`FSTPS` `FLDPS`
	g()
	return a + b
}

//go:noinline
func g() {
}
