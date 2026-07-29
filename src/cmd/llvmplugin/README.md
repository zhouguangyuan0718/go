# GoALLC LLVM pre-codegen plugin

This directory owns the GoALLC-specific LLVM pass pipeline. It deliberately
lives in the Go repository: LLVM only supplies the generic `llc`
`-load-pass-plugin` and pre-codegen callback mechanism.

`GoALLCStatepointPlugin.cpp` adapts that callback to
`goallc::runPreCodeGenPipeline`. GoALLC transformations belong in
`GoALLCPreCodeGen.cpp` or other sources in this directory, not in the LLVM
repository. The core entry point is separate from the plugin adapter so a
future in-process `cmd/compile` integration can call the same pipeline without
going through `llc`.

The SSA-to-LLVM lowering owns the function-level contract: Go ABI definitions
carry `gc "goallc"` and the `go-stack-growth-statepoint` attribute. Loading this
plugin registers the named GC strategy and its metadata printer; the rewrite
pass consumes the frontend markers rather than adding them.

The initial statepoint pass handles ordinary calls in Go ABI functions. It
computes pointer liveness with a backwards CFG dataflow analysis, assigns
stable callsite IDs, emits `gc.statepoint` and `gc.relocate`, and respects
`gc-leaf-function`. Pointer classification is conservative and independent of
LLVM address spaces. The initial implementation keeps live `alloca` addresses
and alloca-derived pointer values in `gc-live`; it does not try to predict
whether instruction selection will rematerialize a frame address or spill a
materialized pointer value.

The first implementation intentionally fails closed for live pointer
aggregates, relocation paths that require new PHIs, `invoke`, call operand
bundles, call parameter attributes, `musttail`, and non-leaf inline assembly.
Base and derived pointers share a stack-map location for this non-moving-heap
phase.

`GoALLCStackMapPrinter.cpp` is the Go-owned boundary between LLVM Machine
StackMaps and GoObj. It uses the standard
`AsmPrinter -> GCMetadataPrinter::emitStackMaps` hook and copies only raw
machine locations into `MCContext`. LLVM's generic `StackMaps.cpp` has no
GoALLC or GoObj branch. LLVM records GoObj statepoint callsites at the CALL
start, matching Go's `PCDATA_StackMapIndex` convention without a command-line
mode. The frontend's stack-growth attribute
asks LLVM to express the late-generated `runtime.morestack` call as a physical
MIR `STATEPOINT` with empty deopt, GC pointer, GC alloca, and GC relocation
sections. Its empty locals bitmap starts at the morestack CALL, so ordinary and
stack-growth calls use the same Machine StackMaps pipeline without relying on
a return-PC convention. GoObj functions that already contain a Machine
`STATEPOINT` but lack the frontend stack-growth attribute fail closed; there is
no slow-path reset-label fallback for a raw morestack call.
GoObj emits the currently constant safe `PCDATA_UnsafePoint` table first and
the statepoint-derived `PCDATA_StackMapIndex` table second, as required by
their Go ABI indexes 0 and 1.
The GoObj writer interprets locations after final layout: an
`Indirect [SP+offset]` location contributes a locals pointer bit, while a
`Direct SP+offset` stack address does not. This permits conservative IR
tracking without confusing an alloca's address with pointer data stored in the
alloca.

For AArch64 GoObj, target frame lowering uses Go's frame-chain layout instead
of the platform ABI frame record: LR is at `SP+0`, this function writes its FP
link at `SP-8` for a future callee, and the caller's existing FP-link word is at
the top of this function's physical frame. The writer derives
`_func.locals = StackSize - 8` and the locals bitmap range
`[SP+8, SP+StackSize-8)` from the final physical `StackSize`. No separate
lowering-to-writer locals-size channel is part of the contract. As in the
native arm64 backend, small frames atomically save LR while updating SP with a
pre-indexed store; frames above `0xf0` compute NewSP and save `(FP, LR)` before
moving SP so asynchronous traceback cannot observe a half-built frame.

The first phase emits count-aligned, empty `FUNCDATA_ArgsPointerMaps`; it does
not yet classify Go ABI argument home slots. `FUNCDATA_StackObjects` is also
not implemented, so pointer-containing address-taken alloca storage is outside
the supported first phase. Tracking an alloca address across a statepoint does
not describe the pointer fields stored inside that object. Both stack objects
and non-empty argument maps remain P0 follow-ups. These boundaries should be
expanded without moving Go policy into LLVM's generic StackMaps implementation.

Build, test, and install it into the LLVM payload that contains `llc`:

```sh
LLVM_PAYLOAD=/path/to/goallc-llvm
PLUGIN_BUILD=/path/to/empty/plugin-build

cmake -S "$GOROOT/src/cmd/llvmplugin" -B "$PLUGIN_BUILD" -G Ninja \
  -DLLVM_DIR="$LLVM_PAYLOAD/lib/cmake/llvm" \
  -DCMAKE_INSTALL_PREFIX="$LLVM_PAYLOAD"
cmake --build "$PLUGIN_BUILD"
ctest --test-dir "$PLUGIN_BUILD" --output-on-failure
cmake --install "$PLUGIN_BUILD"
```

Build the canonical `cmd/objview` from this checkout and pass
`-DGOALLC_OBJVIEW_EXECUTABLE=/path/to/objview` at configure time. This enables
the structured multiple-call test, which verifies both numbered PCDATA tables,
the map selected at each CALL, and the corresponding locals pointer bitmaps.

The installed file is `lib/GoALLCStatepoints.dylib` on Darwin or
`lib/GoALLCStatepoints.so` on Linux. Do not build the plugin against a different
LLVM installation and copy it into the payload.
