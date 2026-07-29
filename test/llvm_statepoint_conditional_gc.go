// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

//go:noinline
func keepAcrossConditionalGC(pointer *int, collect bool) int {
	if collect {
		runtime.GC()
	}
	return *pointer
}

func main() {
	pointer := new(int)
	*pointer = 73
	if keepAcrossConditionalGC(pointer, false) != 73 {
		panic("pointer lost on path which skipped GC")
	}
	if keepAcrossConditionalGC(pointer, true) != 73 {
		panic("pointer lost on path which crossed GC")
	}
}
