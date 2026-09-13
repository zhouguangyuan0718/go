// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import "github.com/goallc/go-llvm"

// simdShift implements the ordinary unsigned-count shift families. Unlike
// LLVM shifts, the Go operations define results for every count, including
// counts equal to or greater than the lane width. Clamp the actual shift
// operand before shifting, so even an unused result cannot introduce poison.
func (lfc *LLVMFuncContext) simdShift(v *Value, info goALLCSIMDOpInfo, laneType llvm.Type, lanes int) llvm.Value {
	vectorType := llvm.VectorType(laneType, lanes)
	x := lfc.simdValueAs(v, v.Args[0], vectorType, ".x")
	width := uint64(info.laneBits)
	all := info.lowering == goALLCSIMDLowerShiftAllLeft || info.lowering == goALLCSIMDLowerShiftAllRight
	left := info.lowering == goALLCSIMDLowerShiftAllLeft || info.lowering == goALLCSIMDLowerShiftLeft
	signed := !left && info.lane == goALLCSIMDLaneInt
	var count, limit, last llvm.Value
	if all {
		count = lfc.GenLV(v.Args[1])
		if count.Type().TypeKind() != llvm.IntegerTypeKind || count.Type().IntTypeWidth() != 64 {
			v.Fatalf("%s requires an unsigned 64-bit scalar count", v.Op)
		}
		limit = llvm.ConstInt(count.Type(), width, false)
		last = llvm.ConstInt(count.Type(), width-1, false)
	} else {
		count = lfc.simdValueAs(v, v.Args[1], vectorType, ".count")
		limit = llvmSIMDIntegerSplat(laneType, lanes, width)
		last = llvmSIMDIntegerSplat(laneType, lanes, width-1)
	}
	// Test the original count, especially before narrowing scalar uint64
	// counts to i8/i16/i32. For example, 256 must not become an i8 zero.
	inRange := lfc.b.CreateICmp(llvm.IntULT, count, limit, v.String()+".inrange")
	safeCount := lfc.b.CreateSelect(inRange, count, last, v.String()+".safe")
	if all {
		if width != 64 {
			safeCount = lfc.b.CreateTrunc(safeCount, laneType, v.String()+".narrow")
		}
		first := lfc.b.CreateInsertElement(llvm.ConstNull(vectorType), safeCount, llvm.ConstInt(GlobalCtxt.Int32Type(), 0, false), "")
		safeCount = lfc.b.CreateShuffleVector(first, llvm.ConstNull(vectorType), llvmVectorShuffleMask(make([]uint64, lanes)...), v.String()+".counts")
	}
	var result llvm.Value
	switch {
	case left:
		result = lfc.b.CreateShl(x, safeCount, v.String()+".shift")
	case signed:
		result = lfc.b.CreateAShr(x, safeCount, v.String()+".shift")
	default:
		result = lfc.b.CreateLShr(x, safeCount, v.String()+".shift")
	}
	// Arithmetic right shift at width-1 already produces the required sign
	// fill. Left and logical right shifts instead produce zero out of range.
	if !signed {
		result = lfc.b.CreateSelect(inRange, result, llvm.ConstNull(vectorType), v.String()+".selected")
	}
	return lfc.simdLaneResult(v, result)
}
