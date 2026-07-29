// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

//go:noinline
func keepAcrossBranchGC(pointer *int, left bool) int {
	adjustment := 0
	if left {
		runtime.GC()
		adjustment = 1
	} else {
		runtime.GC()
		adjustment = 2
	}
	return *pointer + adjustment
}

func main() {
	pointer := new(int)
	*pointer = 73
	if keepAcrossBranchGC(pointer, true) != 74 {
		panic("pointer lost on left GC path")
	}
	if keepAcrossBranchGC(pointer, false) != 75 {
		panic("pointer lost on right GC path")
	}
}
