// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type integer interface {
	int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
}

type unsigned interface {
	uint8 | uint16 | uint32 | uint64
}

func laneMask(bits int) uint64 {
	return ^uint64(0) >> (64 - bits)
}

// Keep every bit of scalar counts: truncating a count before testing it against
// the lane width can turn an overshift into a valid shift.
func shiftCounts(bits int) []uint64 {
	var counts []uint64
	for count := 0; count <= 2*bits+1; count++ {
		counts = append(counts, uint64(count))
	}
	for bit := 0; bit < 64; bit++ {
		counts = append(counts, uint64(1)<<bit, ^(uint64(1) << bit))
	}
	return append(counts, ^uint64(0))
}

func shiftInputs(bits int) []uint64 {
	mask := laneMask(bits)
	sign := uint64(1) << (bits - 1)
	values := []uint64{
		0, 1, 2, mask, mask - 1, sign, sign - 1, sign + 1,
		0x5555555555555555 & mask, 0xaaaaaaaaaaaaaaaa & mask,
		0x0123456789abcdef & mask, 0xfedcba9876543210 & mask,
	}
	for bit := 0; bit < bits; bit++ {
		values = append(values, uint64(1)<<bit, mask^(uint64(1)<<bit))
	}
	return values
}

// The oracle uses ordinary Go shifts, including Go's defined overshift
// behavior. Sign extension happens before the signed right shift, and all
// results are compared as lane-width bit patterns.
//
//go:noinline
func expectedShift(value, count uint64, bits int, signed, right bool) uint64 {
	mask := laneMask(bits)
	value &= mask
	if !right {
		return (value << count) & mask
	}
	if signed && value&(uint64(1)<<(bits-1)) != 0 {
		return uint64(int64(value|^mask)>>count) & mask
	}
	return value >> count
}

func beginCase(abi, enabled bool) bool {
	if !abi {
		skippedCases++
		return false
	}
	if enabled {
		enabledCases++
	} else {
		fallbackCases++
	}
	return true
}

func checkScalarShift[T integer](name string, bits, n int, signed, right, abi, enabled bool, shift func([]T, uint64, []T)) {
	if !beginCase(abi, enabled) {
		return
	}
	x, got := make([]T, n), make([]T, n)
	inputs, counts := shiftInputs(bits), shiftCounts(bits)
	mask := laneMask(bits)
	// Rotate inputs so every lane sees every data edge for every count.
	for round := range inputs {
		for lane := range x {
			x[lane] = T(inputs[(round+lane)%len(inputs)])
		}
		for _, count := range counts {
			shift(x, count, got)
			checkedScalarCounts++
			for lane, value := range x {
				var want uint64
				if enabled {
					want = expectedShift(uint64(value), count, bits, signed, right)
				}
				if actual := uint64(got[lane]) & mask; actual != want {
					panic(fmt.Sprintf("%s enabled=%v round=%d count=%#x lane=%d x=%#x got=%#x want=%#x",
						name, enabled, round, count, lane, uint64(value)&mask, actual, want))
				}
				checkedLanes++
			}
		}
	}
}

func checkVectorShift[T integer, I unsigned](name string, bits, n int, signed, right, abi, enabled bool, shift func([]T, []I, []T)) {
	if !beginCase(abi, enabled) {
		return
	}
	x, got := make([]T, n), make([]T, n)
	indices := make([]I, n)
	inputs, counts := shiftInputs(bits), shiftCounts(bits)
	mask := laneMask(bits)
	for round := range inputs {
		for lane := range x {
			x[lane] = T(inputs[(round+lane)%len(inputs)])
		}
		// Uniform controls isolate count semantics. Mixed controls distinguish
		// per-lane shifts from scalar broadcasts and exercise lane ordering.
		// Both layouts put every control in every lane for every data edge.
		for layout := 0; layout < 2; layout++ {
			for control := range counts {
				for lane := range indices {
					index := control
					if layout != 0 {
						index = (control + lane*(bits+1)) % len(counts)
					}
					indices[lane] = I(counts[index] & mask)
				}
				shift(x, indices, got)
				checkedVectorCounts++
				for lane, value := range x {
					count := uint64(indices[lane])
					var want uint64
					if enabled {
						want = expectedShift(uint64(value), count, bits, signed, right)
					}
					if actual := uint64(got[lane]) & mask; actual != want {
						panic(fmt.Sprintf("%s enabled=%v round=%d layout=%d count=%#x lane=%d x=%#x got=%#x want=%#x",
							name, enabled, round, layout, count, lane, uint64(value)&mask, actual, want))
					}
					checkedLanes++
				}
			}
		}
	}
}

// These calls put literal counts in the source operation itself. A dynamic
// count whose runtime value happens to be zero cannot exercise compiler
// constant folding that returns an existing argument or instruction.
func checkConstantShift[T integer](name string, bits, n int, count uint64, signed, right, add, abi, enabled bool, shift func([]T, []T, []T)) {
	if !abi {
		constantSkippedCases++
		return
	}
	if enabled {
		constantEnabledCases++
	} else {
		constantFallbackCases++
	}
	x, y, got := make([]T, n), make([]T, n), make([]T, n)
	inputs, mask := shiftInputs(bits), laneMask(bits)
	for round := range inputs {
		for lane := range x {
			x[lane] = T(inputs[(round+lane)%len(inputs)])
			y[lane] = T(inputs[(round+lane+3)%len(inputs)])
		}
		shift(x, y, got)
		checkedConstantVectors++
		for lane, value := range x {
			var want uint64
			if enabled {
				bitsValue := uint64(value)
				if add {
					bitsValue += uint64(y[lane])
				}
				want = expectedShift(bitsValue, count, bits, signed, right)
			}
			if actual := uint64(got[lane]) & mask; actual != want {
				panic(fmt.Sprintf("%s enabled=%v round=%d constant=%#x lane=%d x=%#x y=%#x got=%#x want=%#x",
					name, enabled, round, count, lane, uint64(value)&mask, uint64(y[lane])&mask, actual, want))
			}
		}
	}
}
