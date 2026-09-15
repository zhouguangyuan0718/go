// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/internal/goobj"
	"cmd/internal/obj"

	"github.com/goallc/go-llvm"
)

// llvmFunctionModel describes audited contracts of compiler-known functions.
// An absent model is conservative. Keep unconditional declaration properties
// separate from call properties: a GC-leaf call does not require compiling the
// helper's own implementation as a GC-leaf definition.
//
// Future argument-dependent properties belong on the call, never on the shared
// declaration. Parameter properties must use the lowered ABI signature.
type llvmFunctionModel struct {
	noReturn   bool
	gcLeafCall bool
}

// llvmFunctionManager owns the association between exact LLVM function names,
// LLVM declarations and their semantic attributes. Declarations remain lazy,
// and CurrentModule is the only cache of LLVM values. In particular, no cached
// handle can outlive a module or survive replacement of a provisional signature.
type llvmFunctionManager struct {
	models map[string]llvmFunctionModel
}

var llvmFunctions = newLLVMFunctionManager()

func newLLVMFunctionManager() llvmFunctionManager {
	models := map[string]llvmFunctionModel{
		// These raw helpers already have a GC-leaf call contract in both the SSA
		// static-call path and the dedicated memory-operation lowering paths.
		"runtime.memmove":  {gcLeafCall: true},
		"runtime.memequal": {gcLeafCall: true},
		"runtime.wbMove":   {gcLeafCall: true},
		"runtime.wbZero":   {gcLeafCall: true},

		// These APIs terminate the goroutine or process. Do not use the inliner's
		// NeverReturns heuristic: a callee can recover its own panic and return.
		"runtime.Goexit":            {noReturn: true},
		"os.Exit":                   {noReturn: true},
		"testing.(*common).Skip":    {noReturn: true},
		"testing.(*common).Skipf":   {noReturn: true},
		"testing.(*common).SkipNow": {noReturn: true},
		"testing.(*common).Fatal":   {noReturn: true},
		"testing.(*common).Fatalf":  {noReturn: true},
		"testing.(*common).FailNow": {noReturn: true},
	}
	m := llvmFunctionManager{models: make(map[string]llvmFunctionModel)}
	// Register the exact spellings emitted by the existing GoObj naming layer.
	// Builtin indices come from its table rather than being hard-coded here.
	// Plain names are used for definitions and linkshared references. Lookup
	// never strips suffixes or translates an IR name back into a Go symbol.
	for name, model := range models {
		for _, abi := range []obj.ABI{obj.ABIInternal, obj.ABI0} {
			cc := llvmCallConv(abi)
			m.models[llvmFunctionStorageName(name, cc)] = model
			m.models[llvmFunctionStorageName(name+goobj.LinknameSymbolSuffix, cc)] = model
			if builtin, ok := goobj.BuiltinSymbolName(name, int(abi)); ok {
				m.models[llvmFunctionStorageName(builtin, cc)] = model
			}
		}
	}
	return m
}

// name is the final IR function name, including any GoObj or ABI suffixes.
func (m *llvmFunctionManager) getOrInsert(name string, sig llvmFuncSignature, cc llvm.CallConv) llvm.Value {
	fn := CurrentModule.NamedFunction(name)
	if fn.IsNil() {
		fn = llvm.AddFunction(CurrentModule, name+".goallc.final", sig.Type)
		if placeholder := CurrentModule.NamedGlobal(name); !placeholder.IsNil() {
			// An OpAddr may have needed the code address before this function
			// reached the compile queue. Opaque pointers let the provisional
			// global be replaced by the correctly typed function definition.
			placeholder.ReplaceAllUsesWith(fn)
			placeholder.EraseFromParentAsGlobal()
		}
		fn.SetName(name)
	} else if got := fn.GlobalValueType(); got != sig.Type {
		if fn.BasicBlocksCount() != 0 {
			base.Fatalf("conflicting LLVM function type for definition %s", name)
		}
		// Compiler data can refer to an ABI function before AuxCall exposes
		// its exact signature. Replace that provisional declaration now.
		replacement := llvm.AddFunction(CurrentModule, name+".goallc.final", sig.Type)
		fn.ReplaceAllUsesWith(replacement)
		fn.EraseFromParentAsFunction()
		replacement.SetName(name)
		fn = replacement
	}
	configureLLVMFunction(fn, sig, cc)
	if m.models[fn.Name()].noReturn {
		fn.AddFunctionAttr(GlobalCtxt.CreateEnumAttribute(llvm.AttributeKindID("noreturn"), 0))
	}
	return fn
}

// configureCall binds call-only contracts after ordinary ABI configuration.
// All direct runtime call paths use this entry, regardless of whether they
// originated in SSA or were introduced by LLVM IR emission.
func (m *llvmFunctionManager) configureCall(call llvm.Value) {
	fn := call.CalledValue()
	if fn.IsAFunction().IsNil() {
		return
	}
	// These contracts describe the ABIInternal entry. An ABI0 wrapper may
	// have different stack-growth and safepoint behavior.
	if m.models[fn.Name()].gcLeafCall && call.InstructionCallConv() == goABIInternalCallConv {
		markLLVMGCLeafCall(call)
	}
}
