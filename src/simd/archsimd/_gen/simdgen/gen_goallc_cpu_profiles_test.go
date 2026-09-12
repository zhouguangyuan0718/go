// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"simd/archsimd/_gen/internal/goallccpu"
	"testing"
)

func TestGoALLCCPUProfileAliases(t *testing.T) {
	// Every feature the native XED decoder can return must have an explicit
	// LLVM descriptor policy. Unknown future extensions must not inherit the
	// ordinary AVX512 profile merely because their name has that prefix.
	for _, feature := range cpuFeatureMap {
		if feature != "ignore" {
			goallccpu.ProfileForSIMD("amd64", feature)
		}
	}
	for _, test := range []struct{ feature, want string }{
		{"AVX", "x86.avx"}, {"AVX2", "x86.avx2"}, {"AVX512", "x86.avx512"},
		{"AVX512VBMI", "x86.avx512vbmi"}, {"AVX512BITALG", "x86.avx512bitalg"},
		{"AVX512VPOPCNTDQ", "x86.avx512vpopcntdq"}, {"FMA", "x86.fma"}, {"SHA", ""}, {"", ""},
	} {
		if got := goallccpu.ProfileForSIMD("amd64", test.feature); got != test.want {
			t.Errorf("%s = %q, want %q", test.feature, got, test.want)
		}
	}
	if got := goallccpu.ProfileForSIMD("arm64", "AVX512VBMI"); got != "" {
		t.Errorf("cross-architecture profile %q", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("unknown AVX512 extension was accepted")
		}
	}()
	goallccpu.ProfileForSIMD("amd64", "AVX512FUTURE")
}
