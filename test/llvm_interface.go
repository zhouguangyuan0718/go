// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmInterface interface {
	Double() int
	Value(int) int
}

type llvmInterfaceValue struct {
	value int
}

//go:noinline
func (v llvmInterfaceValue) Value(delta int) int {
	return v.value + delta
}

//go:noinline
func (v llvmInterfaceValue) Double() int {
	return v.value * 2
}

func callLLVMInterface(i llvmInterface) int {
	return i.Value(2) + i.Double()
}

func main() {
	println(callLLVMInterface(llvmInterfaceValue{value: 10}))
}
