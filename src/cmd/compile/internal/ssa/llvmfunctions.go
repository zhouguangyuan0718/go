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
// An absent model is conservative. Unconditional properties are attached to
// the function so every direct caller sees the same contract. GC leaf describes
// calls to the function, not how calls inside its body should be compiled.
//
// Future argument-dependent properties belong on the call, never on the shared
// declaration. Parameter properties must use the lowered ABI signature.
type llvmFunctionModel struct {
	noReturn   bool
	gcLeaf     bool
	noFree     bool
	noCallback bool
	noUnwind   bool

	// Attributes for the known ABIInternal signature. Parameter entries use
	// zero-based argument positions; the binding below uses LLVM's 1-based indices.
	parameters [][]llvm.Attribute
}

// llvmFunctionManager owns the association between exact LLVM function names,
// LLVM declarations and their semantic attributes. Declarations remain lazy,
// and CurrentModule is the only cache of LLVM values. In particular, no cached
// handle can outlive a module or survive replacement of a provisional signature.
type llvmFunctionManager struct {
	models map[string]llvmFunctionModel
}

// Attributes belong to GlobalCtxt and can be reused across modules. Unlike
// function handles, they survive declaration replacement and module disposal.
var (
	llvmNoReturnAttribute     = llvmModelAttribute("noreturn")
	llvmGCLeafAttribute       = GlobalCtxt.CreateStringAttribute(goGCLeafFunctionAttr, "")
	llvmNoFreeAttribute       = llvmModelAttribute("nofree")
	llvmNoCallbackAttribute   = llvmModelAttribute("nocallback")
	llvmNoUnwindAttribute     = llvmModelAttribute("nounwind")
	llvmCapturesNoneAttribute = llvmModelAttribute("captures")
	llvmReadOnlyAttribute     = llvmModelAttribute("readonly")
	llvmWriteOnlyAttribute    = llvmModelAttribute("writeonly")
	llvmFunctions             = newLLVMFunctionManager()
)

func newLLVMFunctionManager() llvmFunctionManager {
	// The raw assembly helpers in runtime/memmove_*.s, memclr_*.s and
	// internal/bytealg/equal_*.s neither capture argument pointers, free memory,
	// call back into Go, nor unwind through LLVM EH. Their implementations may
	// read CPU flags, and arm64 memclr updates its cached ZVA block size, so
	// parameter access modes do not imply an argmem-only function effect.
	readPointer := []llvm.Attribute{llvmCapturesNoneAttribute, llvmReadOnlyAttribute}
	writePointer := []llvm.Attribute{llvmCapturesNoneAttribute, llvmWriteOnlyAttribute}
	models := map[string]llvmFunctionModel{
		// These raw helpers already have a GC-leaf call contract in both the SSA
		// static-call path and the dedicated memory-operation lowering paths.
		"runtime.memmove": {
			gcLeaf: true, noFree: true, noCallback: true, noUnwind: true,
			parameters: [][]llvm.Attribute{writePointer, readPointer},
		},
		"runtime.memequal": {
			gcLeaf: true, noFree: true, noCallback: true, noUnwind: true,
			parameters: [][]llvm.Attribute{readPointer, readPointer},
		},
		"runtime.memclrNoHeapPointers": {
			gcLeaf: true, noFree: true, noCallback: true, noUnwind: true,
			parameters: [][]llvm.Attribute{writePointer},
		},
		// The variable-length equality entry reads its size from the closure
		// context and tail-calls the same assembly comparison body.
		"runtime.memequal_varlen": {
			gcLeaf: true, noFree: true, noCallback: true, noUnwind: true,
			parameters: [][]llvm.Attribute{readPointer, readPointer},
		},
		// String arguments are aggregate values in LLVM IR, not pointer
		// parameters. The assembly only compares bytes (and may read CPU flags).
		"runtime.cmpstring": {gcLeaf: true, noFree: true, noCallback: true, noUnwind: true},
		"runtime.wbMove":    {gcLeaf: true},
		"runtime.wbZero":    {gcLeaf: true},

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
	model := m.models[fn.Name()]
	if model.noReturn {
		fn.AddFunctionAttr(llvmNoReturnAttribute)
	}
	// These contracts also hold for the NOSPLIT ABI wrappers.
	if model.gcLeaf {
		fn.AddFunctionAttr(llvmGCLeafAttribute)
	}
	if model.noFree {
		fn.AddFunctionAttr(llvmNoFreeAttribute)
	}
	if model.noCallback {
		fn.AddFunctionAttr(llvmNoCallbackAttribute)
	}
	if model.noUnwind {
		fn.AddFunctionAttr(llvmNoUnwindAttribute)
	}
	// ABI0 pointer parameters denote byval argument slots, not the original
	// pointees. Only ABIInternal uses the parameter contracts below.
	if cc == goABIInternalCallConv {
		for i, parameter := range model.parameters {
			for _, attribute := range parameter {
				fn.AddAttributeAtIndex(i+1, attribute)
			}
		}
	}
	return fn
}

func llvmModelAttribute(name string) llvm.Attribute {
	kind := llvm.AttributeKindID(name)
	if kind == 0 {
		base.Fatalf("unknown LLVM runtime model attribute %q", name)
	}
	// captures is an integer-valued attribute; zero is captures(none).
	return GlobalCtxt.CreateEnumAttribute(kind, 0)
}
