//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"sort"
	"strings"

	"github.com/goallc/go-llvm"
)

// LowerGoObjTypeData lowers the complete linker-data closure rooted at runtime
// type descriptors into LLVM globals. The Go front end remains responsible for
// laying out the descriptors: this code only preserves the already-finalized
// LSym bytes, relocations, and GoObj symbol attributes in LLVM IR.
//
// Keeping this boundary at LSym is intentional. reflectdata is the source of
// truth for runtime layouts and is also used by the native backend; duplicating
// it here would make LLVM and native type descriptors drift independently.
func LowerGoObjTypeData() {
	data := make(map[*obj.LSym]bool, len(base.Ctxt.Data))
	for _, s := range base.Ctxt.Data {
		data[s] = true
	}

	closure := make(map[*obj.LSym]bool)
	var visit func(*obj.LSym)
	visit = func(s *obj.LSym) {
		if s == nil || closure[s] || !data[s] {
			return
		}
		closure[s] = true
		for _, r := range s.R {
			visit(r.Sym)
		}
	}
	for _, s := range base.Ctxt.Data {
		if s.TypeInfo() != nil {
			visit(s)
		}
	}

	if len(closure) == 0 {
		return
	}

	syms := make([]*obj.LSym, 0, len(closure))
	for s := range closure {
		syms = append(syms, s)
	}
	sort.Slice(syms, func(i, j int) bool { return syms[i].Name < syms[j].Name })
	lowerer := newLLVMDataLowerer(data)

	globals := make(map[*obj.LSym]llvm.Value, len(syms))
	for _, s := range syms {
		t := lowerer.dataType(s)
		g := CurrentModule.NamedGlobal(s.Name)
		if g.IsNil() {
			g = llvm.AddGlobal(CurrentModule, t, s.Name)
		} else if g.GlobalValueType() != t {
			base.Fatalf("conflicting LLVM global type for Go data symbol %s", s.Name)
		}
		g.SetSection(llvmDataSection(s))
		g.SetGlobalConstant(llvmDataIsReadOnly(s))
		if s.Align != 0 {
			g.SetAlignment(int(s.Align))
		}
		globals[s] = g
	}

	for _, s := range syms {
		g := globals[s]
		g.SetInitializer(lowerer.dataInitializer(s, globals))
		setGoObjDataFlags(g, s)
		setGoObjRelocMetadata(g, s)
		setGoObjKeepMetadata(g, s)
	}
}

func llvmDataType(s *obj.LSym) llvm.Type {
	fields := llvmDataFields(s, nil, nil)
	return GlobalCtxt.ConstStruct(fields, true).Type()
}

func llvmDataInitializer(s *obj.LSym, globals map[*obj.LSym]llvm.Value, data map[*obj.LSym]bool) llvm.Value {
	fields := llvmDataFields(s, globals, data)
	return GlobalCtxt.ConstStruct(fields, true)
}

// llvmDataFields models an LSym as a packed aggregate. Byte runs preserve the
// frontend's exact layout while relocation slots retain their LLVM constants.
// Using a packed aggregate is necessary because descriptors contain 32-bit
// offsets adjacent to pointer-sized fields.
func llvmDataFields(s *obj.LSym, globals map[*obj.LSym]llvm.Value, data map[*obj.LSym]bool) []llvm.Value {
	dataSize := int(s.Size)
	if dataSize < len(s.P) {
		dataSize = len(s.P)
	}
	bytes := make([]byte, dataSize)
	copy(bytes, s.P)
	relocs := llvmDataStorageRelocs(s)
	sort.Slice(relocs, func(i, j int) bool { return relocs[i].Off < relocs[j].Off })

	fields := make([]llvm.Value, 0, len(relocs)*2+1)
	pos := 0
	for _, r := range relocs {
		off := int(r.Off)
		end := off + int(r.Siz)
		if off < pos || end > dataSize {
			base.Fatalf("invalid relocation [%d,%d) in Go data symbol %s of size %d", off, end, s.Name, dataSize)
		}
		if off > pos {
			fields = append(fields, llvmDataBytes(bytes[pos:off]))
		}
		if globals == nil {
			fields = append(fields, llvmDataRelocZero(r))
		} else {
			fields = append(fields, llvmDataRelocValue(s, r, globals, data))
		}
		pos = end
	}
	if pos < dataSize {
		fields = append(fields, llvmDataBytes(bytes[pos:]))
	}
	if len(fields) == 0 {
		fields = append(fields, llvmDataBytes(nil))
	}
	return fields
}

func llvmDataBytes(b []byte) llvm.Value {
	return GlobalCtxt.ConstString(string(b), false)
}

func llvmDataRelocZero(r obj.Reloc) llvm.Value {
	switch r.Type {
	case objabi.R_ADDR, objabi.R_WEAKADDR:
		if r.Siz != uint8(base.Ctxt.Arch.PtrSize) {
			base.Fatalf("unsupported address relocation size %d", r.Siz)
		}
		return llvm.ConstPointerNull(GlobalCtxt.PointerType(0))
	case objabi.R_ADDROFF, objabi.R_WEAKADDROFF, objabi.R_METHODOFF:
		if r.Siz != 4 {
			base.Fatalf("unsupported offset relocation size %d", r.Siz)
		}
		return llvm.ConstInt(GlobalCtxt.Int32Type(), 0, false)
	default:
		base.Fatalf("unsupported Go data relocation %s", r.Type)
		return llvm.Value{}
	}
}

func llvmDataRelocValue(source *obj.LSym, r obj.Reloc, globals map[*obj.LSym]llvm.Value, data map[*obj.LSym]bool) llvm.Value {
	if r.Sym == nil {
		base.Fatalf("nil relocation target in Go data symbol %s", source.Name)
	}
	target, ok := globals[r.Sym]
	if !ok {
		target = llvmExternalDataRef(r.Sym, data)
	}
	addr := target
	if r.Add != 0 {
		if !target.IsAFunction().IsNil() {
			base.Fatalf("non-zero addend %d on function relocation %s -> %s", r.Add, source.Name, r.Sym.Name)
		}
		addr = llvm.ConstGEP(GlobalCtxt.Int8Type(), target, []llvm.Value{
			llvm.ConstInt(GlobalCtxt.Int64Type(), uint64(r.Add), true),
		})
	}

	switch r.Type {
	case objabi.R_ADDR, objabi.R_WEAKADDR:
		if r.Siz != uint8(base.Ctxt.Arch.PtrSize) {
			base.Fatalf("unsupported address relocation size %d", r.Siz)
		}
		return addr
	case objabi.R_ADDROFF, objabi.R_WEAKADDROFF, objabi.R_METHODOFF:
		if r.Siz != 4 {
			base.Fatalf("unsupported offset relocation size %d", r.Siz)
		}
		return llvm.ConstPtrToInt(addr, GlobalCtxt.Int32Type())
	default:
		base.Fatalf("unsupported Go data relocation %s in %s", r.Type, source.Name)
		return llvm.Value{}
	}
}

func llvmExternalDataRef(s *obj.LSym, data map[*obj.LSym]bool) llvm.Value {
	if data[s] {
		base.Fatalf("unlowered Go data symbol %s", s.Name)
	}
	// Runtime references materialized by reflectdata are frequently still Sxxx
	// at this point. Their ABI nevertheless identifies them as functions (for
	// example runtime.memequal64 in an equality closure), so do not rely on
	// STEXT alone here.
	if s.Type == objabi.STEXT || s.Type == objabi.STEXTFIPS || s.ABI() == obj.ABIInternal {
		if f := CurrentModule.NamedFunction(s.Name); !f.IsNil() {
			return f
		}
		f := llvm.AddFunction(CurrentModule, s.Name, llvm.FunctionType(GlobalCtxt.VoidType(), nil, false))
		f.SetFunctionCallConv(llvmCallConv(s.ABI()))
		return f
	}
	if g := CurrentModule.NamedGlobal(s.Name); !g.IsNil() {
		return g
	}
	return llvm.AddGlobal(CurrentModule, GlobalCtxt.Int8Type(), s.Name)
}

func llvmDataSection(s *obj.LSym) string {
	switch s.Type {
	case objabi.SRODATA, objabi.SRODATAFIPS, objabi.SNOPTRDATA, objabi.SNOPTRDATAFIPS:
		return ".rodata"
	case objabi.SDATA, objabi.SDATAFIPS:
		return ".data"
	case objabi.SBSS, objabi.SNOPTRBSS:
		return ".bss"
	default:
		base.Fatalf("unsupported Go data symbol kind %s for %s", s.Type, s.Name)
		return ""
	}
}

func llvmDataIsReadOnly(s *obj.LSym) bool {
	switch s.Type {
	case objabi.SRODATA, objabi.SRODATAFIPS, objabi.SNOPTRDATA, objabi.SNOPTRDATAFIPS:
		return true
	default:
		return false
	}
}

func setGoObjDataFlags(g llvm.Value, s *obj.LSym) {
	var flag, flag2 uint64
	if s.DuplicateOK() {
		flag |= 1 << 0 // goobj.SymFlagDupok
	}
	if s.Local() {
		flag |= 1 << 1 // goobj.SymFlagLocal
	}
	if s.MakeTypelink() {
		flag |= 1 << 2 // goobj.SymFlagTypelink
	}
	if strings.HasPrefix(s.Name, "type:") && len(s.Name) > len("type:") && s.Name[5] != '.' && s.Type == objabi.SRODATA {
		flag |= 1 << 6 // goobj.SymFlagGoType
	}
	if s.UsedInIface() {
		flag2 |= 1 << 0 // goobj.SymFlagUsedInIface
	}
	if strings.HasPrefix(s.Name, "go:itab.") && s.Type == objabi.SRODATA {
		flag2 |= 1 << 1 // goobj.SymFlagItab
	}
	if s.IsLinkname() {
		flag2 |= 1 << 4 // goobj.SymFlagLinkname
	}
	if s.ABIWrapper() {
		flag2 |= 1 << 5 // goobj.SymFlagABIWrapper
	}
	if flag == 0 && flag2 == 0 {
		return
	}
	g.SetGlobalMetadata(GlobalCtxt.MDKindID("goobj.symbol.flags"), GlobalCtxt.MDNode([]llvm.Metadata{
		llvm.ConstInt(GlobalCtxt.Int32Type(), flag, false).ConstantAsMetadata(),
		llvm.ConstInt(GlobalCtxt.Int32Type(), flag2, false).ConstantAsMetadata(),
	}))
}

func setGoObjRelocMetadata(g llvm.Value, s *obj.LSym) {
	entries := make([]llvm.Metadata, 0, len(s.R))
	for _, r := range s.R {
		switch r.Type {
		case objabi.R_ADDR, objabi.R_WEAKADDR, objabi.R_ADDROFF, objabi.R_WEAKADDROFF, objabi.R_METHODOFF:
			entries = append(entries, GlobalCtxt.MDNode([]llvm.Metadata{
				llvm.ConstInt(GlobalCtxt.Int32Type(), uint64(r.Off), false).ConstantAsMetadata(),
				llvm.ConstInt(GlobalCtxt.Int32Type(), uint64(uint16(r.Type)), false).ConstantAsMetadata(),
			}))
		case objabi.R_KEEP:
			continue
		default:
			base.Fatalf("unsupported Go data relocation %s in %s", r.Type, s.Name)
		}
	}
	if len(entries) != 0 {
		g.SetGlobalMetadata(GlobalCtxt.MDKindID("goobj.relocs"), GlobalCtxt.MDNode(entries))
	}
}

// R_KEEP is a zero-width linker reachability edge, not a storage relocation.
// Keep it separate from !goobj.relocs so llc can synthesize the GoObj record
// without inventing bytes or a target-address expression in the global.
func setGoObjKeepMetadata(g llvm.Value, s *obj.LSym) {
	entries := make([]llvm.Metadata, 0)
	for _, r := range s.R {
		if r.Type != objabi.R_KEEP {
			continue
		}
		if r.Sym == nil {
			base.Fatalf("nil R_KEEP target in %s", s.Name)
		}
		entries = append(entries, GlobalCtxt.MDString(r.Sym.Name))
	}
	if len(entries) != 0 {
		g.SetGlobalMetadata(GlobalCtxt.MDKindID("goobj.keep"), GlobalCtxt.MDNode(entries))
	}
}

func llvmDataStorageRelocs(s *obj.LSym) []obj.Reloc {
	relocs := make([]obj.Reloc, 0, len(s.R))
	for _, r := range s.R {
		if r.Type != objabi.R_KEEP {
			relocs = append(relocs, r)
		}
	}
	return relocs
}
