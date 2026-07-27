// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type llvmSwitchInterface interface {
	Value(int) int
}

type llvmSwitchValue int

//go:noinline
func (v llvmSwitchValue) Value(delta int) int {
	return int(v) + delta
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
	default:
		return -1
	}
}

func main() {
	println(classifyLLVMInterface(nil))
	println(classifyLLVMInterface(llvmSwitchValue(10)))
	println(classifyLLVMInterface("abc"))
	println(classifyLLVMInterface(struct{}{}))
}
