// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmEqualityInterface interface {
	Value() int
}

type llvmEqualityValue int

func (v llvmEqualityValue) Value() int {
	return int(v)
}

func equalLLVMEmpty(a, b any) bool {
	return a == b
}

func equalLLVMNonEmpty(a, b llvmEqualityInterface) bool {
	return a == b
}

func main() {
	var a any = llvmEqualityValue(10)
	var b any = llvmEqualityValue(10)
	var c any = llvmEqualityValue(11)
	println(equalLLVMEmpty(a, b), equalLLVMEmpty(a, c))

	var x llvmEqualityInterface = llvmEqualityValue(10)
	var y llvmEqualityInterface = llvmEqualityValue(10)
	var z llvmEqualityInterface = llvmEqualityValue(11)
	println(equalLLVMNonEmpty(x, y), equalLLVMNonEmpty(x, z))
}
