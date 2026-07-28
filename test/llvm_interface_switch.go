// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmSwitchInterface interface {
	Value(int) int
}

type llvmDoubleSwitchInterface interface {
	Double() int
}

type llvmSwitchValue int
type llvmDoubleSwitchValue int

//go:noinline
func (v llvmSwitchValue) Value(delta int) int {
	return int(v) + delta
}

func (v llvmDoubleSwitchValue) Double() int {
	return int(v) * 2
}

func classifyLLVMInterface(v any) int {
	switch x := v.(type) {
	case nil:
		return 0
	case llvmSwitchValue:
		return int(x)
	case string:
		return len(x)
	case llvmSwitchInterface:
		return x.Value(1)
	case llvmDoubleSwitchInterface:
		return x.Double()
	default:
		return -1
	}
}

func main() {
	println(classifyLLVMInterface(nil))
	println(classifyLLVMInterface(llvmSwitchValue(10)))
	println(classifyLLVMInterface("abc"))
	println(classifyLLVMInterface(llvmDoubleSwitchValue(7)))
	println(classifyLLVMInterface(struct{}{}))
}
