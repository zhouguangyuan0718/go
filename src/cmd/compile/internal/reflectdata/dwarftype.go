// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectdata

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"strings"
)

type DwarfType struct {
	Type *types.Type
}

func (d DwarfType) DwarfName() string {
	name := types.TypeSymName(d.Type)
	return strings.Replace(name, `"".`, objabi.PathToPrefix(base.Ctxt.Pkgpath)+".", -1)
}

func (d DwarfType) Name() string {
	return types.TypeSymName(d.Type)
}

func (d DwarfType) Size() int64 {
	return d.Type.Size()
}

func (d DwarfType) Kind() objabi.SymKind {
	return objabi.SymKind(kinds[d.Type.Kind()])
}

func (d DwarfType) RuntimeType() dwarf.Sym {
	// don't use TypeLinksym, we don't expect dump
	// a runtime type because of dwarf generation.
	return types.TypeSym(d.Type).Linksym()
}

func (d DwarfType) Key() dwarf.Type {
	return DwarfType{Type: d.Type.Key()}
}

func (d DwarfType) Elem() dwarf.Type {
	return DwarfType{d.Type.Elem()}
}

func (d DwarfType) NumElem() int64 {
	if d.Type.IsArray() {
		return d.Type.NumElem()
	}
	if d.Type.IsStruct() {
		return int64(d.Type.NumFields())
	}
	if d.Type.Kind() == types.TFUNC {
		return int64(d.Type.NumParams())
	}
	panic("unreachable")
}

func (d DwarfType) NumResult() int64 {
	return int64(d.Type.NumResults())
}

func (d DwarfType) IsDDD() bool {
	return d.Type.IsVariadic()
}

func (d DwarfType) FieldName(g dwarf.FieldsGroup, i int) string {
	switch g {
	case dwarf.GroupFields:
		return d.Type.FieldName(i)
	case dwarf.GroupParams:
		return DwarfType{d.Type.Params().FieldType(i)}.DwarfName()
	case dwarf.GroupResults:
		return DwarfType{d.Type.Results().FieldType(i)}.DwarfName()
	}
	panic("unreachable")
}

func (d DwarfType) FieldType(g dwarf.FieldsGroup, i int) dwarf.Type {
	switch g {
	case dwarf.GroupFields:
		return DwarfType{d.Type.FieldType(i)}
	case dwarf.GroupParams:
		return DwarfType{d.Type.Params().FieldType(i)}
	case dwarf.GroupResults:
		return DwarfType{d.Type.Results().FieldType(i)}
	}
	panic("unreachable")
}

func (d DwarfType) FieldIsEmbed(i int) bool {
	return d.Type.Field(i).Embedded != 0
}

func (d DwarfType) FieldOffset(i int) int64 {
	return d.Type.Field(i).Offset
}

func (d DwarfType) IsEface() bool {
	return d.Type.IsEmptyInterface()
}
