// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/types"
	"strings"
	"testing"

	"github.com/goallc/go-llvm"
)

func TestLLVMFunctionMemoryModelAttributes(t *testing.T) {
	oldModule := CurrentModule
	module := GlobalCtxt.NewModule("memory_model_signatures")
	CurrentModule = module
	defer func() { CurrentModule = oldModule; module.Dispose() }()
	ptr := GlobalCtxt.PointerType(0)
	size := GlobalCtxt.IntType(int(types.PtrSize * 8))
	str := llvm.StructType([]llvm.Type{ptr, size}, false)
	for _, test := range []struct {
		name   string
		result llvm.Type
		params []llvm.Type
		access []string
	}{
		{"runtime.memmove", GlobalCtxt.VoidType(), []llvm.Type{ptr, ptr, size}, []string{"writeonly", "readonly"}},
		{"runtime.memequal", GlobalCtxt.Int8Type(), []llvm.Type{ptr, ptr, size}, []string{"readonly", "readonly"}},
		{"runtime.memclrNoHeapPointers", GlobalCtxt.VoidType(), []llvm.Type{ptr, size}, []string{"writeonly"}},
		{"runtime.memequal_varlen", GlobalCtxt.Int8Type(), []llvm.Type{ptr, ptr}, []string{"readonly", "readonly"}},
		{"runtime.cmpstring", size, []llvm.Type{str, str}, nil},
	} {
		for _, cc := range []llvm.CallConv{goABIInternalCallConv, goABI0CallConv} {
			sig := llvmFuncSignature{
				Type: llvm.FunctionType(test.result, test.params, false), ClosureContextIndex: -1,
			}
			// Model the actual ABI0 stack carriers, including the result home.
			if cc == goABI0CallConv {
				params := make([]llvm.Type, len(test.params))
				for i, typ := range test.params {
					params[i] = ptr
					sig.Params = append(sig.Params, llvmParamSignature{ValueType: typ, Alignment: int(types.PtrSize), ByVal: true})
				}
				if test.result.TypeKind() != llvm.VoidTypeKind {
					sig.Results = []llvmResultSignature{{ValueType: test.result, Alignment: 1, InMemory: true, ParamIndex: len(params)}}
					params = append(params, ptr)
				}
				sig.Type = llvm.FunctionType(GlobalCtxt.VoidType(), params, false)
			}
			if test.name == "runtime.memequal_varlen" {
				sig.ReturnType = sig.Type.ReturnType()
				sig = sig.withClosureContext()
			}
			fn := getOrInsertLLVMFunction(test.name, sig, cc)
			want := cc == goABIInternalCallConv
			if fn.GetStringAttributeAtIndex(llvmAttributeFunctionIndex, goGCLeafFunctionAttr).C == nil {
				t.Errorf("%s: missing GC-leaf contract", fn.Name())
			}
			for _, attr := range []string{"nofree", "nocallback", "nounwind"} {
				if got := fn.GetEnumFunctionAttribute(llvm.AttributeKindID(attr)).C != nil; !got {
					t.Errorf("%s: missing %s", fn.Name(), attr)
				}
			}
			for i, access := range test.access {
				for _, attr := range []string{"captures", access} {
					if got := fn.GetEnumAttributeAtIndex(i+1, llvm.AttributeKindID(attr)).C != nil; got != want {
						t.Errorf("%s parameter %d: %s = %v, want %v", fn.Name(), i, attr, got, want)
					}
				}
				if fn.GetEnumAttributeAtIndex(i+1, llvm.AttributeKindID("noalias")).C != nil {
					t.Errorf("%s: overlapping memory must remain valid", fn.Name())
				}
			}
			// Integer and aggregate arguments do not get pointer attributes.
			if len(test.access) < len(test.params) && fn.GetEnumAttributeAtIndex(len(test.params), llvm.AttributeKindID("captures")).C != nil {
				t.Errorf("%s: pointer attribute attached to non-pointer argument", fn.Name())
			}
			if fn.GetEnumFunctionAttribute(llvm.AttributeKindID("memory")).C != nil {
				t.Errorf("%s: global/runtime memory effects were restricted", fn.Name())
			}
		}
	}
	if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
		t.Fatalf("invalid modeled declarations: %v\n%s", err, module.String())
	}
}

// The source of a raw comparison or copy remains unchanged, but an alias of a
// copy's destination does not. Check the optimization rather than only attrs.
func TestLLVMFunctionMemoryModelLoadForwarding(t *testing.T) {
	for _, helper := range []string{"runtime.memequal", "runtime.memmove", "runtime.memclrNoHeapPointers"} {
		for _, modeled := range []bool{false, true} {
			for _, alias := range []bool{false, true} {
				t.Run(helper+"/"+map[bool]string{false: "baseline", true: "modeled"}[modeled]+"/"+map[bool]string{false: "distinct", true: "alias"}[alias], func(t *testing.T) {
					oldModule := CurrentModule
					module := GlobalCtxt.NewModule("memory_model_optimization")
					CurrentModule = module
					defer func() { CurrentModule = oldModule; module.Dispose() }()
					ptr, i64 := GlobalCtxt.PointerType(0), GlobalCtxt.Int64Type()
					size := GlobalCtxt.IntType(int(types.PtrSize * 8))
					result := GlobalCtxt.VoidType()
					params := []llvm.Type{ptr, ptr, size}
					if helper == "runtime.memequal" {
						result = GlobalCtxt.Int8Type()
					} else if helper == "runtime.memclrNoHeapPointers" {
						params = []llvm.Type{ptr, size}
					}
					sig := llvmFuncSignature{Type: llvm.FunctionType(result, params, false), ClosureContextIndex: -1}
					var callee llvm.Value
					if modeled {
						callee = getOrInsertLLVMFunction(helper, sig, goABIInternalCallConv)
					} else {
						callee = llvm.AddFunction(module, helper, sig.Type)
						callee.SetFunctionCallConv(goABIInternalCallConv)
					}
					caller := llvm.AddFunction(module, "probe", llvm.FunctionType(i64, []llvm.Type{ptr}, false))
					b := GlobalCtxt.NewBuilder()
					defer b.Dispose()
					b.SetInsertPointAtEnd(llvm.AddBasicBlock(caller, "entry"))
					local := b.CreateAlloca(i64, "local")
					b.CreateStore(llvm.ConstInt(i64, 7, false), local)
					other := caller.Param(0)
					if alias {
						other = local
					}
					args := []llvm.Value{other, local, llvm.ConstInt(size, 8, false)}
					if helper == "runtime.memclrNoHeapPointers" {
						args = []llvm.Value{local, llvm.ConstInt(size, 8, false)}
					}
					call := b.CreateCall(sig.Type, callee, args, "")
					call.SetInstructionCallConv(goABIInternalCallConv)
					b.CreateRet(b.CreateLoad(i64, local, "loaded"))
					options := llvm.NewPassBuilderOptions()
					defer options.Dispose()
					options.SetVerifyEach(true)
					if err := module.RunPasses("function(instcombine,gvn)", llvm.TargetMachine{}, options); err != nil {
						t.Fatal(err)
					}
					wantForward := modeled && (helper == "runtime.memequal" || helper == "runtime.memmove" && !alias)
					if got := strings.Contains(caller.String(), "ret i64 7"); got != wantForward {
						t.Fatalf("load forwarded = %v, want %v\n%s", got, wantForward, module.String())
					}
				})
			}
		}
	}
}
