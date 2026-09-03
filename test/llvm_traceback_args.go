// run -llvm-package-only

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
)

var trace [4096]byte

//go:noinline
func tracebackWithPointerArg(p *int) int {
	if p == nil {
		panic("nil traceback argument")
	}
	return runtime.Stack(trace[:], false)
}

//go:noinline
func contains(haystack, needle []byte) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func main() {
	x := 1
	n := tracebackWithPointerArg(&x)
	if !contains(trace[:n], []byte("main.tracebackWithPointerArg(0x")) {
		panic("traceback lost pointer argument:\n" + string(trace[:n]))
	}
}
