// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type addressedCallResult [20]int

//go:noinline
func makeAddressedCallResult(seed int) (result addressedCallResult) {
	for i := range result {
		result[i] = seed + i
	}
	return
}

func main() {
	result := makeAddressedCallResult(23)
	for i, value := range result {
		if want := 23 + i; value != want {
			panic("addressed call result mismatch")
		}
	}
}
