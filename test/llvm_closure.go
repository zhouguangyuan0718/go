// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

//go:noinline
func makePair(base int) func(int) (int, int) {
	return func(x int) (int, int) {
		return base + x, base - x
	}
}

//go:noinline
func applyPair(f func(int) (int, int), x int) (int, int) {
	return f(x)
}

//go:noinline
func applyPairTwice(f func(int) (int, int), x int) (int, int) {
	a, b := f(x)
	c, d := f(x + 1)
	return a + c, b + d
}

// makeDeepPair recursively invokes the closure until the closure itself takes
// the GoObj morestack slow path. The hidden funcval context must survive stack
// growth in REGCTXT.
//
//go:noinline
func makeDeepPair(base int) func(int) (int, int) {
	var f func(int) (int, int)
	f = func(depth int) (int, int) {
		if depth == 0 {
			return base, base
		}
		a, b := f(depth - 1)
		return a + 1, b - 1
	}
	return f
}

type closureStackValue struct {
	first  *int
	scalar int
	second *int
}

type stackValueClosure func(
	a0, a1, a2, a3, a4, a5, a6, a7 int,
	a8, a9, a10, a11, a12, a13, a14, a15 int,
	value closureStackValue,
) (*int, *int, int)

// makeStackValueClosure exercises the indirect-call ABI path. The sixteen
// integer arguments exhaust the arm64 integer register budget, so the
// pointer-containing aggregate is passed as a typed byval stack argument.
//
//go:noinline
func makeStackValueClosure(bias int) stackValueClosure {
	return func(
		a0, a1, a2, a3, a4, a5, a6, a7 int,
		a8, a9, a10, a11, a12, a13, a14, a15 int,
		value closureStackValue,
	) (*int, *int, int) {
		runtime.GC()
		return value.first, value.second, bias + a0 + a15 + value.scalar
	}
}

func main() {
	f := makePair(40)
	a, b := applyPair(f, 2)
	if a != 42 || b != 38 {
		panic("bad closure result")
	}

	a, b = applyPairTwice(f, 2)
	if a != 85 || b != 75 {
		panic("bad repeated closure result")
	}

	f = makeDeepPair(10040)
	a, b = applyPair(f, 10000)
	if a != 20040 || b != 40 {
		panic("closure context lost across stack growth")
	}

	first, second := 11, 37
	stackCall := makeStackValueClosure(3)
	gotFirst, gotSecond, checksum := stackCall(
		1, 2, 3, 4, 5, 6, 7, 8,
		9, 10, 11, 12, 13, 14, 15, 16,
		closureStackValue{first: &first, scalar: 23, second: &second},
	)
	runtime.GC()
	if gotFirst != &first || *gotFirst != 11 ||
		gotSecond != &second || *gotSecond != 37 ||
		checksum != 43 {
		panic("byval closure stack argument lost across GC")
	}
}
