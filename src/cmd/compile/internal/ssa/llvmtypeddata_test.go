// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"testing"

	"cmd/compile/internal/types"
	"cmd/internal/obj"
)

func TestLLVMDescriptorGoTypeCanonicalizesPredeclaredAliases(t *testing.T) {
	tests := []struct {
		name       string
		alias      *types.Type
		underlying *types.Type
	}{
		{"any", types.AnyType, types.Types[types.TINTER]},
		{"byte", types.ByteType, types.Types[types.TUINT8]},
		{"rune", types.RuneType, types.Types[types.TINT32]},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &obj.LSym{Name: "type:" + test.name}
			s.NewTypeInfo().Type = test.alias
			if got := llvmDescriptorGoType(s); got != test.underlying {
				t.Fatalf("llvmDescriptorGoType() = %v, want %v", got, test.underlying)
			}
		})
	}
}
