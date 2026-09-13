// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && amd64

package simd_test

import (
	"math"
	"simd/archsimd"
	"testing"
)

// Unlike the Go builtins, x86 floating-point Min and Max return the second
// operand when either operand is NaN or both are zero. Compare bits to check
// signed zeros and NaN payloads, including when opposite operand orders occur
// in the same function.
type floatMinMaxCase struct {
	name string
	x, y uint64
	want [4]uint64 // Min(x,y), Min(y,x), Max(x,y), Max(y,x)
}

func TestFloatMinMax32AMD64(t *testing.T) {
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
	cases := []floatMinMaxCase{
		{"NaN-number", qnan1, one, [4]uint64{one, qnan1, one, qnan1}},
		{"number-NaN", one, qnan1, [4]uint64{qnan1, one, qnan1, one}},
		{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan2, qnan1, qnan2, qnan1}},
		{"sNaN-number", snan, one, [4]uint64{one, snan, one, snan}},
		{"number-sNaN", one, snan, [4]uint64{snan, one, snan, one}},
		{"negative-zero", negZero, 0, [4]uint64{0, negZero, 0, negZero}},
		{"positive-zero", 0, negZero, [4]uint64{negZero, 0, negZero, 0}},
		{"finite", two, one, [4]uint64{one, one, two, two}},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}},
	}
	fromBits := func(x uint64) float32 { return math.Float32frombits(uint32(x)) }
	toBits := func(x float32) uint64 { return uint64(math.Float32bits(x)) }
	for _, tc := range []struct {
		name  string
		lanes int
		run   func([]float32, []float32, [4][]float32, bool, bool)
	}{
		{"Float32x4", 4, floatMinMax32x4},
		{"Float32x8", 8, floatMinMax32x8},
		{"Float32x16", 16, floatMinMax32x16},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.lanes == 16 && !archsimd.X86.AVX512() {
				t.Skip("requires AVX512")
			}
			testFloatMinMax(t, tc.lanes, cases, fromBits, toBits, tc.run)
		})
	}
}

func TestFloatMinMax64AMD64(t *testing.T) {
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
		{"NaN-number", qnan1, one, [4]uint64{one, qnan1, one, qnan1}},
		{"number-NaN", one, qnan1, [4]uint64{qnan1, one, qnan1, one}},
		{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan2, qnan1, qnan2, qnan1}},
		{"sNaN-number", snan, one, [4]uint64{one, snan, one, snan}},
		{"number-sNaN", one, snan, [4]uint64{snan, one, snan, one}},
		{"negative-zero", negZero, 0, [4]uint64{0, negZero, 0, negZero}},
		{"positive-zero", 0, negZero, [4]uint64{negZero, 0, negZero, 0}},
		{"finite", two, one, [4]uint64{one, one, two, two}},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}},
	}
	for _, tc := range []struct {
		name  string
		lanes int
		run   func([]float64, []float64, [4][]float64, bool, bool)
	}{
		{"Float64x2", 2, floatMinMax64x2},
		{"Float64x4", 4, floatMinMax64x4},
		{"Float64x8", 8, floatMinMax64x8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.lanes == 8 && !archsimd.X86.AVX512() {
				t.Skip("requires AVX512")
			}
			testFloatMinMax(t, tc.lanes, cases, math.Float64frombits, math.Float64bits, tc.run)
		})
	}
}

func testFloatMinMax[T float](t *testing.T, lanes int, cases []floatMinMaxCase,
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
			if form.masked && !archsimd.X86.AVX512() {
				t.Skip("requires AVX512")
			}
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
							want := c.want[j]
							if form.masked && i%2 != 0 {
								want = 0
							}
							if got := toBits(v); got != want {
								t.Errorf("%s lane %d: got %#x, want %#x", name, i, got, want)
							}
						}
					}
				})
			}
		})
	}
}

//go:noinline
func floatMinMax32x4(a, b []float32, out [4][]float32, reverse, masked bool) {
	x, y := archsimd.LoadFloat32x4(a), archsimd.LoadFloat32x4(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float32x4
	if masked {
		mask := archsimd.Mask32x4FromBits(0x5)
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
func floatMinMax32x8(a, b []float32, out [4][]float32, reverse, masked bool) {
	x, y := archsimd.LoadFloat32x8(a), archsimd.LoadFloat32x8(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float32x8
	if masked {
		mask := archsimd.Mask32x8FromBits(0x55)
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
func floatMinMax32x16(a, b []float32, out [4][]float32, reverse, masked bool) {
	x, y := archsimd.LoadFloat32x16(a), archsimd.LoadFloat32x16(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float32x16
	if masked {
		mask := archsimd.Mask32x16FromBits(0x5555)
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
func floatMinMax64x2(a, b []float64, out [4][]float64, reverse, masked bool) {
	x, y := archsimd.LoadFloat64x2(a), archsimd.LoadFloat64x2(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float64x2
	if masked {
		mask := archsimd.Mask64x2FromBits(0x1)
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
func floatMinMax64x4(a, b []float64, out [4][]float64, reverse, masked bool) {
	x, y := archsimd.LoadFloat64x4(a), archsimd.LoadFloat64x4(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float64x4
	if masked {
		mask := archsimd.Mask64x4FromBits(0x5)
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
func floatMinMax64x8(a, b []float64, out [4][]float64, reverse, masked bool) {
	x, y := archsimd.LoadFloat64x8(a), archsimd.LoadFloat64x8(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float64x8
	if masked {
		mask := archsimd.Mask64x8FromBits(0x55)
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
