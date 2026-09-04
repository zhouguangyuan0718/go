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
func poisonStack() [20]int {
	return [20]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
}

//go:noinline
func tracebackWithPointerArg(p *int) int {
	if p == nil {
		panic("nil traceback argument")
	}
	return runtime.Stack(trace[:], false)
}

//go:nosplit
//go:noinline
func tracebackWithDeadArgs(a, b, c, d, e int32) int {
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

// tracebackLineHasQuestions reports whether the frame containing name has
// exactly want unavailable argument markers before the end of its line.
// LLVM does not yet prove individual register-home stores, so its traceback
// metadata must be conservative rather than treating stale stack data as live.
//
//go:noinline
func tracebackLineHasQuestions(data, name []byte, want int) bool {
	for i := 0; i+len(name) <= len(data); i++ {
		if !contains(data[i:i+len(name)], name) {
			continue
		}
		questions := 0
		for _, c := range data[i:] {
			if c == '\n' {
				return questions == want
			}
			if c == '?' {
				questions++
			}
		}
		return questions == want
	}
	return false
}

func main() {
	x := 1
	n := tracebackWithPointerArg(&x)
	if !contains(trace[:n], []byte("main.tracebackWithPointerArg(0x")) {
		panic("traceback lost pointer argument:\n" + string(trace[:n]))
	}

	poisonStack()
	n = tracebackWithDeadArgs(1, 2, 3, 4, 5)
	if !tracebackLineHasQuestions(trace[:n], []byte("main.tracebackWithDeadArgs("), 5) {
		panic("traceback treated dead register arguments as live:\n" + string(trace[:n]))
	}
}
