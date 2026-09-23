// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && (amd64 || arm64 || wasm)

package simd_test

import (
	"math"
	"runtime"
	"testing"
)

// AMD64 selects the second operand for NaNs and equal zeros. ARM64 propagates
// NaNs, preferring the first operand when both are quiet NaNs. Wasm uses the
// ARM64 expectations for numbers and signed zeros, but permits any NaN payload.
type floatMinMaxCase struct {
	name                 string
	x, y                 uint64
	wantAMD64, wantARM64 [4]uint64 // Min(x,y), Min(y,x), Max(x,y), Max(y,x)
}

func floatMinMax32Cases() []floatMinMaxCase {
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
	return []floatMinMaxCase{
		{"NaN-number", qnan1, one, [4]uint64{one, qnan1, one, qnan1}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-NaN", one, qnan1, [4]uint64{qnan1, one, qnan1, one}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan2, qnan1, qnan2, qnan1}, [4]uint64{qnan1, qnan2, qnan1, qnan2}},
		{"sNaN-number", snan, one, [4]uint64{one, snan, one, snan}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-sNaN", one, snan, [4]uint64{snan, one, snan, one}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"negative-zero", negZero, 0, [4]uint64{0, negZero, 0, negZero}, [4]uint64{negZero, negZero, 0, 0}},
		{"positive-zero", 0, negZero, [4]uint64{negZero, 0, negZero, 0}, [4]uint64{negZero, negZero, 0, 0}},
		{"finite", two, one, [4]uint64{one, one, two, two}, [4]uint64{one, one, two, two}},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}, [4]uint64{negInf, negInf, inf, inf}},
	}
}

func floatMinMax64Cases() []floatMinMaxCase {
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
	return []floatMinMaxCase{
		{"NaN-number", qnan1, one, [4]uint64{one, qnan1, one, qnan1}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-NaN", one, qnan1, [4]uint64{qnan1, one, qnan1, one}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"NaN-payloads", qnan1, qnan2, [4]uint64{qnan2, qnan1, qnan2, qnan1}, [4]uint64{qnan1, qnan2, qnan1, qnan2}},
		{"sNaN-number", snan, one, [4]uint64{one, snan, one, snan}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"number-sNaN", one, snan, [4]uint64{snan, one, snan, one}, [4]uint64{qnan1, qnan1, qnan1, qnan1}},
		{"negative-zero", negZero, 0, [4]uint64{0, negZero, 0, negZero}, [4]uint64{negZero, negZero, 0, 0}},
		{"positive-zero", 0, negZero, [4]uint64{negZero, 0, negZero, 0}, [4]uint64{negZero, negZero, 0, 0}},
		{"finite", two, one, [4]uint64{one, one, two, two}, [4]uint64{one, one, two, two}},
		{"infinities", inf, negInf, [4]uint64{negInf, negInf, inf, inf}, [4]uint64{negInf, negInf, inf, inf}},
	}
}

type floatMinMaxConfig struct {
	allowMasked bool
	allowAnyNaN bool
}

func testFloatMinMax[T float](t *testing.T, lanes int, cases []floatMinMaxCase,
	fromBits func(uint64) T, toBits func(T) uint64,
	run func([]T, []T, [4][]T, bool, bool), config floatMinMaxConfig) {
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
			if form.masked && !config.allowMasked {
				t.Skip("requires AVX512")
			}
			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					wantResults := c.wantARM64
					if runtime.GOARCH == "amd64" {
						wantResults = c.wantAMD64
					}
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
							want := wantResults[j]
							if form.masked && i%2 != 0 {
								want = 0
							}
							if config.allowAnyNaN && math.IsNaN(float64(fromBits(want))) {
								if !math.IsNaN(float64(v)) {
									t.Errorf("%s lane %d: got %#x, want NaN", name, i, toBits(v))
								}
								continue
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
