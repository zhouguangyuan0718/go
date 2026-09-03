// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64

package main

import "math"

// LLVM-LINK: nosplit: main.shannonEntropyBits{{<[^>]+>}} +8 -> runtime.morestack
// LLVM-LINK-NOT: nosplit stack over
func shannonEntropyBits(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	var hist [256]int
	for _, c := range b {
		hist[c]++
	}
	shannon := float64(0)
	invTotal := 1.0 / float64(len(b))
	for _, v := range hist[:] {
		if v > 0 {
			n := float64(v)
			shannon += math.Ceil(-math.Log2(n*invTotal) * n)
		}
	}
	return int(math.Ceil(shannon))
}

func main() {
	input := make([]byte, 256)
	for i := range input {
		input[i] = byte(i)
	}
	if got := shannonEntropyBits(input); got != 2048 {
		panic(got)
	}
}
