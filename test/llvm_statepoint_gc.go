// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

//go:noinline
func keepAcrossGC(pointer *int) int {
	runtime.GC()
	return *pointer
}

func main() {
	pointer := new(int)
	*pointer = 73
	if value := keepAcrossGC(pointer); value != 73 {
		panic("pointer lost across GC")
	}
}
