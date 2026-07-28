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
// LLVM-DAG: @"type:codegen.llvmDescriptorBool" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorInt8" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorUintptr" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorFloat32" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorComplex128" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorString" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorUnsafePointer" = constant <{ %go.runtime.Type, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorArray" = constant <{ %go.runtime.ArrayType, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorChan" = constant <{ %go.runtime.ChanType, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorFunc" = constant <{ %go.runtime.FuncType, %go.runtime.UncommonType, [4 x ptr] }>
// LLVM-DAG: @"type:codegen.llvmDescriptorInterface" = constant <{ %go.runtime.InterfaceType, %go.runtime.UncommonType, [0 x %go.runtime.Imethod] }>
// LLVM-DAG: @"type:codegen.llvmDescriptorMap" = constant <{ %go.runtime.MapType, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorPtr" = constant <{ %go.runtime.PtrType, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorSlice" = constant <{ %go.runtime.SliceType, %go.runtime.UncommonType }>
// LLVM-DAG: @"type:codegen.llvmDescriptorStruct" = constant <{ %go.runtime.StructType, %go.runtime.UncommonType, [2 x %go.runtime.StructField] }>

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
