# GoALLC CPU feature registry

`registry.go` is the authoring source for compiler profiles, LLVM plugin
profiles, runtime snapshot bits, and SIMD descriptor aliases. Regenerate and
check the committed outputs with:

```
go generate cmd/llvmplugin/cpufeatures
go test cmd/llvmplugin/cpufeatures
```

The generator also accepts `-check` (no writes) and `-root` for an explicit Go
source tree. It has no external dependencies. Normal compiler/plugin/runtime
builds use the committed outputs and do not execute the generator.

## Contracts

- Feature bits are effective runtime predicates, after `GODEBUG`. Existing
  bit numbers are an append-only object/runtime ABI, including INITIALIZED.
- `Provides` is an instruction-capability relation, not a Boolean implication.
  For example, an AVX512 variant can use AVX2 instructions without replacing
  an independently disabled AVX2 source predicate with true.
- LLVM target-feature lists are built from the capability dependencies in
  stable order. This refactor preserves existing bundles, including FMA's
  existing target-feature and capability contract.
- Profile order controls FMV subset enumeration. `requestOrder` preserves
  the existing frontend attribute order, which differs from subset order.
- `SIMDAliases` records the existing descriptor mapping. A new operation or
  ISA extension needs an explicit lowering audit; aliases do not establish
  that every instruction from an extension has already been implemented.

## Function planning

`ssa2llvm_cpu.go` builds one immutable `llvmCPUFeaturePlan` before LLVM value
emission. It records entry assumptions and their origin, source predicates
selected for specialization, and requirements keyed by SSA value ID.
Generated SIMD, scalar/atomic compiler intrinsics, and wide register-ABI
calls all use that plan. Guard-path queries are shared by block/requirement.
Emission marks the planned guard loads and requirement-bearing instructions,
then writes the multiversion attribute once.

The existing policies remain distinct: unguarded wide-register calls raise
the function floor; unsupported unguarded generated operations remain
fail-closed. A Midway width floor is an entry assumption, not a request for a
second width dispatcher. Native SSA entry-block facts are preserved rather
than reinterpreted as effective predicates.

The LLVM plugin remains the only owner of specialization, variant capability
validation and runtime dispatch. This refactor does not change source-level
constant-folding requirements, add requirement anchors, change FMV subset
enumeration, or change resolver/Go ABI behavior.
