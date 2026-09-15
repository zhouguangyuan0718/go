// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
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
			// Linker identity is independent of the logical name used by models.
			name := "runtime.Goexit"
			reference := name + goobj.LinknameSymbolSuffix
			fn := llvmFunctions.getOrInsert(name, reference, provisional, goABIInternalCallConv)
			builder := GlobalCtxt.NewBuilder()
			defer builder.Dispose()
			caller := llvm.AddFunction(module, "caller", provisional.Type)
			builder.SetInsertPointAtEnd(llvm.AddBasicBlock(caller, "entry"))
			call := builder.CreateCall(provisional.Type, fn, nil, "")
			call.SetInstructionCallConv(goABIInternalCallConv)
			builder.CreateUnreachable()

			final := provisional
			final.Type = llvm.FunctionType(void, []llvm.Type{GlobalCtxt.PointerType(0)}, false)
			fn = llvmFunctions.getOrInsert(name, reference, final, goABIInternalCallConv)
			if fn.GetEnumFunctionAttribute(llvm.AttributeKindID("noreturn")).C == nil {
				t.Fatal("replacement declaration lost its noreturn model")
			}
			if got := module.NamedFunction(reference); got != fn {
				t.Fatal("manager did not return the current module's replacement declaration")
			}
			if got := llvmFunctions.getOrInsert(name, reference, final, goABIInternalCallConv); got != fn {
				t.Fatal("repeated lookup created a different declaration")
			}
			if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
				t.Fatalf("invalid replacement IR: %v\n%s", err, module.String())
			}
		}()
	}
}

func TestLLVMFunctionModelsKeepGCLeafOnCalls(t *testing.T) {
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
		for _, name := range []string{"runtime.memmove", "runtime.memequal", "runtime.wbMove", "runtime.wbZero", "runtime.mallocgc", "runtime.memmoveLike"} {
			reference := name
			if encoded, ok := goobj.BuiltinSymbolName(name, int(obj.ABIInternal)); ok {
				reference = encoded
			}
			fn := llvmFunctions.getOrInsert(name, reference, sig, cc)
			call := builder.CreateCall(sig.Type, fn, nil, "")
			call.SetInstructionCallConv(cc)
			llvmFunctions.configureCall(name, call)
			wantLeaf := cc == goABIInternalCallConv && name != "runtime.mallocgc" && name != "runtime.memmoveLike"
			if got := call.GetCallSiteStringAttribute(llvmAttributeFunctionIndex, goGCLeafFunctionAttr).C != nil; got != wantLeaf {
				t.Errorf("%s: GC-leaf call = %v, want %v", name, got, wantLeaf)
			}
			if fn.GetStringAttributeAtIndex(llvmAttributeFunctionIndex, goGCLeafFunctionAttr).C != nil {
				t.Errorf("%s: call-only GC-leaf contract leaked to the declaration", name)
			}
		}
	}
	builder.CreateRetVoid()
	if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
		t.Fatalf("invalid modeled call IR: %v\n%s", err, module.String())
	}
}
