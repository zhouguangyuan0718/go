// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-DAG: call i32 @llvm.fptosi.sat.i32.f32(float %x)
// LLVM-DAG: call i64 @llvm.fptosi.sat.i64.f64(double %x)
// LLVM-OPT-DAG: call i32 @llvm.fptosi.sat.i32.f32(float %x)
// LLVM-OPT-DAG: call i64 @llvm.fptosi.sat.i64.f64(double %x)

func llvmFloat32ToInt32(x float32) int32 {
	return int32(x)
}

func llvmFloat64ToInt64(x float64) int64 {
	return int64(x)
}

func llvmFloat32ToUint32(x float32) uint32 {
	return uint32(x)
}

func llvmFloat64ToUint64(x float64) uint64 {
	return uint64(x)
}
