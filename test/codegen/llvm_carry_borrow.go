// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "math/bits"

// LLVM-DAG: call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %x, i64 %y)
// LLVM-DAG: call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %{{.*}}, i64 %carry)
// LLVM-DAG: or i1
// LLVM-DAG: zext i1 %{{.*}} to i64
func llvmAdd64Carry(x, y, carry uint64) (sum, carryOut uint64) {
	return bits.Add64(x, y, carry)
}

// LLVM-DAG: call { i64, i1 } @llvm.usub.with.overflow.i64(i64 %x, i64 %y)
// LLVM-DAG: call { i64, i1 } @llvm.usub.with.overflow.i64(i64 %{{.*}}, i64 %borrow)
// LLVM-DAG: or i1
// LLVM-DAG: zext i1 %{{.*}} to i64
func llvmSub64Borrow(x, y, borrow uint64) (difference, borrowOut uint64) {
	return bits.Sub64(x, y, borrow)
}
