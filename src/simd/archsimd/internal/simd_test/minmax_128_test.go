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

type floatMinMax128Case struct {
	name string
	x, y uint64
	want [4]uint64 // Min(x,y), Min(y,x), Max(x,y), Max(y,x)
	nan  bool      // All four results must be NaN; their payloads are unspecified.
}

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
	cases := []floatMinMax128Case{
		{"negative-zero", negZero, 0, [4]uint64{negZero, negZero, 0, 0}, false},
		{"positive-zero", 0, negZero, [4]uint64{negZero, negZero, 0, 0}, false},
		{"finite", two, one, [4]uint64{one, one, two, two}, false},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}, false},
	}
	if runtime.GOARCH == "arm64" {
		// ARM64 Min/Max propagate quiet NaNs. With two quiet NaNs,
		// the first operand's payload is returned.
		cases = append(cases,
			floatMinMax128Case{"NaN-number", qnan1, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
			floatMinMax128Case{"number-NaN", one, qnan1, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
			floatMinMax128Case{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan1, qnan2, qnan1, qnan2}, false},
			floatMinMax128Case{"sNaN-number", snan, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
			floatMinMax128Case{"number-sNaN", one, snan, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
		)
	} else {
		// WebAssembly Min/Max propagate NaNs, but do not specify their payloads.
		cases = append(cases,
			floatMinMax128Case{name: "NaN-number", x: qnan1, y: one, nan: true},
			floatMinMax128Case{name: "number-NaN", x: one, y: qnan1, nan: true},
			floatMinMax128Case{name: "NaN-payloads", x: qnan1, y: qnan2, nan: true},
			floatMinMax128Case{name: "sNaN-number", x: snan, y: one, nan: true},
			floatMinMax128Case{name: "number-sNaN", x: one, y: snan, nan: true},
		)
	}
	fromBits := func(x uint64) float32 { return math.Float32frombits(uint32(x)) }
	toBits := func(x float32) uint64 { return uint64(math.Float32bits(x)) }
	testFloatMinMax128(t, 4, cases, fromBits, toBits, floatMinMax32x4_128)
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
	cases := []floatMinMax128Case{
		{"negative-zero", negZero, 0, [4]uint64{negZero, negZero, 0, 0}, false},
		{"positive-zero", 0, negZero, [4]uint64{negZero, negZero, 0, 0}, false},
		{"finite", two, one, [4]uint64{one, one, two, two}, false},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}, false},
	}
	if runtime.GOARCH == "arm64" {
		cases = append(cases,
			floatMinMax128Case{"NaN-number", qnan1, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
			floatMinMax128Case{"number-NaN", one, qnan1, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
			floatMinMax128Case{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan1, qnan2, qnan1, qnan2}, false},
			floatMinMax128Case{"sNaN-number", snan, one, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
			floatMinMax128Case{"number-sNaN", one, snan, [4]uint64{qnan1, qnan1, qnan1, qnan1}, false},
		)
	} else {
		cases = append(cases,
			floatMinMax128Case{name: "NaN-number", x: qnan1, y: one, nan: true},
			floatMinMax128Case{name: "number-NaN", x: one, y: qnan1, nan: true},
			floatMinMax128Case{name: "NaN-payloads", x: qnan1, y: qnan2, nan: true},
			floatMinMax128Case{name: "sNaN-number", x: snan, y: one, nan: true},
			floatMinMax128Case{name: "number-sNaN", x: one, y: snan, nan: true},
		)
	}
	testFloatMinMax128(t, 2, cases, math.Float64frombits, math.Float64bits, floatMinMax64x2_128)
}

func testFloatMinMax128[T float](t *testing.T, lanes int, cases []floatMinMax128Case,
	fromBits func(uint64) T, toBits func(T) uint64,
	run func([]T, []T, [4][]T, bool, bool)) {
	t.Helper()
	for _, form := range []struct {
		name            string
		reverse, masked bool
	}{
		{"forward", false, false},
		{"reverse", true, false},
		{"masked-forward", false, true},
		{"masked-reverse", true, true},
	} {
		t.Run(form.name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					x, y := make([]T, lanes), make([]T, lanes)
					var out [4][]T
					for i := range out {
						out[i] = make([]T, lanes)
					}
					for i := range x {
						x[i], y[i] = fromBits(c.x), fromBits(c.y)
					}
					run(x, y, out, form.reverse, form.masked)
					for j, name := range []string{"Min(x,y)", "Min(y,x)", "Max(x,y)", "Max(y,x)"} {
						for i, v := range out[j] {
							if form.masked && i%2 != 0 {
								if got := toBits(v); got != 0 {
									t.Errorf("%s lane %d: got %#x, want 0", name, i, got)
								}
								continue
							}
							if c.nan {
								if !math.IsNaN(float64(v)) {
									t.Errorf("%s lane %d: got %#x, want NaN", name, i, toBits(v))
								}
							} else if got := toBits(v); got != c.want[j] {
								t.Errorf("%s lane %d: got %#x, want %#x", name, i, got, c.want[j])
							}
						}
					}
				})
			}
		})
	}
}

//go:noinline
func floatMinMax32x4_128(a, b []float32, out [4][]float32, reverse, masked bool) {
	x, y := archsimd.LoadFloat32x4(a), archsimd.LoadFloat32x4(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float32x4
	if reverse {
		minYX, minXY = y.Min(x), x.Min(y)
		maxYX, maxXY = y.Max(x), x.Max(y)
	} else {
		minXY, minYX = x.Min(y), y.Min(x)
		maxXY, maxYX = x.Max(y), y.Max(x)
	}
	if masked {
		mask := archsimd.LoadInt32x4([]int32{1, 0, 1, 0}).Greater(archsimd.LoadInt32x4([]int32{0, 0, 0, 0}))
		minXY, minYX = minXY.Masked(mask), minYX.Masked(mask)
		maxXY, maxYX = maxXY.Masked(mask), maxYX.Masked(mask)
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
	if reverse {
		minYX, minXY = y.Min(x), x.Min(y)
		maxYX, maxXY = y.Max(x), x.Max(y)
	} else {
		minXY, minYX = x.Min(y), y.Min(x)
		maxXY, maxYX = x.Max(y), y.Max(x)
	}
	if masked {
		mask := archsimd.LoadInt64x2([]int64{1, 0}).Greater(archsimd.LoadInt64x2([]int64{0, 0}))
		minXY, minYX = minXY.Masked(mask), minYX.Masked(mask)
		maxXY, maxYX = maxXY.Masked(mask), maxYX.Masked(mask)
	}
	minXY.Store(out[0])
	minYX.Store(out[1])
	maxXY.Store(out[2])
	maxYX.Store(out[3])
}
