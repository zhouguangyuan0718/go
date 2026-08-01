// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"strings"
	"testing"

	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
)

func semanticTestSignature(params, results []*types.Type) *types.Type {
	fields := func(typs []*types.Type) []*types.Field {
		out := make([]*types.Field, len(typs))
		for i, typ := range typs {
			out[i] = types.NewField(src.NoXPos, nil, typ)
		}
		return out
	}
	return types.NewSignature(nil, fields(params), fields(results))
}

func TestAuxCallSemanticSignatureCompatibility(t *testing.T) {
	config := testConfigARM64(t).config
	physical := config.ABI1.ABIAnalyzeTypes(
		[]*types.Type{config.Types.Uintptr},
		[]*types.Type{config.Types.Uintptr},
	)
	compatible := semanticTestSignature(
		[]*types.Type{config.Types.BytePtr},
		[]*types.Type{config.Types.BytePtr},
	)
	aux := StaticAuxCallWithSignature(&obj.LSym{Name: "runtime.semantic"}, physical, compatible)
	if got := aux.SemanticSignature(); got != compatible {
		t.Fatalf("SemanticSignature() = %v, want %v", got, compatible)
	}
	if err := aux.ValidateSemanticSignature(); err != nil {
		t.Fatalf("layout-compatible pointer signature rejected: %v", err)
	}
}

func TestAuxCallSemanticSignatureRejectsIncompatibleABI(t *testing.T) {
	config := testConfigARM64(t).config
	physical := config.ABI1.ABIAnalyzeTypes([]*types.Type{config.Types.Uintptr}, nil)
	tests := []struct {
		name string
		sig  *types.Type
		want string
	}{
		{
			name: "parameter count",
			sig:  semanticTestSignature(nil, nil),
			want: "semantic parameter count 0 does not match physical count 1",
		},
		{
			name: "size",
			sig:  semanticTestSignature([]*types.Type{config.Types.UInt32}, nil),
			want: "semantic parameter 0 size 4 does not match physical size 8",
		},
		{
			name: "register class",
			sig:  semanticTestSignature([]*types.Type{config.Types.Float64}, nil),
			want: "semantic parameter 0 register 0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aux := StaticAuxCallWithSignature(&obj.LSym{Name: "runtime.malformed"}, physical, test.sig)
			err := aux.ValidateSemanticSignature()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateSemanticSignature() error = %v, want substring %q", err, test.want)
			}
		})
	}
}
