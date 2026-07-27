// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmMethodType struct {
	value int
}

//go:noinline
func (v llvmMethodType) Value(delta int) int {
	return v.value + delta
}

func (v llvmMethodType) hidden() int {
	return v.value
}

func (v *llvmMethodType) Pointer(delta int) int {
	return v.value + delta
}

func main() {
	println(llvmMethodType{value: 10}.Value(2))
}
