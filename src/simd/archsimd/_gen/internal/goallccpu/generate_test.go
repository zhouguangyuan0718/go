// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goallccpu

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestGeneratedFilesCurrent(t *testing.T) {
	files, err := generatedFiles()
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(sourceRoot(), name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is stale; regenerate with simd/archsimd/_gen", name)
		}
	}
}

func TestGenerateAndCheck(t *testing.T) {
	files, err := generatedFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("got %d outputs, want compiler/plugin/runtime only", len(files))
	}
	root := t.TempDir()
	for name := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := Generate(root, true); err == nil {
		t.Fatal("check accepted missing outputs")
	}
	if err := Generate(root, false); err != nil {
		t.Fatal(err)
	}
	if err := Generate(root, true); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(root, "src/runtime/cpuflags_goallc_gen.go")
	stale := []byte("stale")
	if err := os.WriteFile(name, stale, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Generate(root, true); err == nil {
		t.Fatal("check accepted stale output")
	}
	got, err := os.ReadFile(name)
	if err != nil || !bytes.Equal(got, stale) {
		t.Fatalf("check rewrote stale output: %q, %v", got, err)
	}
}

func TestRuntimeBitABI(t *testing.T) {
	// Existing object files contain these literal numbers. Do not regenerate
	// this compatibility fixture when adding new features to the registry.
	want := map[string]uint{
		"SSE3": 0, "SSSE3": 1, "SSE41": 2, "SSE42": 3,
		"AVX": 4, "FMA": 5, "INITIALIZED": 6, "POPCNT": 7,
		"ARM64LSE": 8, "AVX2": 9, "AVX512": 10,
		"AVX512BITALG": 11, "AVX512VPOPCNTDQ": 12, "AVX512VBMI": 13,
	}
	for _, f := range features {
		if bit, ok := want[f.Name]; ok {
			if f.Bit != bit {
				t.Errorf("%s changed ABI bit %d to %d", f.Name, bit, f.Bit)
			}
			delete(want, f.Name)
		} else if f.Bit <= 13 {
			t.Errorf("new feature %s reuses reserved bit %d", f.Name, f.Bit)
		}
	}
	if len(want) != 0 {
		t.Errorf("removed ABI features: %v", want)
	}
}

func TestPredicateAndCapabilities(t *testing.T) {
	r, err := resolve(features, profiles)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		bit     uint
		caps    uint64
		targets string
	}{
		{"SSE41", 2, 4, "sse4.1"},
		{"AVX", 4, 16, "avx"},
		{"AVX2", 9, 528, "avx,avx2"},
		{"AVX512", 10, 1552, "avx,avx2,avx512f,avx512cd,avx512bw,avx512dq,avx512vl"},
		{"AVX512BITALG", 11, 3600, "avx,avx2,avx512f,avx512cd,avx512bw,avx512dq,avx512vl,avx512bitalg"},
		{"AVX512VPOPCNTDQ", 12, 5648, "avx,avx2,avx512f,avx512cd,avx512bw,avx512dq,avx512vl,avx512vpopcntdq"},
		{"AVX512VBMI", 13, 9744, "avx,avx2,avx512f,avx512cd,avx512bw,avx512dq,avx512vl,avx512vbmi"},
		{"FMA", 5, 32, "fma"},
		{"POPCNT", 7, 128, "popcnt"},
		{"ARM64LSE", 8, 256, "lse"},
	} {
		f := r[test.name]
		if f.Bit != test.bit || f.Mask != test.caps || strings.Join(f.Targets, ",") != test.targets {
			t.Errorf("%s: got bit=%d caps=%d targets=%v, want %+v", test.name, f.Bit, f.Mask, f.Targets, test)
		}
	}
}

func TestRejectInvalidRegistry(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func([]feature, []profile)
	}{
		{"duplicate-bit", func(fs []feature, _ []profile) { fs[1].Bit = fs[0].Bit }},
		{"duplicate-feature", func(fs []feature, _ []profile) { fs[1].Name = fs[0].Name }},
		{"out-of-range-bit", func(fs []feature, _ []profile) { fs[1].Bit = 64 }},
		{"unknown-capability", func(fs []feature, _ []profile) { fs[1].Provides = []string{"future"} }},
		{"cyclic-capability", func(fs []feature, _ []profile) { fs[1].Provides = []string{fs[1].Name} }},
		{"cross-arch-capability", func(fs []feature, _ []profile) { fs[1].Provides = []string{"ARM64LSE"} }},
		{"duplicate-profile", func(_ []feature, ps []profile) { ps[1].Name = ps[0].Name }},
		{"unknown-feature", func(_ []feature, ps []profile) { ps[1].Feature = "future" }},
		{"duplicate-alias", func(_ []feature, ps []profile) { ps[0].SIMDAliases = []string{"AVX"} }},
		{"unprofiled-alias", func(_ []feature, ps []profile) { ps[0].SIMDAliases = []string{"SHA"} }},
		{"duplicate-runtime-guard", func(_ []feature, ps []profile) { ps[1].RuntimeGuard = ps[0].RuntimeGuard }},
	} {
		t.Run(test.name, func(t *testing.T) {
			fs, ps := slices.Clone(features), slices.Clone(profiles)
			test.edit(fs, ps)
			if _, err := resolve(fs, ps); err == nil {
				t.Fatal("accepted invalid registry")
			}
		})
	}
}

func sourceRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../.."))
}
