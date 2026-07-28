// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "unsafe"

// These declarations exercise every Go runtime descriptor kind while keeping
// execution independent of interface, map, channel, and function operations.
type descriptorBool bool
type descriptorInt8 int8
type descriptorInt16 int16
type descriptorInt32 int32
type descriptorInt64 int64
type descriptorInt int
type descriptorUint8 uint8
type descriptorUint16 uint16
type descriptorUint32 uint32
type descriptorUint64 uint64
type descriptorUint uint
type descriptorUintptr uintptr
type descriptorFloat32 float32
type descriptorFloat64 float64
type descriptorComplex64 complex64
type descriptorComplex128 complex128
type descriptorString string
type descriptorUnsafePointer unsafe.Pointer

type descriptorArray [3]uint16
type descriptorChan chan int
type descriptorFunc func(int, string) (bool, error)
type descriptorInterface interface{}
type descriptorMap map[int]string
type descriptorPtr *int
type descriptorSlice []byte
type descriptorStruct struct {
	x int
	y *byte
}

func main() {
	println(7)
}
