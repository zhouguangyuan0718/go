// compile

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is compiled directly by the LLVM ArgsPointerMaps differential
// test. Keep the native Go and GoALLC inputs identical.
package p

type pointerAggregate struct {
	first  *int
	scalar uintptr
	second *int
}

//go:noescape
func safepoint()

//go:noinline
func initializedPointerResult(pointer *int) (
	r0, r1, r2, r3, r4, r5, r6, r7 int,
	r8, r9, r10, r11, r12, r13, r14, r15 int,
	result *int,
) {
	result = pointer
	safepoint()
	return
}

//go:noinline
func partiallyInitializedAggregateResult(first, second *int) (
	r0, r1, r2, r3, r4, r5, r6, r7 int,
	r8, r9, r10, r11, r12, r13, r14 int,
	result pointerAggregate,
) {
	result.first = first
	safepoint()
	result.second = second
	safepoint()
	return
}
