// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-LABEL: define goabiinternal i64 @codegen.goABIStatepointAttributes(
// LLVM-SAME: i64 %x) #[[ATTRS:[0-9]+]] gc "goallc" {
// LLVM: attributes #[[ATTRS]] = { {{.*}}"frame-pointer"="non-leaf"{{.*}}"go-stack-growth-statepoint"{{.*}} }
func goABIStatepointAttributes(x int) int {
	return x + 1
}
