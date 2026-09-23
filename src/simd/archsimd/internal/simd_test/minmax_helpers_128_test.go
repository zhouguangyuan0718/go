// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd && (amd64 || arm64 || wasm)

package simd_test

import "simd/archsimd"

// Compute both operand orders in one function to exercise CSE.
//
//go:noinline
func floatMinMax32x4(a, b []float32, out [4][]float32, reverse, masked bool) {
	x, y := archsimd.LoadFloat32x4(a), archsimd.LoadFloat32x4(b)
	var minXY, minYX, maxXY, maxYX archsimd.Float32x4
	if masked {
		mask := floatMinMaxMask32x4()
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
		mask := floatMinMaxMask64x2()
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
