// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

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

func main() {
	f := makePair(40)
	a, b := applyPair(f, 2)
	if a != 42 || b != 38 {
		panic("bad closure result")
	}

	f = makeDeepPair(10040)
	a, b = applyPair(f, 10000)
	if a != 20040 || b != 40 {
		panic("closure context lost across stack growth")
	}
}
