// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal { ptr, i64, i64 } @codegen.llvmRuntimeMakesliceInt(i64
// LLVM: call goabiinternal ptr @runtime.makeslice(ptr {{.*}}, i64 {{.*}}, i64 {{.*}})
// LLVM: ret { ptr, i64, i64 }
// LLVM: declare goabiinternal ptr @runtime.makeslice(ptr, i64, i64)

// LLVM-LABEL: define goabiinternal { ptr, i64, i64 } @codegen.llvmRuntimeMakeslice(i64
// LLVM: call goabiinternal ptr @runtime.makeslice(ptr {{.*}}, i64 {{.*}}, i64 {{.*}})
// LLVM: ret { ptr, i64, i64 }

// LLVM-LABEL: define goabiinternal ptr @codegen.llvmRuntimeNewobject(
// LLVM: call goabiinternal ptr @runtime.newobject(ptr
// LLVM: ret ptr
// LLVM: declare goabiinternal ptr @runtime.newobject(ptr)

// LLVM-LABEL: define goabiinternal void @codegen.llvmNewproc(
// LLVM: call goabiinternal void @runtime.newproc(ptr
// LLVM-NOT: call goabiinternal void @runtime.newproc(i64
// LLVM: ret void
// LLVM: declare goabiinternal void @runtime.newproc(ptr)
var llvmNewprocSink int

func llvmNewproc(value int) {
	go func() { llvmNewprocSink = value }()
}

func llvmRuntimeNewobject() *[128]byte {
	return new([128]byte)
}

func llvmRuntimeMakeslice(n int) []byte {
	return make([]byte, n)
}

func llvmRuntimeMakesliceInt(n int) []int {
	return make([]int, n)
}
