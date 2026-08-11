// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

type g struct{}

//go:noescape
func getg() *g

// llvmGetGABI0 calls the compiler intrinsic runtime.getg. The cgo unsafe-args
// ABI pins the function to ABI0.
//
//go:cgo_unsafe_args
func llvmGetGABI0() *g {
	return getg()
}
