// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmSourceInterface interface {
	Double() int
	Value(int) int
}

type llvmTargetInterface interface {
	Value(int) int
}

type llvmConversionValue int

//go:noinline
func (v llvmConversionValue) Double() int {
	return int(v) * 2
}

//go:noinline
func (v llvmConversionValue) Value(delta int) int {
	return int(v) + delta
}

func convertLLVMInterface(v llvmSourceInterface) llvmTargetInterface {
	return v
}

func main() {
	var source llvmSourceInterface = llvmConversionValue(10)
	println(convertLLVMInterface(source).Value(2))
	println(convertLLVMInterface(source).Value(3))

	var nilSource llvmSourceInterface
	println(convertLLVMInterface(nilSource) == nil)
}
