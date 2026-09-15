// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/base"

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

// llvmFunctionManager owns the association between logical Go function names,
// LLVM declarations and their semantic attributes. Declarations remain lazy,
// and CurrentModule is the only cache of LLVM values. In particular, no cached
// handle can outlive a module or survive replacement of a provisional signature.
type llvmFunctionManager struct {
	models map[string]llvmFunctionModel
}

var llvmFunctions = llvmFunctionManager{models: map[string]llvmFunctionModel{
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
}}

// name is the logical Go name; referenceName may contain GoObj builtin or
// linkname encoding. Attribute lookup must not depend on the storage spelling.
func (m *llvmFunctionManager) getOrInsert(name, referenceName string, sig llvmFuncSignature, cc llvm.CallConv) llvm.Value {
	storageName := llvmFunctionStorageName(referenceName, cc)
	fn := CurrentModule.NamedFunction(storageName)
	if fn.IsNil() {
		fn = llvm.AddFunction(CurrentModule, storageName+".goallc.final", sig.Type)
		if placeholder := CurrentModule.NamedGlobal(storageName); !placeholder.IsNil() {
			// An OpAddr may have needed the code address before this function
			// reached the compile queue. Opaque pointers let the provisional
			// global be replaced by the correctly typed function definition.
			placeholder.ReplaceAllUsesWith(fn)
			placeholder.EraseFromParentAsGlobal()
		}
		fn.SetName(storageName)
	} else if got := fn.GlobalValueType(); got != sig.Type {
		if fn.BasicBlocksCount() != 0 {
			base.Fatalf("conflicting LLVM function type for definition %s", name)
		}
		// Compiler data can refer to an ABI function before AuxCall exposes
		// its exact signature. Replace that provisional declaration now.
		replacement := llvm.AddFunction(CurrentModule, storageName+".goallc.final", sig.Type)
		fn.ReplaceAllUsesWith(replacement)
		fn.EraseFromParentAsFunction()
		replacement.SetName(storageName)
		fn = replacement
	}
	configureLLVMFunction(fn, sig, cc)
	if m.models[name].noReturn {
		fn.AddFunctionAttr(GlobalCtxt.CreateEnumAttribute(llvm.AttributeKindID("noreturn"), 0))
	}
	return fn
}

// configureCall binds call-only contracts after ordinary ABI configuration.
// All direct runtime call paths use this entry, regardless of whether they
// originated in SSA or were introduced by LLVM IR emission.
func (m *llvmFunctionManager) configureCall(name string, call llvm.Value) {
	// These contracts describe the ABIInternal entry. An ABI0 wrapper may
	// have different stack-growth and safepoint behavior.
	if m.models[name].gcLeafCall && call.InstructionCallConv() == goABIInternalCallConv {
		markLLVMGCLeafCall(call)
	}
}
