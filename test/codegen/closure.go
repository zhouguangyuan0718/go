// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal { i64, i64 } @codegen.callLLVMPairClosureTwice(
// LLVM-NOT: alloca
// LLVM: load volatile i8, ptr %f, align 1
// LLVM: call goabiinternal { i64, i64 } {{%.*}}(i64 %x, ptr nest %f)
// LLVM: load volatile i8, ptr %f, align 1
// LLVM: call goabiinternal { i64, i64 } {{%.*}}(i64 {{%.*}}, ptr nest %f)

// LLVM-LABEL: define goabiinternal { i64, i64 } @codegen.callLLVMPairClosure(
// LLVM-NOT: alloca
// LLVM: load volatile i8, ptr %f, align 1
// LLVM-NEXT: {{%.*}} = load ptr, ptr %f, align 8
// LLVM-NEXT: {{%.*}} = call goabiinternal { i64, i64 } {{%.*}}(i64 %x, ptr nest %f)

//go:noinline
func makeLLVMPairClosure(base int) func(int) (int, int) {
	return func(x int) (int, int) {
		return base + x, base - x
	}
}

//go:noinline
func callLLVMPairClosure(f func(int) (int, int), x int) (int, int) {
	return f(x)
}

//go:noinline
func callLLVMPairClosureTwice(f func(int) (int, int), x int) (int, int) {
	a, b := f(x)
	c, d := f(x + 1)
	return a + c, b + d
}

// LLVM-LABEL: define goabiinternal void @codegen.llvmAllocaLoop(
// LLVM: alloca i64, align 8
// LLVM: br label
// LLVM-NOT: alloca
// LLVM: ret void

//go:noinline
func llvmAllocaLoop(n int) {
	for i := 0; i < n; i++ {
		x := i
		llvmTouchLocal(&x)
	}
}

//go:noinline
func llvmTouchLocal(p *int) {
	*p++
}

// LLVM: store ptr @codegen.makeLLVMPairClosure.func1, ptr {{%.*}}, align 8
// LLVM: define goabiinternal { i64, i64 } @codegen.makeLLVMPairClosure.func1(i64 %x, ptr nest %.closureptr)
