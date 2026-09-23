// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && (amd64 || arm64 || wasm)

package simd_test

import (
	"math"
	"testing"
)

type floatMinMaxCase struct {
	name string
	x, y uint64
	want [4]uint64 // Min(x,y), Min(y,x), Max(x,y), Max(y,x)
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
