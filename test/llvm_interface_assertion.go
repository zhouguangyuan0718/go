// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmAssertionInterface interface {
	Value(int) int
}

type llvmAssertionValue int

//go:noinline
func (v llvmAssertionValue) Value(delta int) int {
	return int(v) + delta
}

func assertLLVMConcrete(v any) (llvmAssertionValue, bool) {
	x, ok := v.(llvmAssertionValue)
	return x, ok
}

func assertLLVMInterface(v any) (llvmAssertionInterface, bool) {
	x, ok := v.(llvmAssertionInterface)
	return x, ok
}

func mustLLVMConcrete(v any) llvmAssertionValue {
	return v.(llvmAssertionValue)
}

func mustLLVMInterface(v any) llvmAssertionInterface {
	return v.(llvmAssertionInterface)
}

func main() {
	var v any = llvmAssertionValue(10)
	x, ok := assertLLVMConcrete(v)
	println(int(x), ok)

	i, ok := assertLLVMInterface(v)
	println(i.Value(3), ok)

	_, ok = assertLLVMConcrete("wrong")
	println(ok)

	i, ok = assertLLVMInterface(nil)
	println(i == nil, ok)

	println(int(mustLLVMConcrete(v)))
	println(mustLLVMInterface(v).Value(4))
}
