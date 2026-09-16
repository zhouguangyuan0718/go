// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"runtime"
	"unsafe"
)

//go:linkname clearPointers runtime.memclrHasPointers
func clearPointers(p unsafe.Pointer, n uintptr)

//go:noinline
func checkPanic(f func()) {
	deferred, returned := false, false
	func() {
		defer func() {
			if recover() == nil {
				panic("missing recovery")
			}
			deferred = true
		}()
		f()
		returned = true
	}()
	if returned {
		panic("panic helper returned")
	}
	if !deferred {
		panic("lost defer side effect")
	}
}

//go:noinline
func divide(n int) { _ = 1 / n }

//go:noinline
func shift(n int) { _ = 1 << n }

//go:noinline
func bounds(n int) { _ = [1]int{}[n] }

//go:noinline
func makeSlice(n int) { _ = make([]int, n) }

//go:noinline
func assertType(x any) { _ = x.(int) }

//go:noinline
func equal(a, b any) bool { return a == b }

func main() {
	checkPanic(func() { divide(0) })
	checkPanic(func() { shift(-1) })
	checkPanic(func() { bounds(2) })
	checkPanic(func() { makeSlice(-1) })
	checkPanic(func() { assertType("wrong type") })
	checkPanic(func() { panic("explicit") })

	for _, pair := range [][2]any{
		{int8(1), int8(2)}, {int16(1), int16(2)}, {int32(1), int32(2)},
		{int64(1), int64(2)}, {[2]int64{1, 2}, [2]int64{1, 3}},
		{float32(1), float32(2)}, {float64(1), float64(2)},
		{complex64(1 + 2i), complex64(1 + 3i)}, {complex128(1 + 2i), complex128(1 + 3i)},
		{"same string", "different string"},
	} {
		if !equal(pair[0], pair[0]) || equal(pair[0], pair[1]) {
			panic("scalar equality")
		}
		m := map[any]int{pair[0]: 7, pair[1]: 9}
		if m[pair[0]] != 7 || m[pair[1]] != 9 {
			panic("scalar hash")
		}
	}
	zero, negzero, nan := 0.0, math.Copysign(0, -1), math.NaN()
	if !equal(zero, negzero) || equal(nan, nan) {
		panic("floating equality")
	}
	m := map[float64]int{zero: 3}
	if m[negzero] != 3 {
		panic("signed zero hash")
	}
	m[nan], m[nan] = 5, 6
	if len(m) != 3 {
		panic("NaN hash entries")
	}
	for range 100 {
		m[nan] = 7
	}
	if len(m) != 103 {
		panic("NaN hash state")
	}

	var nilch chan int
	ch := make(chan int, 3)
	ch <- 1
	if len(nilch) != 0 || cap(nilch) != 0 || len(ch) != 1 || cap(ch) != 3 {
		panic("channel query")
	}

	// Exercise barrier clearing while another goroutine repeatedly starts GC.
	done := make(chan struct{})
	go func() {
		for range 10 {
			runtime.GC()
		}
		close(done)
	}()
	for range 100 {
		p := new([128]*int)
		for i := range p {
			p[i] = new(int)
			*p[i] = i
		}
		clearPointers(unsafe.Pointer(&p[0]), unsafe.Sizeof(*p))
		for _, v := range p {
			if v != nil {
				panic("pointer clear")
			}
		}
	}
	<-done
}
