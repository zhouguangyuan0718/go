// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && (arm64 || wasm)

package simd_test

import (
	"math"
	"runtime"
	"simd/archsimd"
	"testing"
)

func TestFloatMinMax32_128(t *testing.T) {
	const (
		one     = 0x3f800000
		two     = 0x40000000
		negZero = 0x80000000
		qnan1   = 0x7fc00001
		qnan2   = 0x7fc00002
		snan    = 0x7f800001
		inf     = 0x7f800000
		negInf  = 0xff800000
	)
	// ARM64 preserves the first payload when both operands are NaN.
	// Wasm permits any NaN payload, which the shared helper accounts for.
	cases := []floatMinMaxCase{
		{"NaN-number", qnan1, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-NaN", one, qnan1, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan1, qnan2, qnan1, qnan2}},
		{"sNaN-number", snan, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-sNaN", one, snan, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"negative-zero", negZero, 0, [4]uint64{negZero, negZero, 0, 0}},
		{"positive-zero", 0, negZero, [4]uint64{negZero, negZero, 0, 0}},
		{"finite", two, one, [4]uint64{one, one, two, two}},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}},
	}
	fromBits := func(x uint64) float32 { return math.Float32frombits(uint32(x)) }
	toBits := func(x float32) uint64 { return uint64(math.Float32bits(x)) }
	testFloatMinMax(t, 4, cases, fromBits, toBits, floatMinMax32x4_128,
		floatMinMaxConfig{allowMasked: true, allowAnyNaN: runtime.GOARCH == "wasm"})
}

func TestFloatMinMax64_128(t *testing.T) {
	const (
		one     = 0x3ff0000000000000
		two     = 0x4000000000000000
		negZero = 0x8000000000000000
		qnan1   = 0x7ff8000000000001
		qnan2   = 0x7ff8000000000002
		snan    = 0x7ff0000000000001
		inf     = 0x7ff0000000000000
		negInf  = 0xfff0000000000000
	)
	cases := []floatMinMaxCase{
		{"NaN-number", qnan1, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-NaN", one, qnan1, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan1, qnan2, qnan1, qnan2}},
		{"sNaN-number", snan, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-sNaN", one, snan, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"negative-zero", negZero, 0, [4]uint64{negZero, negZero, 0, 0}},
		{"positive-zero", 0, negZero, [4]uint64{negZero, negZero, 0, 0}},
		{"finite", two, one, [4]uint64{one, one, two, two}},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}},
	}
	testFloatMinMax(t, 2, cases, math.Float64frombits, math.Float64bits, floatMinMax64x2_128,
		floatMinMaxConfig{allowMasked: true, allowAnyNaN: runtime.GOARCH == "wasm"})
}

//go:noinline
func floatMinMax32x4_128(a, b []float32, out [4][]float32, reverse, masked bool) {
	x, y := archsimd.LoadFloat32x4(a), archsimd.LoadFloat32x4(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float32x4
	if masked {
		mask := archsimd.LoadInt32x4([]int32{1, 0, 1, 0}).Greater(archsimd.LoadInt32x4([]int32{0, 0, 0, 0}))
		if reverse {
			minYX, minXY = y.Min(x).Masked(mask), x.Min(y).Masked(mask)
			maxYX, maxXY = y.Max(x).Masked(mask), x.Max(y).Masked(mask)
		} else {
			minXY, minYX = x.Min(y).Masked(mask), y.Min(x).Masked(mask)
			maxXY, maxYX = x.Max(y).Masked(mask), y.Max(x).Masked(mask)
		}
	} else if reverse {
		minYX, minXY = y.Min(x), x.Min(y)
		maxYX, maxXY = y.Max(x), x.Max(y)
	} else {
		minXY, minYX = x.Min(y), y.Min(x)
		maxXY, maxYX = x.Max(y), y.Max(x)
	}
	minXY.Store(out[0])
	minYX.Store(out[1])
	maxXY.Store(out[2])
	maxYX.Store(out[3])
}

//go:noinline
func floatMinMax64x2_128(a, b []float64, out [4][]float64, reverse, masked bool) {
	x, y := archsimd.LoadFloat64x2(a), archsimd.LoadFloat64x2(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float64x2
	if masked {
		mask := archsimd.LoadInt64x2([]int64{1, 0}).Greater(archsimd.LoadInt64x2([]int64{0, 0}))
		if reverse {
			minYX, minXY = y.Min(x).Masked(mask), x.Min(y).Masked(mask)
			maxYX, maxXY = y.Max(x).Masked(mask), x.Max(y).Masked(mask)
		} else {
			minXY, minYX = x.Min(y).Masked(mask), y.Min(x).Masked(mask)
			maxXY, maxYX = x.Max(y).Masked(mask), y.Max(x).Masked(mask)
		}
	} else if reverse {
		minYX, minXY = y.Min(x), x.Min(y)
		maxYX, maxXY = y.Max(x), x.Max(y)
	} else {
		minXY, minYX = x.Min(y), y.Min(x)
		maxXY, maxYX = x.Max(y), y.Max(x)
	}
	minXY.Store(out[0])
	minYX.Store(out[1])
	maxXY.Store(out[2])
	maxYX.Store(out[3])
}
