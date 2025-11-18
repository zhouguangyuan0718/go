//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"

	"github.com/goallc/go-llvm"
)

type LLVMFuncContext struct {
	BBs    map[ID]llvm.BasicBlock
	Vs     map[ID]llvm.Value
	F      *Func
	LF     llvm.Value
	b      llvm.Builder
	ArgIdx int
}

func (lfc *LLVMFuncContext) CompileBlock(BB *Block) {
	lfc.b.SetInsertPointAtEnd(lfc.BBs[BB.ID])

	var lVal llvm.Value
	for _, v := range BB.Values {
		switch v.Op {
		case OpInitMem:
		case OpSP:
		case OpSB:
		case OpLocalAddr:
			lVal = lfc.b.CreateAlloca(getLLVMType(v.Type.Elem()), v.String())
		case OpArg:
			lVal = lfc.LF.Param(lfc.ArgIdx)
			// TODO Set arg name
			lfc.ArgIdx++
		case OpConst64:
			lVal = llvm.ConstInt(getLLVMType(v.Type), uint64(auxIntToInt64(v.AuxInt)), v.Type.IsSigned())
		case OpAdd64:
			lVal = lfc.b.CreateAdd(lfc.Vs[v.Args[0].ID], lfc.Vs[v.Args[1].ID], v.String())
		case OpLess64:
			lVal = lfc.b.CreateICmp(llvm.IntSLT, lfc.Vs[v.Args[0].ID], lfc.Vs[v.Args[1].ID], v.String())
		case OpCopy:
			lVal = lfc.Vs[v.Args[0].ID] // TODO ?
		case OpMakeResult:
			if len(v.Args) == 2 {
				lVal = lfc.Vs[v.Args[0].ID]
			} else {
				// TODO multiple return
			}
		case OpPhi:
			lVal = lfc.b.CreatePHI(getLLVMType(v.Type), v.String())
			var incomingLVals []llvm.Value
			for _, incoming := range v.Args {
				incomingLVals = append(incomingLVals, lfc.Vs[incoming.ID])
			}
			var predecessors []llvm.BasicBlock
			for _, pred := range BB.Preds {
				predecessors = append(predecessors, lfc.BBs[pred.Block().ID])
			}
			lVal.AddIncoming(incomingLVals, predecessors)
		}
		lfc.Vs[v.ID] = lVal
	}
	switch BB.Kind {
	case BlockRet:
		lfc.b.CreateRet(lfc.Vs[BB.Controls[0].ID])
	case BlockIf:
		lfc.b.CreateCondBr(lfc.Vs[BB.Controls[0].ID], lfc.BBs[BB.Succs[0].Block().ID], lfc.BBs[BB.Succs[1].Block().ID])
	case BlockPlain:
		lfc.b.CreateBr(lfc.BBs[BB.Succs[0].Block().ID])
	default:
	}
}

func LLVMCompile(f *Func) {
	FCtxt := &LLVMFuncContext{
		BBs: map[ID]llvm.BasicBlock{},
		Vs:  map[ID]llvm.Value{},
		F:   f,
		b:   GlobalCtxt.NewBuilder(),
	}

	FCtxt.LF = llvm.AddFunction(CurrentModule, base.Ctxt.Pkgpath+"."+f.Name, getLLVMType(f.Type))
	for _, BB := range f.Blocks {
		FCtxt.BBs[BB.ID] = GlobalCtxt.AddBasicBlock(FCtxt.LF, BB.String())
	}
	for _, BB := range f.Blocks {
		FCtxt.CompileBlock(BB)
	}
}

var CurrentModule llvm.Module
var type2lTypes = map[*types.Type]llvm.Type{}

var GlobalCtxt = llvm.GlobalContext()

func getLLVMType(typ *types.Type) llvm.Type {
	if t, ok := type2lTypes[typ]; ok {
		return t
	}
	switch typ.Kind() {
	case types.TFUNC:
		var llvmRetType llvm.Type
		var llvmParamTypes []llvm.Type
		if typ.NumResults() == 1 {
			llvmRetType = getLLVMType(typ.Result(0).Type)
			numParam := typ.NumParams()
			for i := 0; i < numParam; i++ {
				llvmParamTypes = append(llvmParamTypes, getLLVMType(typ.Param(i).Type))
			}
		}
		return llvm.FunctionType(llvmRetType, llvmParamTypes, false)
	case types.TPTR:
		return llvm.PointerType(getLLVMType(typ.Elem()), 0)
	default:
		panic("unhandled default case")
	}
}

func InitModule(pkg *types.Pkg) {
	type2lTypes[types.Types[types.TINT8]] = GlobalCtxt.Int8Type()
	type2lTypes[types.Types[types.TUINT8]] = GlobalCtxt.Int8Type()
	type2lTypes[types.Types[types.TINT16]] = GlobalCtxt.Int16Type()
	type2lTypes[types.Types[types.TUINT16]] = GlobalCtxt.Int16Type()
	type2lTypes[types.Types[types.TUINT32]] = GlobalCtxt.Int32Type()
	type2lTypes[types.Types[types.TUINT32]] = GlobalCtxt.Int32Type()
	type2lTypes[types.Types[types.TINT64]] = GlobalCtxt.Int64Type()
	type2lTypes[types.Types[types.TUINT8]] = GlobalCtxt.Int64Type()
	if types.PtrSize == 8 {
		type2lTypes[types.Types[types.TINT]] = GlobalCtxt.Int64Type()
		type2lTypes[types.Types[types.TUINT]] = GlobalCtxt.Int64Type()
	} else {
		type2lTypes[types.Types[types.TINT]] = GlobalCtxt.Int32Type()
		type2lTypes[types.Types[types.TUINT]] = GlobalCtxt.Int32Type()
	}
	type2lTypes[types.Types[types.TUINTPTR]] = GlobalCtxt.Int8Type()
	type2lTypes[types.Types[types.TFLOAT32]] = GlobalCtxt.FloatType()
	type2lTypes[types.Types[types.TFLOAT64]] = GlobalCtxt.DoubleType()
	//type2lTypes[types.Types[types.TCOMPLEX64]] =
	//type2lTypes[types.Types[types.TCOMPLEX128]] =
	type2lTypes[types.Types[types.TBOOL]] = GlobalCtxt.Int1Type()

	CurrentModule = GlobalCtxt.NewModule(pkg.Path)
}

func Output(fileName string) error {
	return llvm.LLVMPrintModuleToFile(CurrentModule, fileName)
}
