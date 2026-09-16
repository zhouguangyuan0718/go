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

func main() {
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
