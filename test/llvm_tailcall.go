// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type staticBase struct {
	value int
}

//go:noinline
func (b *staticBase) add(delta int) int {
	return b.value + delta
}

type staticWrapper struct {
	*staticBase
}

var staticMethod = (*staticWrapper).add

func main() {
	static := &staticWrapper{staticBase: &staticBase{value: 40}}
	if got := staticMethod(static, 2); got != 42 {
		panic("static tail call")
	}
}
