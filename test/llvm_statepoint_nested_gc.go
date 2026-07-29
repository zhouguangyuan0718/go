// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

//go:noinline
func keepAcrossNestedGC(pointer *int, path int) int {
	adjustment := 0
	if path >= 2 {
		if path == 3 {
			runtime.GC()
			adjustment = 1
		} else {
			adjustment = 2
		}
	} else {
		adjustment = 3
	}
	return *pointer + adjustment
}

func main() {
	pointer := new(int)
	*pointer = 73
	if keepAcrossNestedGC(pointer, 0) != 76 {
		panic("pointer lost on outer skip path")
	}
	if keepAcrossNestedGC(pointer, 2) != 75 {
		panic("pointer lost on inner skip path")
	}
	if keepAcrossNestedGC(pointer, 3) != 74 {
		panic("pointer lost across nested GC")
	}
}
