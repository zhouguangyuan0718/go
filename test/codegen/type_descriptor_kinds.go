// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "unsafe"

// LLVM-DAG: %go.runtime.Type = type <{
// LLVM-DAG: %go.runtime.ArrayType = type <{
// LLVM-DAG: %go.runtime.ChanType = type <{
// LLVM-DAG: %go.runtime.FuncType = type <{
// LLVM-DAG: %go.runtime.InterfaceType = type <{
// LLVM-DAG: %go.runtime.MapType = type <{
// LLVM-DAG: %go.runtime.PtrType = type <{
// LLVM-DAG: %go.runtime.SliceType = type <{
// LLVM-DAG: %go.runtime.StructType = type <{
// LLVM-DAG: %go.runtime.Imethod = type <{
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorBool = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorInt8 = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorUintptr = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorFloat32 = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorComplex128 = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorString = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorUnsafePointer = type <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorArray = type <{ %go.runtime.ArrayType, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorChan = type <{ %go.runtime.ChanType, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorFunc = type <{ %go.runtime.FuncType, %go.runtime.UncommonType, [4 x ptr] }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorInterface = type <{ %go.runtime.InterfaceType, %go.runtime.UncommonType, [0 x %go.runtime.Imethod] }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorMap = type <{ %go.runtime.MapType, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorPtr = type <{ %go.runtime.PtrType, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorSlice = type <{ %go.runtime.SliceType, %go.runtime.UncommonType }>
// LLVM-DAG: %go.descriptor.codegen.llvmDescriptorStruct = type <{ %go.runtime.StructType, %go.runtime.UncommonType, [2 x %go.runtime.StructField] }>

type llvmDescriptorBool bool
type llvmDescriptorInt8 int8
type llvmDescriptorInt16 int16
type llvmDescriptorInt32 int32
type llvmDescriptorInt64 int64
type llvmDescriptorInt int
type llvmDescriptorUint8 uint8
type llvmDescriptorUint16 uint16
type llvmDescriptorUint32 uint32
type llvmDescriptorUint64 uint64
type llvmDescriptorUint uint
type llvmDescriptorUintptr uintptr
type llvmDescriptorFloat32 float32
type llvmDescriptorFloat64 float64
type llvmDescriptorComplex64 complex64
type llvmDescriptorComplex128 complex128
type llvmDescriptorString string
type llvmDescriptorUnsafePointer unsafe.Pointer

type llvmDescriptorArray [3]uint16
type llvmDescriptorChan chan int
type llvmDescriptorFunc func(int, string) (bool, error)
type llvmDescriptorInterface interface{}
type llvmDescriptorMap map[int]string
type llvmDescriptorPtr *int
type llvmDescriptorSlice []byte
type llvmDescriptorStruct struct {
	x int
	y *byte
}

func useLLVMDescriptorKinds(v llvmDescriptorStruct) int {
	return v.x
}
