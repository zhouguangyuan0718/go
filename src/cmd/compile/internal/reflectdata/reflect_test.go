// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectdata

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/sys"
	"testing"
)

func init() {
	types.PtrSize = 8
	types.RegSize = 8
	types.MaxWidth = 1 << 50
	base.Ctxt = &obj.Link{Arch: &obj.LinkArch{Arch: &sys.Arch{Alignment: 1, CanMergeLoads: true}}}
	typecheck.InitUniverse()
}

func TestTypeLinksymCanonicalizesPredeclaredAliases(t *testing.T) {
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
			aliasSym := TypeLinksym(test.alias)
			underlyingSym := TypeLinksym(test.underlying)
			if aliasSym != underlyingSym {
				t.Fatalf("alias and underlying type use different linker symbols: %q and %q", aliasSym.Name, underlyingSym.Name)
			}
			if got := aliasSym.TypeInfo().Type; got != test.underlying {
				t.Fatalf("TypeInfo.Type = %v, want canonical %v", got, test.underlying)
			}
		})
	}
}
