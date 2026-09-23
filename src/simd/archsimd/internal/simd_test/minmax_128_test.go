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
	fromBits := func(x uint64) float32 { return math.Float32frombits(uint32(x)) }
	toBits := func(x float32) uint64 { return uint64(math.Float32bits(x)) }
	testFloatMinMax(t, 4, floatMinMax32Cases(), fromBits, toBits, floatMinMax32x4,
		floatMinMaxConfig{allowMasked: true, allowAnyNaN: runtime.GOARCH == "wasm"})
}

func TestFloatMinMax64_128(t *testing.T) {
	testFloatMinMax(t, 2, floatMinMax64Cases(), math.Float64frombits, math.Float64bits, floatMinMax64x2,
		floatMinMaxConfig{allowMasked: true, allowAnyNaN: runtime.GOARCH == "wasm"})
}

func floatMinMaxMask32x4() archsimd.Mask32x4 {
	return archsimd.LoadInt32x4([]int32{1, 0, 1, 0}).Greater(archsimd.LoadInt32x4([]int32{0, 0, 0, 0}))
}

func floatMinMaxMask64x2() archsimd.Mask64x2 {
	return archsimd.LoadInt64x2([]int64{1, 0}).Greater(archsimd.LoadInt64x2([]int64{0, 0}))
}
