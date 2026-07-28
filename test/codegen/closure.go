// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal { i64, i64 } @codegen.callLLVMPairClosureTwice(
// LLVM-NOT: alloca
// LLVM-NOT: load volatile i8
// LLVM: load ptr, ptr %f, align 8
// LLVM: call goabiinternal { i64, i64 } {{%.*}}(i64 %x, ptr nest %f)
// LLVM-NOT: load volatile i8
// LLVM: load ptr, ptr %f, align 8
// LLVM: call goabiinternal { i64, i64 } {{%.*}}(i64 {{%.*}}, ptr nest %f)
// LLVM-NOT: load volatile i8

// LLVM-OPT-LABEL: define goabiinternal { i64, i64 } @codegen.callLLVMPairClosureTwice(
// LLVM-OPT-NOT: load volatile
// LLVM-OPT: load ptr, ptr %f, align 8
// LLVM-OPT: tail call goabiinternal { i64, i64 } {{%.*}}(i64 %x, ptr nest %f)
// LLVM-OPT-NOT: load volatile
// LLVM-OPT: load ptr, ptr %f, align 8
// LLVM-OPT: tail call goabiinternal { i64, i64 } {{%.*}}(i64 {{%.*}}, ptr nest %f)
// LLVM-OPT-NOT: load volatile

// LLVM-LABEL: define goabiinternal { i64, i64 } @codegen.callLLVMPairClosure(
// LLVM-NOT: alloca
// LLVM-NOT: load volatile i8
// LLVM: {{%.*}} = load ptr, ptr %f, align 8
// LLVM-NEXT: {{%.*}} = call goabiinternal { i64, i64 } {{%.*}}(i64 %x, ptr nest %f)

// LLVM-OPT-LABEL: define goabiinternal { i64, i64 } @codegen.callLLVMPairClosure(
// LLVM-OPT-NOT: load volatile
// LLVM-OPT: load ptr, ptr %f, align 8
// LLVM-OPT-NEXT: {{%.*}} = tail call goabiinternal { i64, i64 } {{%.*}}(i64 %x, ptr nest %f)

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

//go:noinline
func callLLVMNilClosure() {
	var f func()
	f()
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

// LLVM-LABEL: define goabiinternal void @codegen.callLLVMNilClosure(
// LLVM-NOT: load volatile
// LLVM: load ptr, ptr null, align 8
// LLVM-NEXT: call goabiinternal void {{%.*}}(ptr nest null)
// LLVM-OPT-LABEL: define goabiinternal void @codegen.callLLVMNilClosure(
// LLVM-OPT-NOT: unreachable
// LLVM-OPT-NOT: load volatile
// LLVM-OPT: load ptr, ptr null, align {{[0-9]+}}
// LLVM-OPT-NEXT: tail call goabiinternal void {{%.*}}(ptr nest null)

// LLVM: store ptr @codegen.makeLLVMPairClosure.func1, ptr {{%.*}}, align 8
// LLVM: define goabiinternal { i64, i64 } @codegen.makeLLVMPairClosure.func1(i64 %x, ptr nest %.closureptr)
// LLVM-DAG: attributes #{{[0-9]+}} = { {{.*}}null_pointer_is_valid{{.*}} }
// LLVM-OPT-DAG: attributes #{{[0-9]+}} = { {{.*}}null_pointer_is_valid{{.*}} }
