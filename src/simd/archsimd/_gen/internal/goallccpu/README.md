# Shared GoALLC CPU data

This package is part of the existing archsimd generation pipeline, not a
standalone generator. From `src/simd/archsimd/_gen`:

```
go run .                 # existing full pipeline, including CPU tables
go run . -cpu-only       # only CPU tables; no XED/ARM ISA inputs needed
go run . -check-cpu      # check CPU tables without writes or ISA inputs
go test ./internal/goallccpu ./simdgen/... ./sgutil/...
```

The existing `-goroot` destination and `-n` dry-run options apply to the CPU
stage too. Full generation still requires the usual ISA inputs.

`registry.go` is the hand-maintained feature/profile source. The SIMD
descriptor generator calls `ProfileForSIMD` directly using the upstream
operation's CPUFeature. There is no generated source dependency inside the
generator. The top-level driver calls `Generate` once before architecture
generation to emit three committed outputs:

- `cmd/compile/internal/ssa/ssa2llvm_cpu_profiles_gen.go`
- `cmd/llvmplugin/GoALLCCPUFeatures.def`
- `runtime/cpuflags_goallc_gen.go`

Normal compiler/plugin/runtime builds consume those files without running
the generator or importing this package.

Feature bits are effective runtime predicates after GODEBUG, with append-only
ABI positions. Provides describes instruction capabilities, never additional
true predicates. Target bundles and the two existing frontend/plugin profile
orders are preserved. Explicit SIMD aliases preserve reviewed lowering policy;
unknown extensions must be audited instead of inheriting a prefix match.

There are no SSA opcode names in this registry or a second scalar/atomic
opcode-to-feature table in the LLVM backend. SIMD requirements come from
upstream operation descriptors; scalar/atomic specialization follows Go's
existing hardware/fallback feature checks; wide-call requirements come from
the ABI. Ordinary LLVM intrinsics and atomics carry no ISA requirement
metadata: LLVM legalizes them for the selected target. The per-function
planner combines Go feature checks and SIMD/ABI requirements with entry
assumptions before LLVM emission. The LLVM plugin remains the only
owner of CPU specialization and dispatch. Midway width dispatch, unguarded
wide-call floors, fail-closed unguarded SIMD policy and FMV subset enumeration
are unchanged.
