// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/types"
	"cmd/internal/goobj"
	"cmd/internal/obj"
	"testing"

	"github.com/goallc/go-llvm"
)

func TestLLVMFunctionModelSurvivesDeclarationReplacement(t *testing.T) {
	oldModule := CurrentModule
	defer func() { CurrentModule = oldModule }()

	// Reuse the manager across modules, as repeated compilation and LLVM tests
	// do. It must never return a value belonging to a previous module.
	for range 2 {
		func() {
			module := GlobalCtxt.NewModule("function_models")
			defer module.Dispose()
			CurrentModule = module
			void := GlobalCtxt.VoidType()
			provisional := llvmFuncSignature{
				Type: llvm.FunctionType(void, nil, false), ReturnType: void,
				ClosureContextIndex: -1,
			}
			// The model is looked up using the actual encoded IR name.
			name := "runtime.Goexit"
			reference := name + goobj.LinknameSymbolSuffix
			fn := getOrInsertLLVMFunction(reference, provisional, goABIInternalCallConv)
			builder := GlobalCtxt.NewBuilder()
			defer builder.Dispose()
			caller := llvm.AddFunction(module, "caller", provisional.Type)
			builder.SetInsertPointAtEnd(llvm.AddBasicBlock(caller, "entry"))
			call := builder.CreateCall(provisional.Type, fn, nil, "")
			call.SetInstructionCallConv(goABIInternalCallConv)
			builder.CreateUnreachable()

			final := provisional
			final.Type = llvm.FunctionType(void, []llvm.Type{GlobalCtxt.PointerType(0)}, false)
			fn = getOrInsertLLVMFunction(reference, final, goABIInternalCallConv)
			if fn.GetEnumFunctionAttribute(llvm.AttributeKindID("noreturn")).C == nil {
				t.Fatal("replacement declaration lost its noreturn model")
			}
			if got := module.NamedFunction(reference); got != fn {
				t.Fatal("manager did not return the current module's replacement declaration")
			}
			if got := getOrInsertLLVMFunction(reference, final, goABIInternalCallConv); got != fn {
				t.Fatal("repeated lookup created a different declaration")
			}
			if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
				t.Fatalf("invalid replacement IR: %v\n%s", err, module.String())
			}
		}()
	}
}

func TestLLVMFunctionModelsBindGCLeafToFunctions(t *testing.T) {
	oldModule := CurrentModule
	module := GlobalCtxt.NewModule("function_model_leaf_calls")
	CurrentModule = module
	defer func() { CurrentModule = oldModule; module.Dispose() }()
	builder := GlobalCtxt.NewBuilder()
	defer builder.Dispose()
	void := GlobalCtxt.VoidType()
	sig := llvmFuncSignature{
		Type: llvm.FunctionType(void, nil, false), ReturnType: void,
		ClosureContextIndex: -1,
	}
	caller := llvm.AddFunction(module, "caller", sig.Type)
	builder.SetInsertPointAtEnd(llvm.AddBasicBlock(caller, "entry"))
	for _, cc := range []llvm.CallConv{goABIInternalCallConv, goABI0CallConv} {
		for _, test := range []struct {
			name string
			leaf bool
		}{
			{"runtime.memmove", true},
			{"runtime.memequal", true},
			{"runtime.wbMove", true},
			{"runtime.wbZero", true},
			{"runtime.mallocgc", false},
			{"runtime.memmoveLike", false},
			// Neither an arbitrary suffix nor a wrong builtin index identifies
			// a registered IR function, even with a recognized name prefix.
			{"runtime.memmove<unknown>", false},
			{"runtime.memmove<builtin.999999>", false},
		} {
			references := []string{test.name, test.name + goobj.LinknameSymbolSuffix}
			abi := obj.ABIInternal
			if cc == goABI0CallConv {
				abi = obj.ABI0
			}
			if encoded, ok := goobj.BuiltinSymbolName(test.name, int(abi)); ok {
				references = append(references, encoded)
			}
			callSig := sig
			var args []llvm.Value
			if test.name == "runtime.memmove" || test.name == "runtime.memequal" {
				ptr := GlobalCtxt.PointerType(0)
				size := GlobalCtxt.IntType(int(types.PtrSize * 8))
				result := void
				if test.name == "runtime.memequal" {
					result = GlobalCtxt.Int8Type()
				}
				callSig.Type = llvm.FunctionType(result, []llvm.Type{ptr, ptr, size}, false)
				args = []llvm.Value{llvm.ConstNull(ptr), llvm.ConstNull(ptr), llvm.ConstInt(size, 0, false)}
			}
			if test.name == "runtime.mallocgc" && cc == goABIInternalCallConv {
				ptr := GlobalCtxt.PointerType(0)
				size := GlobalCtxt.IntType(int(types.PtrSize * 8))
				i8 := GlobalCtxt.Int8Type()
				callSig.Type = llvm.FunctionType(ptr, []llvm.Type{size, ptr, i8}, false)
				args = []llvm.Value{llvm.ConstInt(size, 0, false), llvm.ConstNull(ptr), llvm.ConstInt(i8, 1, false)}
			}
			for _, reference := range references {
				fn := getOrInsertLLVMFunction(reference, callSig, cc)
				call := builder.CreateCall(callSig.Type, fn, args, "")
				call.SetInstructionCallConv(cc)
				wantLeaf := test.leaf
				if got := fn.GetStringAttributeAtIndex(llvmAttributeFunctionIndex, goGCLeafFunctionAttr).C != nil; got != wantLeaf {
					t.Errorf("%s: GC-leaf function = %v, want %v", fn.Name(), got, wantLeaf)
				}
				if call.GetCallSiteStringAttribute(llvmAttributeFunctionIndex, goGCLeafFunctionAttr).C != nil {
					t.Errorf("%s: function contract was duplicated on the call", fn.Name())
				}
			}
		}
	}
	builder.CreateRetVoid()
	if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
		t.Fatalf("invalid modeled call IR: %v\n%s", err, module.String())
	}
}
