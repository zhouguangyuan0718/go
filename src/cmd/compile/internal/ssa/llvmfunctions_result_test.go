// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"strings"
	"testing"

	"github.com/goallc/go-llvm"
)

// Exercise optimizer consequences, including 32-bit Go int ranges. Keep an
// unmodeled control so a vacuous predicate cannot make the regression pass.
func TestLLVMFunctionResultEffects(t *testing.T) {
	for _, bits := range []int{32, 64} {
		for _, name := range []string{"runtime.mallocgc", "runtime.convT16", "runtime.cmpstring", "runtime.chanlen", "runtime.chancap", "runtime.countrunes"} {
			for _, modeled := range []bool{false, true} {
				func() {
					old := CurrentModule
					m := GlobalCtxt.NewModule("result_effects")
					CurrentModule = m
					defer func() { CurrentModule = old; m.Dispose() }()
					ptr, i1 := GlobalCtxt.PointerType(0), GlobalCtxt.Int1Type()
					size := GlobalCtxt.IntType(bits)
					str := llvm.StructType([]llvm.Type{ptr, size}, false)
					result, params := size, []llvm.Type{ptr}
					switch name {
					case "runtime.mallocgc":
						result, params = ptr, []llvm.Type{size, ptr, GlobalCtxt.Int8Type()}
					case "runtime.convT16":
						result, params = ptr, []llvm.Type{GlobalCtxt.Int16Type()}
					case "runtime.cmpstring":
						params = []llvm.Type{str, str}
					case "runtime.countrunes":
						params = []llvm.Type{str}
					}
					sig := llvmFuncSignature{Type: llvm.FunctionType(result, params, false), ClosureContextIndex: -1}
					var callee llvm.Value
					if modeled {
						callee = getOrInsertLLVMFunction(name, sig, goABIInternalCallConv)
					} else {
						callee = llvm.AddFunction(m, name, sig.Type)
						callee.SetFunctionCallConv(goABIInternalCallConv)
					}
					fn := llvm.AddFunction(m, "probe", llvm.FunctionType(i1, params, false))
					b := GlobalCtxt.NewBuilder()
					defer b.Dispose()
					b.SetInsertPointAtEnd(llvm.AddBasicBlock(fn, "entry"))
					args := fn.Params()
					if name == "runtime.mallocgc" {
						args[0] = llvm.ConstInt(size, 16, false)
					}
					call := b.CreateCall(sig.Type, callee, args, "value")
					call.SetInstructionCallConv(goABIInternalCallConv)
					var bad llvm.Value
					switch name {
					case "runtime.mallocgc", "runtime.convT16":
						bad = b.CreateICmp(llvm.IntEQ, call, llvm.ConstNull(ptr), "bad")
					case "runtime.cmpstring":
						low := b.CreateICmp(llvm.IntSLT, call, llvm.ConstInt(size, ^uint64(0), true), "low")
						high := b.CreateICmp(llvm.IntSGT, call, llvm.ConstInt(size, 1, false), "high")
						bad = b.CreateOr(low, high, "bad")
					default:
						bad = b.CreateICmp(llvm.IntSLT, call, llvm.ConstInt(size, 0, false), "bad")
					}
					if name == "runtime.mallocgc" {
						i64 := GlobalCtxt.Int64Type()
						objectSizeType := llvm.FunctionType(i64, []llvm.Type{ptr, i1, i1, i1}, false)
						objectSize := llvm.AddFunction(m, "llvm.objectsize.i64.p0", objectSizeType)
						n := b.CreateCall(objectSizeType, objectSize, []llvm.Value{call, llvm.ConstInt(i1, 0, false), llvm.ConstInt(i1, 1, false), llvm.ConstInt(i1, 0, false)}, "size")
						bad = b.CreateOr(bad, b.CreateICmp(llvm.IntNE, n, llvm.ConstInt(i64, 16, false), "wrongsize"), "invalid")
					}
					b.CreateRet(bad)
					opts := llvm.NewPassBuilderOptions()
					defer opts.Dispose()
					opts.SetVerifyEach(true)
					if err := m.RunPasses("function(instcombine,simplifycfg)", llvm.TargetMachine{}, opts); err != nil {
						t.Fatal(err)
					}
					if got := strings.Contains(fn.String(), "ret i1 false"); got != modeled {
						t.Fatalf("%s bits=%d modeled=%v: unexpected predicate\n%s", name, bits, modeled, m.String())
					}
					if name == "runtime.mallocgc" {
						// Nonnull is not permission to remove an allocation or its safe point.
						if !strings.Contains(fn.String(), "call goabiinternal") {
							t.Fatal("allocation removed")
						}
						if modeled && !strings.Contains(callee.String(), "allocsize(0)") {
							t.Fatal("missing allocation size")
						}
					}
				}()
			}
		}
	}
}

func TestLLVMFunctionABI0ResultAttributes(t *testing.T) {
	old := CurrentModule
	m := GlobalCtxt.NewModule("abi0_results")
	CurrentModule = m
	defer func() { CurrentModule = old; m.Dispose() }()
	ptr := GlobalCtxt.PointerType(0)
	// ABI0 takes argument/result slots and returns void. Result contracts must
	// not leak onto the function's void return or the allocation's size slot.
	for _, name := range []string{"runtime.mallocgc", "runtime.convT16", "runtime.cmpstring", "runtime.chanlen"} {
		sig := llvmFuncSignature{Type: llvm.FunctionType(GlobalCtxt.VoidType(), []llvm.Type{ptr, ptr, ptr, ptr}, false), ClosureContextIndex: -1}
		fn := getOrInsertLLVMFunction(name, sig, goABI0CallConv)
		for _, attr := range []string{"nonnull", "dereferenceable", "range"} {
			if fn.GetEnumAttributeAtIndex(0, llvm.AttributeKindID(attr)).C != nil {
				t.Fatalf("%s has result %s", name, attr)
			}
		}
		if fn.GetEnumFunctionAttribute(llvm.AttributeKindID("allocsize")).C != nil {
			t.Fatalf("%s has allocsize", name)
		}
	}
	if err := llvm.VerifyModule(m, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
}
