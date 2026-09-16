// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"fmt"
	"strings"
	"testing"

	"github.com/goallc/go-llvm"
)

func TestLLVMFunctionAllocationCallEffects(t *testing.T) {
	for _, helper := range []string{"runtime.mallocgc", "runtime.mallocgcTinySC2", "runtime.mallocgcSmallNoScanSC1", "runtime.mallocgcSmallScanNoHeaderSC1"} {
		for _, size := range []int64{-1, 0, 8} {
			for _, zero := range []int64{-1, 0, 1} {
				for _, use := range []string{"unused", "read", "alias"} {
					t.Run(fmt.Sprintf("%s/size=%d/zero=%d/%s", helper, size, zero, use), func(t *testing.T) {
						old := CurrentModule
						m := GlobalCtxt.NewModule("allocation_effects")
						CurrentModule = m
						defer func() { CurrentModule = old; m.Dispose() }()
						i8, i64, ptr := GlobalCtxt.Int8Type(), GlobalCtxt.Int64Type(), GlobalCtxt.PointerType(0)
						sig := llvmFuncSignature{Type: llvm.FunctionType(ptr, []llvm.Type{i64, ptr, i8}, false), ClosureContextIndex: -1}
						callee := getOrInsertLLVMFunction(helper, sig, goABIInternalCallConv)
						fn := llvm.AddFunction(m, "probe", llvm.FunctionType(i64, []llvm.Type{i64, i8, ptr}, false))
						b := GlobalCtxt.NewBuilder()
						defer b.Dispose()
						b.SetInsertPointAtEnd(llvm.AddBasicBlock(fn, "entry"))
						args := []llvm.Value{fn.Param(0), llvm.ConstNull(ptr), fn.Param(1)}
						if size >= 0 {
							args[0] = llvm.ConstInt(i64, uint64(size), false)
						}
						if zero >= 0 {
							args[2] = llvm.ConstInt(i8, uint64(zero), false)
						}
						call := b.CreateCall(sig.Type, callee, args, "p")
						call.SetInstructionCallConv(goABIInternalCallConv)
						llvmFunctions.bindCall(call, callee, args, goABIInternalCallConv, nil)
						// Conditional information must not leak into shared declarations.
						if callee.GetEnumFunctionAttribute(llvm.AttributeKindID("allockind")).C != nil || callee.GetEnumAttributeAtIndex(0, llvm.AttributeKindID("noalias")).C != nil {
							t.Fatal("allocation contract leaked into declaration")
						}
						switch use {
						case "unused":
							b.CreateRet(llvm.ConstInt(i64, 0, false))
						case "read":
							// Zero size has no readable bytes; compare identity instead.
							if size == 0 {
								second := b.CreateCall(sig.Type, callee, args, "q")
								second.SetInstructionCallConv(goABIInternalCallConv)
								llvmFunctions.bindCall(second, callee, args, goABIInternalCallConv, nil)
								b.CreateRet(b.CreateZExt(b.CreateICmp(llvm.IntEQ, call, second, "same"), i64, "result"))
							} else {
								b.CreateRet(b.CreateLoad(i64, call, "loaded"))
							}
						case "alias":
							if size == 0 {
								b.CreateRet(llvm.ConstInt(i64, 0, false))
							} else {
								b.CreateStore(llvm.ConstInt(i64, 1, false), fn.Param(2))
								b.CreateStore(llvm.ConstInt(i64, 2, false), call)
								b.CreateRet(b.CreateLoad(i64, fn.Param(2), "loaded"))
							}
						}
						opts := llvm.NewPassBuilderOptions()
						defer opts.Dispose()
						opts.SetVerifyEach(true)
						if err := m.RunPasses("default<O2>", llvm.TargetMachine{}, opts); err != nil {
							t.Fatal(err)
						}
						ir := fn.String()
						folded := size > 0 && (use != "read" || zero == 1)
						if got := !strings.Contains(ir, "call goabiinternal"); got != folded {
							t.Fatalf("call elimination=%v want=%v\n%s", got, folded, m.String())
						}
						if folded {
							want := "ret i64 0"
							if use == "alias" {
								want = "ret i64 1"
							}
							if !strings.Contains(ir, want) {
								t.Fatalf("missing %s\n%s", want, ir)
							}
						}
					})
				}
			}
		}
	}
}

// The runtime type argument, not a guessed LLVM pointee type, determines size.
func TestLLVMFunctionNewObjectCalls(t *testing.T) {
	for _, n := range []int64{-1, 0, 128} {
		for _, cc := range []llvm.CallConv{goABIInternalCallConv, goABI0CallConv} {
			t.Run(fmt.Sprintf("size=%d/cc=%d", n, cc), func(t *testing.T) {
				old := CurrentModule
				m := GlobalCtxt.NewModule("newobject_effects")
				CurrentModule = m
				defer func() { CurrentModule = old; m.Dispose() }()
				ptr := GlobalCtxt.PointerType(0)
				result := ptr
				if cc == goABI0CallConv {
					result = GlobalCtxt.VoidType()
				}
				sig := llvmFuncSignature{Type: llvm.FunctionType(result, []llvm.Type{ptr}, false), ClosureContextIndex: -1}
				callee := getOrInsertLLVMFunction("runtime.newobject", sig, cc)
				fn := llvm.AddFunction(m, "probe", llvm.FunctionType(GlobalCtxt.VoidType(), []llvm.Type{ptr}, false))
				b := GlobalCtxt.NewBuilder()
				defer b.Dispose()
				b.SetInsertPointAtEnd(llvm.AddBasicBlock(fn, "entry"))
				args := []llvm.Value{fn.Param(0)}
				call := b.CreateCall(sig.Type, callee, args, "")
				call.SetInstructionCallConv(cc)
				descriptor := &Value{Op: OpArg}
				if n >= 0 {
					typ := types.NewArray(types.Types[types.TUINT8], n)
					types.CalcSize(typ)
					sym := &obj.LSym{Name: "type:allocation-test"}
					sym.NewTypeInfo().Type = typ
					descriptor = &Value{Op: OpAddr, Aux: sym}
				}
				source := &Value{Args: []*Value{descriptor}}
				llvmFunctions.bindCall(call, callee, args, cc, source)
				b.CreateRetVoid()
				want := n > 0 && cc == goABIInternalCallConv
				if got := strings.Contains(call.String(), "noalias"); got != want {
					t.Fatalf("noalias=%v want=%v\n%s", got, want, m.String())
				}
				opts := llvm.NewPassBuilderOptions()
				defer opts.Dispose()
				opts.SetVerifyEach(true)
				if err := m.RunPasses("default<O2>", llvm.TargetMachine{}, opts); err != nil {
					t.Fatal(err)
				}
				if got := !strings.Contains(fn.String(), "call goabi"); got != want {
					t.Fatalf("elided=%v want=%v\n%s", got, want, m.String())
				}
			})
		}
	}
}
