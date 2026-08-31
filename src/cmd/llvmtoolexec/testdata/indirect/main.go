// Copyright 2026 The GoALLC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:noinline
//go:nosplit
func invoke(fn func(*int), value *int) {
	fn(value)
}

type caller interface {
	call(*int)
}

type targetType struct{}

//go:noinline
func (targetType) call(*int) {}

//go:noinline
//go:nosplit
func invokeInterface(target caller, value *int) {
	target.call(value)
}

//go:noinline
func target(*int) {}

func main() {
	var value int
	invoke(target, &value)
	invokeInterface(targetType{}, &value)
}
