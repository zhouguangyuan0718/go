// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
	"unsafe"
)

//go:linkname allocate runtime.mallocgc
func allocate(size uintptr, typ unsafe.Pointer, zero bool) unsafe.Pointer

//go:noinline
func box16(x uint16) any { return x }

//go:noinline
func box32(x uint32) any { return x }

//go:noinline
func box64(x uint64) any { return x }

//go:noinline
func boxString(x string) any { return x }

//go:noinline
func boxSlice(x []byte) any { return x }

//go:noinline
func makeBytes(n int) []byte { return make([]byte, n) }

//go:noinline
func constantTinyPair() {
	a := (*byte)(allocate(1, nil, true))
	b := (*byte)(allocate(1, nil, true))
	*a = 17
	*b = 29
	runtime.GC()
	if a == b || *a != 17 || *b != 29 {
		panic("tiny objects alias")
	}
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
}

//go:noinline
func constantZeroPair() {
	// These raw runtime calls currently return the same zerobase. Do not
	// change that concrete IR contract while modeling positive allocations.
	a := allocate(0, nil, true)
	b := allocate(0, nil, true)
	if a != b {
		panic("zero allocation identity")
	}
}

//go:noinline
func constantZeroRead() uint64 { return *(*uint64)(allocate(8, nil, true)) }

//go:noinline
func newTiny() *byte { return new(byte) }

//go:noinline
func newNoScan() *[80]byte { return new([80]byte) }

//go:noinline
func newScan() *[10]*int { return new([10]*int) }

//go:noinline
func newLarge() *[128]byte { return new([128]byte) }

func sourceAllocations() {
	a, b := newTiny(), newTiny()
	c, d, e := newNoScan(), newScan(), newLarge()
	if *a != 0 || *b != 0 || *c != [80]byte{} || *d != [10]*int{} || *e != [128]byte{} {
		panic("source allocation not zeroed")
	}
	*a, *b, c[79], e[127] = 11, 13, 17, 19
	x := 23
	d[9] = &x
	runtime.GC()
	if a == b || *a != 11 || *b != 13 || c[79] != 17 || *d[9] != 23 || e[127] != 19 {
		panic("source allocation corrupted")
	}
}

func main() {
	sourceAllocations()
	constantTinyPair()
	constantZeroPair()
	if constantZeroRead() != 0 {
		panic("zero allocation contents")
	}
	// Include zerobase, adjacent tiny allocations and allocations beyond the
	// small-object classes. Live returned pointers must survive a safe point.
	for _, n := range []uintptr{0, 1, 2, 3, 7, 16, 1024, 40000} {
		for _, zero := range []bool{false, true} {
			p := allocate(n, nil, zero)
			if p == nil {
				panic("nil allocation")
			}
			b := unsafe.Slice((*byte)(p), int(n))
			for i := range b {
				if zero && b[i] != 0 {
					panic("allocation not zeroed")
				}
				b[i] = byte(i + 17)
			}
			runtime.GC()
			for i, v := range b {
				if v != byte(i+17) {
					panic("allocation lost across GC")
				}
			}
			runtime.KeepAlive(p)
		}
	}
	// Cover both cached and freshly allocated boxes, including empty payloads.
	for _, n := range []uint64{0, 1, 255, 256, 65535} {
		a, b, c := box16(uint16(n)), box32(uint32(n)), box64(n)
		runtime.GC()
		if a.(uint16) != uint16(n) || b.(uint32) != uint32(n) || c.(uint64) != n {
			panic("boxed scalar")
		}
	}
	for _, s := range []string{"", "payload"} {
		if boxString(s).(string) != s {
			panic("boxed string")
		}
	}
	if boxSlice(nil).([]byte) != nil {
		panic("boxed nil slice")
	}
	if len(boxSlice([]byte{}).([]byte)) != 0 {
		panic("boxed empty slice")
	}
	if makeBytes(0) == nil {
		panic("nil empty slice")
	}
	recovered := false
	func() {
		defer func() { recovered = recover() != nil }()
		_ = makeBytes(-1)
	}()
	if !recovered {
		panic("allocation panic removed")
	}
}
