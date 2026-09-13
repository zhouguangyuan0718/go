// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "simd/archsimd"

const archConstantShiftCases = 23

//go:noinline
func constantShiftAllLeftZeroUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllLeft(0)
}

//go:noinline
func constantShiftAllLeftLastUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllLeft(31)
}

//go:noinline
func constantShiftAllLeftWidthUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllLeft(32)
}

//go:noinline
func constantShiftAllLeftHighUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllLeft(1 << 32)
}

//go:noinline
func constantShiftAllLeftMaxUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllLeft(^uint64(0))
}

//go:noinline
func constantShiftAllRightZeroUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllRight(0)
}

//go:noinline
func constantShiftAllRightLastUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllRight(31)
}

//go:noinline
func constantShiftAllRightWidthUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllRight(32)
}

//go:noinline
func constantShiftAllRightHighUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllRight(1 << 32)
}

//go:noinline
func constantShiftAllRightMaxUint32x4(x archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllRight(^uint64(0))
}

//go:noinline
func constantShiftAllLeftZeroUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllLeft(0)
}

//go:noinline
func constantShiftAllLeftLastUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllLeft(31)
}

//go:noinline
func constantShiftAllLeftWidthUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllLeft(32)
}

//go:noinline
func constantShiftAllLeftHighUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllLeft(1 << 32)
}

//go:noinline
func constantShiftAllLeftMaxUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllLeft(^uint64(0))
}

//go:noinline
func constantShiftAllRightZeroUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllRight(0)
}

//go:noinline
func constantShiftAllRightLastUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllRight(31)
}

//go:noinline
func constantShiftAllRightWidthUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllRight(32)
}

//go:noinline
func constantShiftAllRightHighUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllRight(1 << 32)
}

//go:noinline
func constantShiftAllRightMaxUint32x8(x archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllRight(^uint64(0))
}

//go:noinline
func constantShiftAllRightZeroInt64x2(x archsimd.Int64x2) archsimd.Int64x2 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x2{}
	}
	return x.ShiftAllRight(0)
}

//go:noinline
func constantShiftAllRightMaxInt64x2(x archsimd.Int64x2) archsimd.Int64x2 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x2{}
	}
	return x.ShiftAllRight(^uint64(0))
}

// A zero shift returns the preceding Add instruction. Its own requirement
// must remain inside the stronger guard, without moving onto that Add.
//
//go:noinline
func constantShiftAllRightZeroAfterAddInt64x2(x, y archsimd.Int64x2) archsimd.Int64x2 {
	if !archsimd.X86.AVX() {
		return archsimd.Int64x2{}
	}
	z := x.Add(y)
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x2{}
	}
	return z.ShiftAllRight(0)
}

func checkConstantArch() {
	checkConstantShift[uint32]("constantShiftAllLeftZeroUint32x4", 32, 4, 0, false, false, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllLeftZeroUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftLastUint32x4", 32, 4, 31, false, false, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllLeftLastUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftWidthUint32x4", 32, 4, 32, false, false, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllLeftWidthUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftHighUint32x4", 32, 4, 1<<32, false, false, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllLeftHighUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftMaxUint32x4", 32, 4, ^uint64(0), false, false, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllLeftMaxUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightZeroUint32x4", 32, 4, 0, false, true, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllRightZeroUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightLastUint32x4", 32, 4, 31, false, true, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllRightLastUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightWidthUint32x4", 32, 4, 32, false, true, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllRightWidthUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightHighUint32x4", 32, 4, 1<<32, false, true, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllRightHighUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightMaxUint32x4", 32, 4, ^uint64(0), false, true, false, true, archsimd.X86.AVX(), func(x, y, got []uint32) {
		z := constantShiftAllRightMaxUint32x4(archsimd.LoadUint32x4(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftZeroUint32x8", 32, 8, 0, false, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllLeftZeroUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftLastUint32x8", 32, 8, 31, false, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllLeftLastUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftWidthUint32x8", 32, 8, 32, false, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllLeftWidthUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftHighUint32x8", 32, 8, 1<<32, false, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllLeftHighUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllLeftMaxUint32x8", 32, 8, ^uint64(0), false, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllLeftMaxUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightZeroUint32x8", 32, 8, 0, false, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllRightZeroUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightLastUint32x8", 32, 8, 31, false, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllRightLastUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightWidthUint32x8", 32, 8, 32, false, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllRightWidthUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightHighUint32x8", 32, 8, 1<<32, false, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllRightHighUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[uint32]("constantShiftAllRightMaxUint32x8", 32, 8, ^uint64(0), false, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x, y, got []uint32) {
		z := constantShiftAllRightMaxUint32x8(archsimd.LoadUint32x8(x))
		z.Store(got)
	})
	checkConstantShift[int64]("constantShiftAllRightZeroInt64x2", 64, 2, 0, true, true, false, true, archsimd.X86.AVX512(), func(x, y, got []int64) {
		z := constantShiftAllRightZeroInt64x2(archsimd.LoadInt64x2(x))
		z.Store(got)
	})
	checkConstantShift[int64]("constantShiftAllRightMaxInt64x2", 64, 2, ^uint64(0), true, true, false, true, archsimd.X86.AVX512(), func(x, y, got []int64) {
		z := constantShiftAllRightMaxInt64x2(archsimd.LoadInt64x2(x))
		z.Store(got)
	})
	checkConstantShift[int64]("constantShiftAllRightZeroAfterAddInt64x2", 64, 2, 0, true, true, true, true, archsimd.X86.AVX() && archsimd.X86.AVX512(), func(x, y, got []int64) {
		z := constantShiftAllRightZeroAfterAddInt64x2(archsimd.LoadInt64x2(x), archsimd.LoadInt64x2(y))
		z.Store(got)
	})
}

const archShiftCases = 72

//go:noinline
func ordinaryShiftAllLeftInt16x8(x archsimd.Int16x8, count uint64) archsimd.Int16x8 {
	if !archsimd.X86.AVX() {
		return archsimd.Int16x8{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt16x16(x archsimd.Int16x16, count uint64) archsimd.Int16x16 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int16x16{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt16x32(x archsimd.Int16x32, count uint64) archsimd.Int16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x32{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt32x4(x archsimd.Int32x4, count uint64) archsimd.Int32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Int32x4{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt32x8(x archsimd.Int32x8, count uint64) archsimd.Int32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int32x8{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt32x16(x archsimd.Int32x16, count uint64) archsimd.Int32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int32x16{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt64x2(x archsimd.Int64x2, count uint64) archsimd.Int64x2 {
	if !archsimd.X86.AVX() {
		return archsimd.Int64x2{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt64x4(x archsimd.Int64x4, count uint64) archsimd.Int64x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int64x4{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftInt64x8(x archsimd.Int64x8, count uint64) archsimd.Int64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x8{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint16x8(x archsimd.Uint16x8, count uint64) archsimd.Uint16x8 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint16x8{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint16x16(x archsimd.Uint16x16, count uint64) archsimd.Uint16x16 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint16x16{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint16x32(x archsimd.Uint16x32, count uint64) archsimd.Uint16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x32{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint32x4(x archsimd.Uint32x4, count uint64) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint32x8(x archsimd.Uint32x8, count uint64) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint32x16(x archsimd.Uint32x16, count uint64) archsimd.Uint32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint32x16{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint64x2(x archsimd.Uint64x2, count uint64) archsimd.Uint64x2 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint64x2{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint64x4(x archsimd.Uint64x4, count uint64) archsimd.Uint64x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint64x4{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllLeftUint64x8(x archsimd.Uint64x8, count uint64) archsimd.Uint64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint64x8{}
	}
	return x.ShiftAllLeft(count)
}

//go:noinline
func ordinaryShiftAllRightInt16x8(x archsimd.Int16x8, count uint64) archsimd.Int16x8 {
	if !archsimd.X86.AVX() {
		return archsimd.Int16x8{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt16x16(x archsimd.Int16x16, count uint64) archsimd.Int16x16 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int16x16{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt16x32(x archsimd.Int16x32, count uint64) archsimd.Int16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x32{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt32x4(x archsimd.Int32x4, count uint64) archsimd.Int32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Int32x4{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt32x8(x archsimd.Int32x8, count uint64) archsimd.Int32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int32x8{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt32x16(x archsimd.Int32x16, count uint64) archsimd.Int32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int32x16{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt64x2(x archsimd.Int64x2, count uint64) archsimd.Int64x2 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x2{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt64x4(x archsimd.Int64x4, count uint64) archsimd.Int64x4 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x4{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightInt64x8(x archsimd.Int64x8, count uint64) archsimd.Int64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x8{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint16x8(x archsimd.Uint16x8, count uint64) archsimd.Uint16x8 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint16x8{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint16x16(x archsimd.Uint16x16, count uint64) archsimd.Uint16x16 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint16x16{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint16x32(x archsimd.Uint16x32, count uint64) archsimd.Uint16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x32{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint32x4(x archsimd.Uint32x4, count uint64) archsimd.Uint32x4 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint32x8(x archsimd.Uint32x8, count uint64) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint32x16(x archsimd.Uint32x16, count uint64) archsimd.Uint32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint32x16{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint64x2(x archsimd.Uint64x2, count uint64) archsimd.Uint64x2 {
	if !archsimd.X86.AVX() {
		return archsimd.Uint64x2{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint64x4(x archsimd.Uint64x4, count uint64) archsimd.Uint64x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint64x4{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftAllRightUint64x8(x archsimd.Uint64x8, count uint64) archsimd.Uint64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint64x8{}
	}
	return x.ShiftAllRight(count)
}

//go:noinline
func ordinaryShiftLeftInt16x8(x archsimd.Int16x8, count archsimd.Uint16x8) archsimd.Int16x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x8{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt16x16(x archsimd.Int16x16, count archsimd.Uint16x16) archsimd.Int16x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x16{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt16x32(x archsimd.Int16x32, count archsimd.Uint16x32) archsimd.Int16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x32{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt32x4(x archsimd.Int32x4, count archsimd.Uint32x4) archsimd.Int32x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int32x4{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt32x8(x archsimd.Int32x8, count archsimd.Uint32x8) archsimd.Int32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int32x8{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt32x16(x archsimd.Int32x16, count archsimd.Uint32x16) archsimd.Int32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int32x16{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt64x2(x archsimd.Int64x2, count archsimd.Uint64x2) archsimd.Int64x2 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int64x2{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt64x4(x archsimd.Int64x4, count archsimd.Uint64x4) archsimd.Int64x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int64x4{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftInt64x8(x archsimd.Int64x8, count archsimd.Uint64x8) archsimd.Int64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x8{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint16x8(x archsimd.Uint16x8, count archsimd.Uint16x8) archsimd.Uint16x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x8{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint16x16(x archsimd.Uint16x16, count archsimd.Uint16x16) archsimd.Uint16x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x16{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint16x32(x archsimd.Uint16x32, count archsimd.Uint16x32) archsimd.Uint16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x32{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint32x4(x archsimd.Uint32x4, count archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint32x8(x archsimd.Uint32x8, count archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint32x16(x archsimd.Uint32x16, count archsimd.Uint32x16) archsimd.Uint32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint32x16{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint64x2(x archsimd.Uint64x2, count archsimd.Uint64x2) archsimd.Uint64x2 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint64x2{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint64x4(x archsimd.Uint64x4, count archsimd.Uint64x4) archsimd.Uint64x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint64x4{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftLeftUint64x8(x archsimd.Uint64x8, count archsimd.Uint64x8) archsimd.Uint64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint64x8{}
	}
	return x.ShiftLeft(count)
}

//go:noinline
func ordinaryShiftRightInt16x8(x archsimd.Int16x8, count archsimd.Uint16x8) archsimd.Int16x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x8{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt16x16(x archsimd.Int16x16, count archsimd.Uint16x16) archsimd.Int16x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x16{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt16x32(x archsimd.Int16x32, count archsimd.Uint16x32) archsimd.Int16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int16x32{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt32x4(x archsimd.Int32x4, count archsimd.Uint32x4) archsimd.Int32x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int32x4{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt32x8(x archsimd.Int32x8, count archsimd.Uint32x8) archsimd.Int32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Int32x8{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt32x16(x archsimd.Int32x16, count archsimd.Uint32x16) archsimd.Int32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int32x16{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt64x2(x archsimd.Int64x2, count archsimd.Uint64x2) archsimd.Int64x2 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x2{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt64x4(x archsimd.Int64x4, count archsimd.Uint64x4) archsimd.Int64x4 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x4{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightInt64x8(x archsimd.Int64x8, count archsimd.Uint64x8) archsimd.Int64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Int64x8{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint16x8(x archsimd.Uint16x8, count archsimd.Uint16x8) archsimd.Uint16x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x8{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint16x16(x archsimd.Uint16x16, count archsimd.Uint16x16) archsimd.Uint16x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x16{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint16x32(x archsimd.Uint16x32, count archsimd.Uint16x32) archsimd.Uint16x32 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint16x32{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint32x4(x archsimd.Uint32x4, count archsimd.Uint32x4) archsimd.Uint32x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x4{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint32x8(x archsimd.Uint32x8, count archsimd.Uint32x8) archsimd.Uint32x8 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint32x8{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint32x16(x archsimd.Uint32x16, count archsimd.Uint32x16) archsimd.Uint32x16 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint32x16{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint64x2(x archsimd.Uint64x2, count archsimd.Uint64x2) archsimd.Uint64x2 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint64x2{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint64x4(x archsimd.Uint64x4, count archsimd.Uint64x4) archsimd.Uint64x4 {
	if !archsimd.X86.AVX2() {
		return archsimd.Uint64x4{}
	}
	return x.ShiftRight(count)
}

//go:noinline
func ordinaryShiftRightUint64x8(x archsimd.Uint64x8, count archsimd.Uint64x8) archsimd.Uint64x8 {
	if !archsimd.X86.AVX512() {
		return archsimd.Uint64x8{}
	}
	return x.ShiftRight(count)
}

// Load, store, and argument passing must be legal even when the operation's
// own guard is false. 256-bit values need AVX; 512-bit values need AVX512.
func checkArch() {
	checkScalarShift[int16]("ShiftAllLeftInt16x8", 16, 8, true, false, true, archsimd.X86.AVX(), func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllLeftInt16x8(archsimd.LoadInt16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllLeftInt16x16", 16, 16, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllLeftInt16x16(archsimd.LoadInt16x16(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllLeftInt16x32", 16, 32, true, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllLeftInt16x32(archsimd.LoadInt16x32(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllLeftInt32x4", 32, 4, true, false, true, archsimd.X86.AVX(), func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllLeftInt32x4(archsimd.LoadInt32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllLeftInt32x8", 32, 8, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllLeftInt32x8(archsimd.LoadInt32x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllLeftInt32x16", 32, 16, true, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllLeftInt32x16(archsimd.LoadInt32x16(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllLeftInt64x2", 64, 2, true, false, true, archsimd.X86.AVX(), func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllLeftInt64x2(archsimd.LoadInt64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllLeftInt64x4", 64, 4, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllLeftInt64x4(archsimd.LoadInt64x4(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllLeftInt64x8", 64, 8, true, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllLeftInt64x8(archsimd.LoadInt64x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllLeftUint16x8", 16, 8, false, false, true, archsimd.X86.AVX(), func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllLeftUint16x8(archsimd.LoadUint16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllLeftUint16x16", 16, 16, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllLeftUint16x16(archsimd.LoadUint16x16(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllLeftUint16x32", 16, 32, false, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllLeftUint16x32(archsimd.LoadUint16x32(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllLeftUint32x4", 32, 4, false, false, true, archsimd.X86.AVX(), func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllLeftUint32x4(archsimd.LoadUint32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllLeftUint32x8", 32, 8, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllLeftUint32x8(archsimd.LoadUint32x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllLeftUint32x16", 32, 16, false, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllLeftUint32x16(archsimd.LoadUint32x16(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllLeftUint64x2", 64, 2, false, false, true, archsimd.X86.AVX(), func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllLeftUint64x2(archsimd.LoadUint64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllLeftUint64x4", 64, 4, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllLeftUint64x4(archsimd.LoadUint64x4(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllLeftUint64x8", 64, 8, false, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllLeftUint64x8(archsimd.LoadUint64x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllRightInt16x8", 16, 8, true, true, true, archsimd.X86.AVX(), func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllRightInt16x8(archsimd.LoadInt16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllRightInt16x16", 16, 16, true, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllRightInt16x16(archsimd.LoadInt16x16(x), count)
		z.Store(got)
	})
	checkScalarShift[int16]("ShiftAllRightInt16x32", 16, 32, true, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int16, count uint64, got []int16) {
		z := ordinaryShiftAllRightInt16x32(archsimd.LoadInt16x32(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllRightInt32x4", 32, 4, true, true, true, archsimd.X86.AVX(), func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllRightInt32x4(archsimd.LoadInt32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllRightInt32x8", 32, 8, true, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllRightInt32x8(archsimd.LoadInt32x8(x), count)
		z.Store(got)
	})
	checkScalarShift[int32]("ShiftAllRightInt32x16", 32, 16, true, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int32, count uint64, got []int32) {
		z := ordinaryShiftAllRightInt32x16(archsimd.LoadInt32x16(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllRightInt64x2", 64, 2, true, true, true, archsimd.X86.AVX512(), func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllRightInt64x2(archsimd.LoadInt64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllRightInt64x4", 64, 4, true, true, archsimd.X86.AVX(), archsimd.X86.AVX512(), func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllRightInt64x4(archsimd.LoadInt64x4(x), count)
		z.Store(got)
	})
	checkScalarShift[int64]("ShiftAllRightInt64x8", 64, 8, true, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int64, count uint64, got []int64) {
		z := ordinaryShiftAllRightInt64x8(archsimd.LoadInt64x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllRightUint16x8", 16, 8, false, true, true, archsimd.X86.AVX(), func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllRightUint16x8(archsimd.LoadUint16x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllRightUint16x16", 16, 16, false, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllRightUint16x16(archsimd.LoadUint16x16(x), count)
		z.Store(got)
	})
	checkScalarShift[uint16]("ShiftAllRightUint16x32", 16, 32, false, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint16, count uint64, got []uint16) {
		z := ordinaryShiftAllRightUint16x32(archsimd.LoadUint16x32(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllRightUint32x4", 32, 4, false, true, true, archsimd.X86.AVX(), func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllRightUint32x4(archsimd.LoadUint32x4(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllRightUint32x8", 32, 8, false, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllRightUint32x8(archsimd.LoadUint32x8(x), count)
		z.Store(got)
	})
	checkScalarShift[uint32]("ShiftAllRightUint32x16", 32, 16, false, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint32, count uint64, got []uint32) {
		z := ordinaryShiftAllRightUint32x16(archsimd.LoadUint32x16(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllRightUint64x2", 64, 2, false, true, true, archsimd.X86.AVX(), func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllRightUint64x2(archsimd.LoadUint64x2(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllRightUint64x4", 64, 4, false, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllRightUint64x4(archsimd.LoadUint64x4(x), count)
		z.Store(got)
	})
	checkScalarShift[uint64]("ShiftAllRightUint64x8", 64, 8, false, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint64, count uint64, got []uint64) {
		z := ordinaryShiftAllRightUint64x8(archsimd.LoadUint64x8(x), count)
		z.Store(got)
	})
	checkVectorShift[int16, uint16]("ShiftLeftInt16x8", 16, 8, true, false, true, archsimd.X86.AVX512(), func(x []int16, count []uint16, got []int16) {
		z := ordinaryShiftLeftInt16x8(archsimd.LoadInt16x8(x), archsimd.LoadUint16x8(count))
		z.Store(got)
	})
	checkVectorShift[int16, uint16]("ShiftLeftInt16x16", 16, 16, true, false, archsimd.X86.AVX(), archsimd.X86.AVX512(), func(x []int16, count []uint16, got []int16) {
		z := ordinaryShiftLeftInt16x16(archsimd.LoadInt16x16(x), archsimd.LoadUint16x16(count))
		z.Store(got)
	})
	checkVectorShift[int16, uint16]("ShiftLeftInt16x32", 16, 32, true, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int16, count []uint16, got []int16) {
		z := ordinaryShiftLeftInt16x32(archsimd.LoadInt16x32(x), archsimd.LoadUint16x32(count))
		z.Store(got)
	})
	checkVectorShift[int32, uint32]("ShiftLeftInt32x4", 32, 4, true, false, true, archsimd.X86.AVX2(), func(x []int32, count []uint32, got []int32) {
		z := ordinaryShiftLeftInt32x4(archsimd.LoadInt32x4(x), archsimd.LoadUint32x4(count))
		z.Store(got)
	})
	checkVectorShift[int32, uint32]("ShiftLeftInt32x8", 32, 8, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int32, count []uint32, got []int32) {
		z := ordinaryShiftLeftInt32x8(archsimd.LoadInt32x8(x), archsimd.LoadUint32x8(count))
		z.Store(got)
	})
	checkVectorShift[int32, uint32]("ShiftLeftInt32x16", 32, 16, true, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int32, count []uint32, got []int32) {
		z := ordinaryShiftLeftInt32x16(archsimd.LoadInt32x16(x), archsimd.LoadUint32x16(count))
		z.Store(got)
	})
	checkVectorShift[int64, uint64]("ShiftLeftInt64x2", 64, 2, true, false, true, archsimd.X86.AVX2(), func(x []int64, count []uint64, got []int64) {
		z := ordinaryShiftLeftInt64x2(archsimd.LoadInt64x2(x), archsimd.LoadUint64x2(count))
		z.Store(got)
	})
	checkVectorShift[int64, uint64]("ShiftLeftInt64x4", 64, 4, true, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int64, count []uint64, got []int64) {
		z := ordinaryShiftLeftInt64x4(archsimd.LoadInt64x4(x), archsimd.LoadUint64x4(count))
		z.Store(got)
	})
	checkVectorShift[int64, uint64]("ShiftLeftInt64x8", 64, 8, true, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int64, count []uint64, got []int64) {
		z := ordinaryShiftLeftInt64x8(archsimd.LoadInt64x8(x), archsimd.LoadUint64x8(count))
		z.Store(got)
	})
	checkVectorShift[uint16, uint16]("ShiftLeftUint16x8", 16, 8, false, false, true, archsimd.X86.AVX512(), func(x []uint16, count []uint16, got []uint16) {
		z := ordinaryShiftLeftUint16x8(archsimd.LoadUint16x8(x), archsimd.LoadUint16x8(count))
		z.Store(got)
	})
	checkVectorShift[uint16, uint16]("ShiftLeftUint16x16", 16, 16, false, false, archsimd.X86.AVX(), archsimd.X86.AVX512(), func(x []uint16, count []uint16, got []uint16) {
		z := ordinaryShiftLeftUint16x16(archsimd.LoadUint16x16(x), archsimd.LoadUint16x16(count))
		z.Store(got)
	})
	checkVectorShift[uint16, uint16]("ShiftLeftUint16x32", 16, 32, false, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint16, count []uint16, got []uint16) {
		z := ordinaryShiftLeftUint16x32(archsimd.LoadUint16x32(x), archsimd.LoadUint16x32(count))
		z.Store(got)
	})
	checkVectorShift[uint32, uint32]("ShiftLeftUint32x4", 32, 4, false, false, true, archsimd.X86.AVX2(), func(x []uint32, count []uint32, got []uint32) {
		z := ordinaryShiftLeftUint32x4(archsimd.LoadUint32x4(x), archsimd.LoadUint32x4(count))
		z.Store(got)
	})
	checkVectorShift[uint32, uint32]("ShiftLeftUint32x8", 32, 8, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint32, count []uint32, got []uint32) {
		z := ordinaryShiftLeftUint32x8(archsimd.LoadUint32x8(x), archsimd.LoadUint32x8(count))
		z.Store(got)
	})
	checkVectorShift[uint32, uint32]("ShiftLeftUint32x16", 32, 16, false, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint32, count []uint32, got []uint32) {
		z := ordinaryShiftLeftUint32x16(archsimd.LoadUint32x16(x), archsimd.LoadUint32x16(count))
		z.Store(got)
	})
	checkVectorShift[uint64, uint64]("ShiftLeftUint64x2", 64, 2, false, false, true, archsimd.X86.AVX2(), func(x []uint64, count []uint64, got []uint64) {
		z := ordinaryShiftLeftUint64x2(archsimd.LoadUint64x2(x), archsimd.LoadUint64x2(count))
		z.Store(got)
	})
	checkVectorShift[uint64, uint64]("ShiftLeftUint64x4", 64, 4, false, false, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint64, count []uint64, got []uint64) {
		z := ordinaryShiftLeftUint64x4(archsimd.LoadUint64x4(x), archsimd.LoadUint64x4(count))
		z.Store(got)
	})
	checkVectorShift[uint64, uint64]("ShiftLeftUint64x8", 64, 8, false, false, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint64, count []uint64, got []uint64) {
		z := ordinaryShiftLeftUint64x8(archsimd.LoadUint64x8(x), archsimd.LoadUint64x8(count))
		z.Store(got)
	})
	checkVectorShift[int16, uint16]("ShiftRightInt16x8", 16, 8, true, true, true, archsimd.X86.AVX512(), func(x []int16, count []uint16, got []int16) {
		z := ordinaryShiftRightInt16x8(archsimd.LoadInt16x8(x), archsimd.LoadUint16x8(count))
		z.Store(got)
	})
	checkVectorShift[int16, uint16]("ShiftRightInt16x16", 16, 16, true, true, archsimd.X86.AVX(), archsimd.X86.AVX512(), func(x []int16, count []uint16, got []int16) {
		z := ordinaryShiftRightInt16x16(archsimd.LoadInt16x16(x), archsimd.LoadUint16x16(count))
		z.Store(got)
	})
	checkVectorShift[int16, uint16]("ShiftRightInt16x32", 16, 32, true, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int16, count []uint16, got []int16) {
		z := ordinaryShiftRightInt16x32(archsimd.LoadInt16x32(x), archsimd.LoadUint16x32(count))
		z.Store(got)
	})
	checkVectorShift[int32, uint32]("ShiftRightInt32x4", 32, 4, true, true, true, archsimd.X86.AVX2(), func(x []int32, count []uint32, got []int32) {
		z := ordinaryShiftRightInt32x4(archsimd.LoadInt32x4(x), archsimd.LoadUint32x4(count))
		z.Store(got)
	})
	checkVectorShift[int32, uint32]("ShiftRightInt32x8", 32, 8, true, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []int32, count []uint32, got []int32) {
		z := ordinaryShiftRightInt32x8(archsimd.LoadInt32x8(x), archsimd.LoadUint32x8(count))
		z.Store(got)
	})
	checkVectorShift[int32, uint32]("ShiftRightInt32x16", 32, 16, true, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int32, count []uint32, got []int32) {
		z := ordinaryShiftRightInt32x16(archsimd.LoadInt32x16(x), archsimd.LoadUint32x16(count))
		z.Store(got)
	})
	checkVectorShift[int64, uint64]("ShiftRightInt64x2", 64, 2, true, true, true, archsimd.X86.AVX512(), func(x []int64, count []uint64, got []int64) {
		z := ordinaryShiftRightInt64x2(archsimd.LoadInt64x2(x), archsimd.LoadUint64x2(count))
		z.Store(got)
	})
	checkVectorShift[int64, uint64]("ShiftRightInt64x4", 64, 4, true, true, archsimd.X86.AVX(), archsimd.X86.AVX512(), func(x []int64, count []uint64, got []int64) {
		z := ordinaryShiftRightInt64x4(archsimd.LoadInt64x4(x), archsimd.LoadUint64x4(count))
		z.Store(got)
	})
	checkVectorShift[int64, uint64]("ShiftRightInt64x8", 64, 8, true, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []int64, count []uint64, got []int64) {
		z := ordinaryShiftRightInt64x8(archsimd.LoadInt64x8(x), archsimd.LoadUint64x8(count))
		z.Store(got)
	})
	checkVectorShift[uint16, uint16]("ShiftRightUint16x8", 16, 8, false, true, true, archsimd.X86.AVX512(), func(x []uint16, count []uint16, got []uint16) {
		z := ordinaryShiftRightUint16x8(archsimd.LoadUint16x8(x), archsimd.LoadUint16x8(count))
		z.Store(got)
	})
	checkVectorShift[uint16, uint16]("ShiftRightUint16x16", 16, 16, false, true, archsimd.X86.AVX(), archsimd.X86.AVX512(), func(x []uint16, count []uint16, got []uint16) {
		z := ordinaryShiftRightUint16x16(archsimd.LoadUint16x16(x), archsimd.LoadUint16x16(count))
		z.Store(got)
	})
	checkVectorShift[uint16, uint16]("ShiftRightUint16x32", 16, 32, false, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint16, count []uint16, got []uint16) {
		z := ordinaryShiftRightUint16x32(archsimd.LoadUint16x32(x), archsimd.LoadUint16x32(count))
		z.Store(got)
	})
	checkVectorShift[uint32, uint32]("ShiftRightUint32x4", 32, 4, false, true, true, archsimd.X86.AVX2(), func(x []uint32, count []uint32, got []uint32) {
		z := ordinaryShiftRightUint32x4(archsimd.LoadUint32x4(x), archsimd.LoadUint32x4(count))
		z.Store(got)
	})
	checkVectorShift[uint32, uint32]("ShiftRightUint32x8", 32, 8, false, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint32, count []uint32, got []uint32) {
		z := ordinaryShiftRightUint32x8(archsimd.LoadUint32x8(x), archsimd.LoadUint32x8(count))
		z.Store(got)
	})
	checkVectorShift[uint32, uint32]("ShiftRightUint32x16", 32, 16, false, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint32, count []uint32, got []uint32) {
		z := ordinaryShiftRightUint32x16(archsimd.LoadUint32x16(x), archsimd.LoadUint32x16(count))
		z.Store(got)
	})
	checkVectorShift[uint64, uint64]("ShiftRightUint64x2", 64, 2, false, true, true, archsimd.X86.AVX2(), func(x []uint64, count []uint64, got []uint64) {
		z := ordinaryShiftRightUint64x2(archsimd.LoadUint64x2(x), archsimd.LoadUint64x2(count))
		z.Store(got)
	})
	checkVectorShift[uint64, uint64]("ShiftRightUint64x4", 64, 4, false, true, archsimd.X86.AVX(), archsimd.X86.AVX2(), func(x []uint64, count []uint64, got []uint64) {
		z := ordinaryShiftRightUint64x4(archsimd.LoadUint64x4(x), archsimd.LoadUint64x4(count))
		z.Store(got)
	})
	checkVectorShift[uint64, uint64]("ShiftRightUint64x8", 64, 8, false, true, archsimd.X86.AVX512(), archsimd.X86.AVX512(), func(x []uint64, count []uint64, got []uint64) {
		z := ordinaryShiftRightUint64x8(archsimd.LoadUint64x8(x), archsimd.LoadUint64x8(count))
		z.Store(got)
	})
}
