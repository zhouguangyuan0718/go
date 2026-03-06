//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"fmt"

	"github.com/goallc/go-llvm"
)

type LLVMFuncContext struct {
	BBs        map[ID]llvm.BasicBlock
	Vs         map[ID]llvm.Value
	F          *Func
	LF         llvm.Value
	b          llvm.Builder
	ReturnType llvm.Type
	ArgIdx     int
}

func (lfc *LLVMFuncContext) FinishPhi() {
	for _, BB := range lfc.F.Blocks {
		for _, v := range BB.Values {
			if v.Op != OpPhi {
				continue
			}
			var incomingLVals []llvm.Value
			for _, incoming := range v.Args {
				incomingLVals = append(incomingLVals, lfc.Vs[incoming.ID])
			}
			var predecessors []llvm.BasicBlock
			for _, pred := range BB.Preds {
				predecessors = append(predecessors, lfc.BBs[pred.Block().ID])
			}
			lfc.Vs[v.ID].AddIncoming(incomingLVals, predecessors)
		}
	}
}

func (lfc *LLVMFuncContext) GenLV(v *Value) llvm.Value {
	if lv, ok := lfc.Vs[v.ID]; ok {
		return lv
	}
	var lVal llvm.Value
	arg0 := func() llvm.Value { return lfc.GenLV(v.Args[0]) }
	arg1 := func() llvm.Value { return lfc.GenLV(v.Args[1]) }
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
	case OpConst8, OpConst16, OpConst32, OpConst64:
		lVal = llvm.ConstInt(getLLVMType(v.Type), uint64(auxIntToInt64(v.AuxInt)), v.Type.IsSigned())
	case OpConstBool:
		if auxIntToBool(v.AuxInt) {
			lVal = llvm.ConstInt(getLLVMType(v.Type), 1, false)
		} else {
			lVal = llvm.ConstInt(getLLVMType(v.Type), 0, false)
		}
	case OpConst32F:
		lVal = llvm.ConstFloat(getLLVMType(v.Type), float64(auxIntToFloat32(v.AuxInt)))
	case OpConst64F:
		lVal = llvm.ConstFloat(getLLVMType(v.Type), auxIntToFloat64(v.AuxInt))
	case OpConstNil:
		lVal = llvm.ConstNull(getLLVMType(v.Type))
	case OpConstString:
		str := auxToString(v.Aux)
		strData := llvm.ConstString(auxToString(v.Aux), false)
		strVal := llvm.AddGlobal(CurrentModule, strData.Type(), str)
		strVal.SetInitializer(strData)
		strVal.SetUnnamedAddr(true)
		strVal.SetLinkage(llvm.LinkOnceAnyLinkage)
		strVal.SetGlobalConstant(true)
		strLen := llvm.ConstInt(getLLVMType(types.Types[types.TINT]), uint64(len(auxToString(v.Aux))), true)
		lVal = llvm.ConstNamedStruct(getLLVMType(v.Type), []llvm.Value{strVal, strLen})
	case OpAdd64, OpAdd32, OpAdd16, OpAdd8:
		lVal = lfc.b.CreateAdd(arg0(), arg1(), v.String())
	case OpAdd32F, OpAdd64F:
		lVal = lfc.b.CreateFAdd(arg0(), arg1(), v.String())
	case OpSub64, OpSub32, OpSub16, OpSub8:
		lVal = lfc.b.CreateSub(arg0(), arg1(), v.String())
	case OpSub32F, OpSub64F:
		lVal = lfc.b.CreateFSub(arg0(), arg1(), v.String())
	case OpMul64, OpMul32, OpMul16, OpMul8:
		lVal = lfc.b.CreateMul(arg0(), arg1(), v.String())
	case OpMul32F, OpMul64F:
		lVal = lfc.b.CreateFMul(arg0(), arg1(), v.String())
	case OpDiv64, OpDiv32, OpDiv16, OpDiv8:
		lVal = lfc.b.CreateSDiv(arg0(), arg1(), v.String())
	case OpDiv64u, OpDiv32u, OpDiv16u, OpDiv8u:
		lVal = lfc.b.CreateUDiv(arg0(), arg1(), v.String())
	case OpDiv32F, OpDiv64F:
		lVal = lfc.b.CreateFDiv(arg0(), arg1(), v.String())
	case OpMod64, OpMod32, OpMod16, OpMod8:
		lVal = lfc.b.CreateSRem(arg0(), arg1(), v.String())
	case OpMod64u, OpMod32u, OpMod16u, OpMod8u:
		lVal = lfc.b.CreateURem(arg0(), arg1(), v.String())
	case OpAnd64, OpAnd32, OpAnd16, OpAnd8, OpAndB:
		lVal = lfc.b.CreateAnd(arg0(), arg1(), v.String())
	case OpOr64, OpOr32, OpOr16, OpOr8, OpOrB:
		lVal = lfc.b.CreateOr(arg0(), arg1(), v.String())
	case OpXor64, OpXor32, OpXor16, OpXor8, OpXorB:
		lVal = lfc.b.CreateXor(arg0(), arg1(), v.String())
	case OpCom64, OpCom32, OpCom16, OpCom8:
		lVal = lfc.b.CreateNot(arg0(), v.String())
	case OpNeg64, OpNeg32, OpNeg16, OpNeg8:
		lVal = lfc.b.CreateNeg(arg0(), v.String())
	case OpNeg32F, OpNeg64F:
		lVal = lfc.b.CreateFNeg(arg0(), v.String())
	case OpNot:
		lVal = lfc.b.CreateNot(arg0(), v.String())
	case OpEq64, OpEq32, OpEq16, OpEq8, OpEqB, OpEqPtr:
		lVal = lfc.b.CreateICmp(llvm.IntEQ, arg0(), arg1(), v.String())
	case OpEq32F, OpEq64F:
		lVal = lfc.b.CreateFCmp(llvm.FloatOEQ, arg0(), arg1(), v.String())
	case OpNeq64, OpNeq32, OpNeq16, OpNeq8, OpNeqB, OpNeqPtr:
		lVal = lfc.b.CreateICmp(llvm.IntNE, arg0(), arg1(), v.String())
	case OpNeq32F, OpNeq64F:
		lVal = lfc.b.CreateFCmp(llvm.FloatONE, arg0(), arg1(), v.String())
	case OpLess64:
		lVal = lfc.b.CreateICmp(llvm.IntSLT, arg0(), arg1(), v.String())
	case OpLess64U, OpLess32U, OpLess16U, OpLess8U:
		lVal = lfc.b.CreateICmp(llvm.IntULT, arg0(), arg1(), v.String())
	case OpLess32, OpLess16, OpLess8:
		lVal = lfc.b.CreateICmp(llvm.IntSLT, arg0(), arg1(), v.String())
	case OpLess32F, OpLess64F:
		lVal = lfc.b.CreateFCmp(llvm.FloatOLT, arg0(), arg1(), v.String())
	case OpLeq64, OpLeq32, OpLeq16, OpLeq8:
		lVal = lfc.b.CreateICmp(llvm.IntSLE, arg0(), arg1(), v.String())
	case OpLeq64U, OpLeq32U, OpLeq16U, OpLeq8U:
		lVal = lfc.b.CreateICmp(llvm.IntULE, arg0(), arg1(), v.String())
	case OpLeq32F, OpLeq64F:
		lVal = lfc.b.CreateFCmp(llvm.FloatOLE, arg0(), arg1(), v.String())
	case OpLsh64x64, OpLsh64x32, OpLsh64x16, OpLsh64x8,
		OpLsh32x64, OpLsh32x32, OpLsh32x16, OpLsh32x8,
		OpLsh16x64, OpLsh16x32, OpLsh16x16, OpLsh16x8,
		OpLsh8x64, OpLsh8x32, OpLsh8x16, OpLsh8x8:
		lVal = lfc.b.CreateShl(arg0(), arg1(), v.String())
	case OpRsh64x64, OpRsh64x32, OpRsh64x16, OpRsh64x8,
		OpRsh32x64, OpRsh32x32, OpRsh32x16, OpRsh32x8,
		OpRsh16x64, OpRsh16x32, OpRsh16x16, OpRsh16x8,
		OpRsh8x64, OpRsh8x32, OpRsh8x16, OpRsh8x8:
		lVal = lfc.b.CreateAShr(arg0(), arg1(), v.String())
	case OpRsh64Ux64, OpRsh64Ux32, OpRsh64Ux16, OpRsh64Ux8,
		OpRsh32Ux64, OpRsh32Ux32, OpRsh32Ux16, OpRsh32Ux8,
		OpRsh16Ux64, OpRsh16Ux32, OpRsh16Ux16, OpRsh16Ux8,
		OpRsh8Ux64, OpRsh8Ux32, OpRsh8Ux16, OpRsh8Ux8:
		lVal = lfc.b.CreateLShr(arg0(), arg1(), v.String())
	case OpSignExt8to16, OpSignExt8to32, OpSignExt8to64,
		OpSignExt16to32, OpSignExt16to64, OpSignExt32to64:
		lVal = lfc.b.CreateSExt(arg0(), getLLVMType(v.Type), v.String())
	case OpZeroExt8to16, OpZeroExt8to32, OpZeroExt8to64,
		OpZeroExt16to32, OpZeroExt16to64, OpZeroExt32to64:
		lVal = lfc.b.CreateZExt(arg0(), getLLVMType(v.Type), v.String())
	case OpTrunc64to32, OpTrunc64to16, OpTrunc64to8,
		OpTrunc32to16, OpTrunc32to8, OpTrunc16to8:
		lVal = lfc.b.CreateTrunc(arg0(), getLLVMType(v.Type), v.String())
	case OpCvt32to32F, OpCvt32to64F, OpCvt64to32F, OpCvt64to64F:
		lVal = lfc.b.CreateSIToFP(arg0(), getLLVMType(v.Type), v.String())
	case OpCvt32Uto32F, OpCvt32Uto64F, OpCvt64Uto32F, OpCvt64Uto64F:
		lVal = lfc.b.CreateUIToFP(arg0(), getLLVMType(v.Type), v.String())
	case OpCvt32Fto32, OpCvt32Fto64, OpCvt64Fto32, OpCvt64Fto64:
		lVal = lfc.b.CreateFPToSI(arg0(), getLLVMType(v.Type), v.String())
	case OpCvt32Fto32U, OpCvt32Fto64U, OpCvt64Fto32U, OpCvt64Fto64U:
		lVal = lfc.b.CreateFPToUI(arg0(), getLLVMType(v.Type), v.String())
	case OpCvt32Fto64F, OpCvt64Fto32F:
		lVal = lfc.b.CreateFPCast(arg0(), getLLVMType(v.Type), v.String())
	case OpCopy, OpConvert:
		lVal = arg0()
	case OpBitLen32, OpBitLen64:
		lVal = arg0() // TODO implement intrinsic lowering
	case OpOffPtr:
		off := llvm.ConstInt(getLLVMType(types.Types[types.TINT]), uint64(auxIntToInt64(v.AuxInt)), true)
		ptr := lfc.b.CreateBitCast(arg0(), llvm.PointerType(GlobalCtxt.Int8Type(), 0), v.String()+".i8")
		lVal = lfc.b.CreateBitCast(lfc.b.CreateGEP(GlobalCtxt.Int8Type(), ptr, []llvm.Value{off}, v.String()), getLLVMType(v.Type), v.String())
	case OpPtrIndex:
		lVal = lfc.b.CreateGEP(getLLVMType(v.Type.Elem()), arg0(), []llvm.Value{arg1()}, v.String())
	case OpAddr:
		lVal = arg0()
	case OpMakeResult:
		switch len(v.Args) {
		case 1:
		case 2:
			lVal = lfc.GenLV(v.Args[0])
		default:
			lVal = llvm.Undef(getLLVMType(lfc.F.Type).ReturnType())
			numRes := len(v.Args) - 1
			for i := 0; i < numRes; i++ {
				lVal = lfc.b.CreateInsertValue(lVal, lfc.GenLV(v.Args[i]), i, "")
			}
			lVal.SetName(v.String())
		}
	case OpPhi:
		lVal = lfc.b.CreatePHI(getLLVMType(v.Type), v.String())
	case OpLoad:
		lVal = lfc.b.CreateLoad(getLLVMType(v.Type), arg0(), v.String())
	case OpNilCheck:
		lVal = lfc.Vs[v.Args[0].ID] // TODO nil check
	case OpStore:
		lVal = lfc.b.CreateStore(arg1(), arg0())
	case OpStructSelect:
		lVal = lfc.b.CreateExtractValue(lfc.GenLV(v.Args[0]), int(auxIntToInt32(v.AuxInt)), v.String())
	case OpStructMake:
		lVal = llvm.Undef(getLLVMType(v.Type))
		numFields := v.Type.NumFields()
		for i := 0; i < numFields; i++ {
			lVal = lfc.b.CreateInsertValue(lVal, lfc.GenLV(v.Args[i]), i, "")
		}
		lVal.SetName(v.String())
	case OpStringLen, OpSliceLen:
		lVal = lfc.b.CreateExtractValue(lfc.GenLV(v.Args[0]), 1, v.String())
	case OpSliceCap:
		lVal = lfc.b.CreateExtractValue(lfc.GenLV(v.Args[0]), 2, v.String())
	default:
		fmt.Println("skip value: ", v)
	}
	lfc.Vs[v.ID] = lVal
	return lVal
}

func (lfc *LLVMFuncContext) CompileBlock(BB *Block) {
	lfc.b.SetInsertPointAtEnd(lfc.BBs[BB.ID])
	//fmt.Println("compiling: ", BB.String())
	//for _, v := range BB.Values {
	//	fmt.Println("value: ", v)
	//}
	//fmt.Println()
	for _, v := range BB.Values {
		lfc.GenLV(v)
	}
	switch BB.Kind {
	case BlockRet:
		if BB.Controls[0].Type.NumFields() == 1 { // TODO type like int,mem
			lfc.b.CreateRetVoid()
		} else {
			lfc.b.CreateRet(lfc.Vs[BB.Controls[0].ID])
		}
	case BlockIf:
		lfc.b.CreateCondBr(lfc.Vs[BB.Controls[0].ID], lfc.BBs[BB.Succs[0].Block().ID], lfc.BBs[BB.Succs[1].Block().ID])
	case BlockPlain:
		lfc.b.CreateBr(lfc.BBs[BB.Succs[0].Block().ID])
	default:

	}
}

func (lfc *LLVMFuncContext) MappingName() {
	for s, vs := range lfc.F.NamedValues {
		for _, v := range vs {
			if v.Op != OpArg {
				continue
			}
			if lv, ok := lfc.Vs[v.ID]; ok {
				llvm.LLVMSetValueName2(lv, s.N.Sym().Name)
			}
		}
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
	FCtxt.FinishPhi()
	FCtxt.MappingName()

	err := llvm.VerifyFunction(FCtxt.LF, llvm.PrintMessageAction)
	if err != nil {
		//TODO handle error
	}
}

var CurrentModule llvm.Module
var type2lTypes = map[*types.Type]llvm.Type{}

var GlobalCtxt = llvm.GlobalContext()

func getLLVMType(typ *types.Type) llvm.Type {
	if t, ok := type2lTypes[typ]; ok {
		return t
	}
	var lType llvm.Type
	switch typ.Kind() {
	case types.TFUNC:
		var llvmRetType llvm.Type
		var llvmParamTypes []llvm.Type
		switch typ.NumResults() {
		case 0:
			llvmRetType = GlobalCtxt.VoidType()
		case 1:
			llvmRetType = getLLVMType(typ.Result(0).Type)
		default:
			var llvmResTypes []llvm.Type
			for i := 0; i < typ.NumResults(); i++ {
				llvmResTypes = append(llvmResTypes, getLLVMType(typ.Result(i).Type))
			}
			llvmRetType = llvm.StructType(llvmResTypes, false)
			// TODO need add attribute to distinguish with real struct type
		}
		numParam := typ.NumParams()
		for i := 0; i < numParam; i++ {
			llvmParamTypes = append(llvmParamTypes, getLLVMType(typ.Param(i).Type))
		}
		lType = llvm.FunctionType(llvmRetType, llvmParamTypes, false)
	case types.TPTR:
		lType = llvm.PointerType(getLLVMType(typ.Elem()), 0)
	case types.TSTRUCT:
		var llvmFieldTypes []llvm.Type
		numField := typ.NumFields()
		for i := 0; i < numField; i++ {
			llvmFieldTypes = append(llvmFieldTypes, getLLVMType(typ.FieldType(i)))
		}
		if typ.Sym() != nil {
			lType = GlobalCtxt.StructCreateNamed(typ.NameString())
			lType.StructSetBody(llvmFieldTypes, false)
		} else {
			lType = llvm.StructType(llvmFieldTypes, false)
		}
	case types.TSTRING:
		stringFieldTypes := []llvm.Type{GlobalCtxt.PointerType(0), getLLVMType(types.Types[types.TINT])}
		lType = GlobalCtxt.StructCreateNamed("string")
		lType.StructSetBody(stringFieldTypes, false)
	case types.TSLICE:
		sliceFieldTypes := []llvm.Type{GlobalCtxt.PointerType(0),
			getLLVMType(types.Types[types.TINT]),
			getLLVMType(types.Types[types.TINT])}
		lType = GlobalCtxt.StructCreateNamed("slice")
		lType.StructSetBody(sliceFieldTypes, false)
	default:
		fmt.Println("unknown type", typ, ", kind", typ.Kind())
		panic("unhandled default case")
	}
	type2lTypes[typ] = lType
	return lType
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
		type2lTypes[types.Types[types.TUINTPTR]] = GlobalCtxt.Int64Type()
	} else {
		type2lTypes[types.Types[types.TINT]] = GlobalCtxt.Int32Type()
		type2lTypes[types.Types[types.TUINT]] = GlobalCtxt.Int32Type()
		type2lTypes[types.Types[types.TUINTPTR]] = GlobalCtxt.Int32Type()
	}

	type2lTypes[types.Types[types.TUNSAFEPTR]] = GlobalCtxt.PointerType(0)
	type2lTypes[types.Types[types.TFLOAT32]] = GlobalCtxt.FloatType()
	type2lTypes[types.Types[types.TFLOAT64]] = GlobalCtxt.DoubleType()
	//type2lTypes[types.Types[types.TCOMPLEX64]] =
	//type2lTypes[types.Types[types.TCOMPLEX128]] =
	type2lTypes[types.Types[types.TBOOL]] = GlobalCtxt.Int1Type()

	type2lTypes[types.ByteType] = GlobalCtxt.Int8Type()
	type2lTypes[types.RuneType] = GlobalCtxt.Int32Type()

	CurrentModule = GlobalCtxt.NewModule(pkg.Path)
}

func Output(fileName string) error {
	return llvm.LLVMPrintModuleToFile(CurrentModule, fileName)
}
