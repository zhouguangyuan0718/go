// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// A named map descriptor carries an R_KEEP edge to its underlying map type.
// The static interface conversion forces reflectdata to materialize both
// descriptors without requiring map-operation lowering.
type llvmMapDescriptor map[int]string

var keep any = (llvmMapDescriptor)(nil)

func main() {
	println(7)
}
