// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "simd/archsimd"

const archConstantShiftCases = 13

//go:noinline
func constantShiftAllLeftZeroUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllLeft(0)
}

//go:noinline
func constantShiftAllLeftLastUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllLeft(31)
}

//go:noinline
func constantShiftAllLeftWidthUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllLeft(32)
}

//go:noinline
func constantShiftAllLeftHighUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllLeft(1 << 32)
}

//go:noinline
func constantShiftAllLeftMaxUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllLeft(^uint64(0))
}

//go:noinline
func constantShiftAllRightZeroUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllRight(0)
}

//go:noinline
func constantShiftAllRightLastUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllRight(31)
}

//go:noinline
func constantShiftAllRightWidthUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllRight(32)
}

//go:noinline
func constantShiftAllRightHighUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllRight(1 << 32)
}

//go:noinline
func constantShiftAllRightMaxUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	return x.ShiftAllRight(^uint64(0))
}

//go:noinline
func constantShiftAllRightZeroInt64x2(x archsimd.Int64x2) archsimd.Int64x2 {
	return x.ShiftAllRight(0)
}

//go:noinline
func constantShiftAllRightMaxInt64x2(x archsimd.Int64x2) archsimd.Int64x2 {
	return x.ShiftAllRight(^uint64(0))
}

// A zero shift can return the preceding Add instruction unchanged.
//
//go:noinline
func constantShiftAllRightZeroAfterAddInt64x2(x, y archsimd.Int64x2) archsimd.Int64x2 {
	z := x.Add(y)
	return z.ShiftAllRight(0)
}

func checkConstantArch() {
	checkConstantShift[uint32]("constantShiftAllLeftZeroUint32x4", 32, 4, 0, false, false, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllLeftZeroUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftLastUint32x4", 32, 4, 31, false, false, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllLeftLastUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftWidthUint32x4", 32, 4, 32, false, false, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllLeftWidthUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftHighUint32x4", 32, 4, 1<<32, false, false, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllLeftHighUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftMaxUint32x4", 32, 4, ^uint64(0), false, false, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllLeftMaxUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightZeroUint32x4", 32, 4, 0, false, true, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllRightZeroUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightLastUint32x4", 32, 4, 31, false, true, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllRightLastUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightWidthUint32x4", 32, 4, 32, false, true, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllRightWidthUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightHighUint32x4", 32, 4, 1<<32, false, true, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllRightHighUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightMaxUint32x4", 32, 4, ^uint64(0), false, true, false, true, true, func(x, y, got []uint32) {
		z := constantShiftAllRightMaxUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[int64]("constantShiftAllRightZeroInt64x2", 64, 2, 0, true, true, false, true, true, func(x, y, got []int64) {
		z := constantShiftAllRightZeroInt64x2(archsimd.LoadInt64x2(x))
		z.Store(got)
	})
	checkConstantShift[int64]("constantShiftAllRightMaxInt64x2", 64, 2, ^uint64(0), true, true, false, true, true, func(x, y, got []int64) {
		z := constantShiftAllRightMaxInt64x2(archsimd.LoadInt64x2(x))
		z.Store(got)
	})
	checkConstantShift[int64]("constantShiftAllRightZeroAfterAddInt64x2", 64, 2, 0, true, true, true, true, true, func(x, y, got []int64) {
		z := constantShiftAllRightZeroAfterAddInt64x2(archsimd.LoadInt64x2(x), archsimd.LoadInt64x2(y))
		z.Store(got)
	})
}

const archShiftCases = 16

//go:noinline
func ordinaryShiftAllLeftInt8x16(x archsimd.Int8x16, count uint64) archsimd.Int8x16 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt16x8(x archsimd.Int16x8, count uint64) archsimd.Int16x8 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt32x4(x archsimd.Int32x4, count uint64) archsimd.Int32x4 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt64x2(x archsimd.Int64x2, count uint64) archsimd.Int64x2 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint8x16(x archsimd.Uint8x16, count uint64) archsimd.Uint8x16 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint16x8(x archsimd.Uint16x8, count uint64) archsimd.Uint16x8 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint32x4(x archsimd.Uint32x4, count uint64) archsimd.Uint32x4 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint64x2(x archsimd.Uint64x2, count uint64) archsimd.Uint64x2 {
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllRightInt8x16(x archsimd.Int8x16, count uint64) archsimd.Int8x16 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt16x8(x archsimd.Int16x8, count uint64) archsimd.Int16x8 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt32x4(x archsimd.Int32x4, count uint64) archsimd.Int32x4 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt64x2(x archsimd.Int64x2, count uint64) archsimd.Int64x2 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint8x16(x archsimd.Uint8x16, count uint64) archsimd.Uint8x16 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint16x8(x archsimd.Uint16x8, count uint64) archsimd.Uint16x8 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint32x4(x archsimd.Uint32x4, count uint64) archsimd.Uint32x4 {
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint64x2(x archsimd.Uint64x2, count uint64) archsimd.Uint64x2 {
	return x.ShiftAllRight(count)
}

func checkArch() {
	checkScalarShift[int8]("ShiftAllLeftInt8x16", 8, 16, true, false, true, true, func(x []int8, count uint64, got []int8) {
		z := ordinaryShiftAllLeftInt8x16(archsimd.LoadInt8x16(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllLeftInt16x8", 16, 8, true, false, true, true, func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllLeftInt16x8(archsimd.LoadInt16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllLeftInt32x4", 32, 4, true, false, true, true, func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllLeftInt32x4(archsimd.LoadInt32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllLeftInt64x2", 64, 2, true, false, true, true, func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllLeftInt64x2(archsimd.LoadInt64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[uint8]("ShiftAllLeftUint8x16", 8, 16, false, false, true, true, func(x []uint8, count uint64, got []uint8) {
		z := ordinaryShiftAllLeftUint8x16(archsimd.LoadUint8x16(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllLeftUint16x8", 16, 8, false, false, true, true, func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllLeftUint16x8(archsimd.LoadUint16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllLeftUint32x4", 32, 4, false, false, true, true, func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllLeftUint32x4(archsimd.LoadUint32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllLeftUint64x2", 64, 2, false, false, true, true, func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllLeftUint64x2(archsimd.LoadUint64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[int8]("ShiftAllRightInt8x16", 8, 16, true, true, true, true, func(x []int8, count uint64, got []int8) {
		z := ordinaryShiftAllRightInt8x16(archsimd.LoadInt8x16(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllRightInt16x8", 16, 8, true, true, true, true, func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllRightInt16x8(archsimd.LoadInt16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllRightInt32x4", 32, 4, true, true, true, true, func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllRightInt32x4(archsimd.LoadInt32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllRightInt64x2", 64, 2, true, true, true, true, func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllRightInt64x2(archsimd.LoadInt64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[uint8]("ShiftAllRightUint8x16", 8, 16, false, true, true, true, func(x []uint8, count uint64, got []uint8) {
		z := ordinaryShiftAllRightUint8x16(archsimd.LoadUint8x16(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllRightUint16x8", 16, 8, false, true, true, true, func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllRightUint16x8(archsimd.LoadUint16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllRightUint32x4", 32, 4, false, true, true, true, func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllRightUint32x4(archsimd.LoadUint32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllRightUint64x2", 64, 2, false, true, true, true, func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllRightUint64x2(archsimd.LoadUint64x2(x), count)
		z.Store(got)
	})
}
