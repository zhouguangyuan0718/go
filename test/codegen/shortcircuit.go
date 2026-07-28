// asmcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal i64 @codegen.efaceExtract(
// LLVM: extractvalue { ptr, ptr } %e, 0
// LLVM: icmp eq ptr %{{.*}}, @"type:int"
// LLVM: extractvalue { ptr, ptr } %e, 1

func efaceExtract(e interface{}) int {
	// This should be compiled with only
	// a single conditional jump.
	// amd64:-"JMP"
	if x, ok := e.(int); ok {
		return x
	}
	return 0
}
