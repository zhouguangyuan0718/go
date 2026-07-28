// asmcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// These tests check that allocating a 0-size object does not
// introduce a call to runtime.newobject.

package codegen

// LLVM-DAG: @runtime.zerobase = external global i8
// LLVM-DAG: define goabiinternal ptr @codegen.zeroAllocNew1()
// LLVM-DAG: ret ptr @runtime.zerobase
// LLVM-DAG: define goabiinternal ptr @codegen.zeroAllocNew2()
// LLVM-DAG: define goabiinternal { ptr, i64, i64 } @codegen.zeroAllocSliceLit()
// LLVM-DAG: ret { ptr, i64, i64 } { ptr @runtime.zerobase, i64 0, i64 0 }

func zeroAllocNew1() *struct{} {
	// 386:-`CALL runtime\.newobject` `LEAL runtime.zerobase`
	// amd64:-`CALL runtime\.newobject` `LEAQ runtime.zerobase`
	// arm:-`CALL runtime\.newobject` `MOVW [$]runtime.zerobase`
	// arm64:-`CALL runtime\.newobject` `MOVD [$]runtime.zerobase`
	// riscv64:-`CALL runtime\.newobject` `MOV [$]runtime.zerobase`
	return new(struct{})
}

func zeroAllocNew2() *[0]int {
	// 386:-`CALL runtime\.newobject` `LEAL runtime.zerobase`
	// amd64:-`CALL runtime\.newobject` `LEAQ runtime.zerobase`
	// arm:-`CALL runtime\.newobject` `MOVW [$]runtime.zerobase`
	// arm64:-`CALL runtime\.newobject` `MOVD [$]runtime.zerobase`
	// riscv64:-`CALL runtime\.newobject` `MOV [$]runtime.zerobase`
	return new([0]int)
}

func zeroAllocSliceLit() []int {
	// 386:-`CALL runtime\.newobject` `LEAL runtime.zerobase`
	// amd64:-`CALL runtime\.newobject` `LEAQ runtime.zerobase`
	// arm:-`CALL runtime\.newobject` `MOVW [$]runtime.zerobase`
	// arm64:-`CALL runtime\.newobject` `MOVD [$]runtime.zerobase`
	// riscv64:-`CALL runtime\.newobject` `MOV [$]runtime.zerobase`
	return []int{}
}
