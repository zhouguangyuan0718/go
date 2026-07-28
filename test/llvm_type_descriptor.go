// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// This named type requires reflectdata to emit both its descriptor and the
// descriptor for its pointer type. Keep the program otherwise simple so this
// is an execution test of LLVM data lowering and GoObj linking.
type llvmTypeDescriptor struct {
	value int
}

func main() {
	println(llvmTypeDescriptor{value: 7}.value)
}
