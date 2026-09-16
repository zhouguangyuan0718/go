# Compiler-known runtime LLVM contracts

The registry in `src/cmd/compile/internal/ssa/llvmfunctions.go` binds contracts
by exact IR name, with independent function flags and precreated LLVM parameter
attributes. It does not classify names by prefix at lookup time.

## Audit scope

The current audit covers the 285 distinct function names declared in
`src/cmd/compile/internal/typecheck/_builtin/runtime.go` or referenced by literal
`LookupRuntimeFunc` calls in `src/cmd/compile/internal/ssagen/ssa.go`, plus the
function entries in `src/cmd/internal/goobj/mkbuiltin.go` (excluding the TLS variable).
124 have an explicit model; the remaining 161 are listed below.
Also modeled are `Goexit`, `memequal_varlen`, `memhash_varlen`, `panicmem`, `panicmemAddr`, `throw`, plus the existing
`os.Exit` and testing termination methods. Generated `mallocgcSmallNoScanSC*`
and `mallocgcSmallScanNoHeaderSC*` entries share the allocation exclusions below.
LLVM intrinsics and the private write-barrier record intrinsic retain their
separate lowering contracts.

## Added contracts and limits

| Implementations | Contracts |
| --- | --- |
| Raw assembly copy, zero, equality, string comparison and Duff devices | GC leaf, `nofree`, `nocallback`, `nounwind`; direct source/destination pointer operands have `captures(none)` with `readonly`/`writeonly`. |
| Fixed-size, float, complex and string equality in `runtime/alg.go` | `nounwind`; two read-only, non-capturing pointers. |
| Concrete memory, string, float and complex hashes in `runtime/alg.go`, including `memhash_varlen` | `nounwind`; one read-only, non-capturing pointer. Audit follows `internal/runtime/maps/memhash_*.go` and `runtime_hash*.go`. |
| Rune decoding/counting, complex division and scalar/soft-float conversions | `nounwind`; aggregate/scalar parameters get no pointer attributes. |
| `rand` and `rand32` | `nounwind`; random state remains mutable and observable. |
| `chanlen`, `chancap` | `nounwind`; read-only, non-capturing channel header pointer, including nil/timer-channel handling. |
| `selectsetpc` | `nounwind`; write-only, non-capturing output pointer. |
| Typed copy/clear, pointer-containing clear, cgo pointer checks, write-barrier entry points | GC leaf and `nounwind`. Audit follows `mbarrier.go`, `mbitmap.go`, `mwbbuf.go`, `cgocheck.go` and assembly entry points. Their slow paths operate on the system stack without a goroutine safe point. |
| Panic helpers, bounds failures, signal panic, `throw`, `block`, Goexit and existing termination APIs | `noreturn`. Recovery resumes an older frame; the helper does not return to its call site. No callback-free or GC-leaf promise follows. |

Function contracts other than memory-effect restrictions apply to both ABIInternal and ABI0 entries. Their NOSPLIT
ABI wrappers preserve these contracts. Parameter attributes only apply to
ABIInternal: ABI0 byval parameters denote argument slots, not the original
pointees, and return homes are additional parameters.

Memory effects describe the logical operation visible to callers. Stack growth,
stack-map processing and pointer relocation are implemented by the Go calling
convention and statepoints. They do not by themselves turn a logical read-only
operation into a writer. These contracts apply to both declarations and
runtime definitions: definition creation uses the same function manager before
adding basic blocks.

Concrete equality, deterministic memory/string hashes, rune scanning and channel
queries have `memory(read)` on ABIInternal entries. Pure numeric/soft-float
helpers and zero-width equality/hash helpers have `memory(none)`. These finite
operations also have `willreturn`. Reads are not restricted to argument memory:
hash seeds, CPU flags, string payloads and closure contexts must remain visible.
The three raw assembly comparisons additionally have `nosync`.

Go helpers remain non-leaf unless independently justified. The statepoint pass
does not use `memory(read)` or `memory(none)` as evidence of GC leaf. Surviving
calls still relocate live pointers, and the rewrite does not copy these memory
restrictions onto the statepoint intrinsic. Optimization tests verify repeated
call elimination and dead-call removal, while a backend regression verifies
that surviving hash/numeric calls still produce statepoints and relocations.

ABI0 wrappers keep unrestricted memory effects because their result homes are
caller-owned memory that must be written. Logical `willreturn` remains valid
for those wrappers.

Other helpers retain unrestricted memory effects: arm64 zeroing updates a
cache, floating-point and complex hashes update random state for NaNs, and bulk
barriers update GC metadata. Barrier destinations are not `writeonly`: old
pointers must be read before overwriting them. No nonnull, noalias, allocation
result, or dereferenceability promise is introduced.

The LLVM definitions of these attributes are in the
[LLVM language reference](https://llvm.org/docs/LangRef.html#function-attributes).
`nounwind` does not mean that a function has no runtime side effects or that
asynchronous faults cannot occur.

## Entries left without a new model

Each row lists exact runtime names (with the `runtime.` prefix omitted).
These exclusions preserve conservative behavior where the audited fast path
does not establish a contract for the full implementation and build modes.
They are not a claim that no more precise contract could ever be proved.

### Liveness

`KeepAlive` has a dedicated compiler liveness contract and a source fallback that references printing. Do not substitute memory-effect or capture assumptions for liveness.

`KeepAlive`.

### Coverage registration

`addCovMeta` forwards to `internal/coverage/rtcov.AddMeta`, which retains the metadata pointer and updates global registration structures.

`addCovMeta`.

### Instrumentation

Race, MSan, ASan and fuzzing hooks call external runtimes and can have callbacks and address-observation effects. No common leaf or capture contract is assumed across build modes.

`asanread`, `asanregisterglobals`, `asanwrite`, `libfuzzerHookEqualFold`, `libfuzzerHookStrCmp`, `libfuzzerTraceCmp1`, `libfuzzerTraceCmp2`, `libfuzzerTraceCmp4`, `libfuzzerTraceCmp8`, `libfuzzerTraceConstCmp1`, `libfuzzerTraceConstCmp2`, `libfuzzerTraceConstCmp4`, `libfuzzerTraceConstCmp8`, `msanmove`, `msanread`, `msanwrite`, `racefuncenter`, `racefuncexit`, `raceread`, `racereadrange`, `racewrite`, `racewriterange`.

### Interface caches

`iface.go` performs cache updates, allocations and possible assertion panics. Metadata can be retained in itabs/caches.

`assertE2I`, `assertE2I2`, `interfaceSwitch`, `typeAssert`.

### Channels and select

`chan.go` and `select.go` can park, synchronize with other goroutines, retain element addresses in sudogs, invoke timer machinery or panic. Nonblocking variants share these paths.

`chanrecv1`, `chanrecv2`, `chansend1`, `closechan`, `makechan`, `makechan64`, `selectgo`, `selectnbrecv`, `selectnbsend`.

### Pointer validation

The checks inspect pointer addresses, runtime allocation metadata and failure paths. Unsafe slice/string checks can panic and run user defers. Do not infer raw-pointer capture or leaf contracts from the successful path.

`checkptrAlignment`, `checkptrArithmetic`, `unsafeslicecheckptr`, `unsafestringcheckptr`.

### Allocation and conversion

These helpers allocate, may run GC/instrumentation/error paths, retain type/data pointers, or return aliases of scratch/input storage. `slicecopy` and `typedslicecopy` also call sanitizer hooks. A contract common to all supported build modes is not inferred from the fast path.

`concatbyte2`, `concatbyte3`, `concatbyte4`, `concatbyte5`, `concatbytes`, `concatstring2`, `concatstring3`, `concatstring4`, `concatstring5`, `concatstrings`, `convT`, `convT16`, `convT32`, `convT64`, `convTnoptr`, `convTslice`, `convTstring`, `growslice`, `growsliceBuf`, `growsliceBufNoAlias`, `growsliceNoAlias`, `intstring`, `makeslice`, `makeslice64`, `makeslicecopy`, `mallocgc`, `mallocgcTinySC2`, `moveSlice`, `moveSliceNoCap`, `moveSliceNoCapNoScan`, `moveSliceNoScan`, `newobject`, `slicebytetostring`, `slicebytetostringtmp`, `slicecopy`, `slicerunetostring`, `stringtoslicebyte`, `stringtoslicerune`, `typedslicecopy`.

### Scheduling and defer

These entries schedule execution, register callbacks, or observe/change panic/defer state. `gorecover` also inspects physical frames. Preserve their existing dedicated lowering contracts.

`deferproc`, `deferprocStack`, `deferprocat`, `deferrangefunc`, `deferreturn`, `gorecover`, `goschedguarded`, `newproc`.

### Dynamic equality and hashing

`alg.go` dispatches through type metadata and can panic for uncomparable or unhashable dynamic types. The panic path invokes user defers; the simple scalar-helper contracts do not cover it.

`efaceeq`, `ifaceeq`, `interequal`, `interhash`, `nilinterequal`, `nilinterhash`.

### Integer division

Zero divisors reach panic handling and user defers. Unlike software floating point, these are not unconditionally arithmetic-only helpers.

`int64div`, `int64mod`, `uint64div`, `uint64mod`.

### Maps

Map operations can allocate, synchronize, retain keys/types, call generated hash/equality functions, or panic. Type metadata and caches do not have a uniform read-only/non-capturing contract.

`makemap`, `makemap64`, `makemap_small`, `mapIterNext`, `mapIterStart`, `mapaccess1`, `mapaccess1_fast32`, `mapaccess1_fast64`, `mapaccess1_faststr`, `mapaccess1_fat`, `mapaccess2`, `mapaccess2_fast32`, `mapaccess2_fast64`, `mapaccess2_faststr`, `mapaccess2_fat`, `mapassign`, `mapassign_fast32`, `mapassign_fast32ptr`, `mapassign_fast64`, `mapassign_fast64ptr`, `mapassign_faststr`, `mapclear`, `mapdelete`, `mapdelete_fast32`, `mapdelete_fast64`, `mapdelete_faststr`.

### Stack and indirect-call trampolines

The morestack entries switch and resume stacks through backend-specific control flow. Retpolines transfer to arbitrary indirect targets. They do not have ordinary GC-leaf or no-callback contracts; morestackc is separately modeled as noreturn because its Go body unconditionally throws.

`morestack`, `morestack_noctxt`, `retpolineAX`, `retpolineBP`, `retpolineBX`, `retpolineCX`, `retpolineDI`, `retpolineDX`, `retpolineR10`, `retpolineR11`, `retpolineR12`, `retpolineR13`, `retpolineR14`, `retpolineR15`, `retpolineR8`, `retpolineR9`, `retpolineSI`.

### Printing

`print.go` reaches `write` and the externally replaceable `overrideWrite` callback in `time_nofake.go`. Pointer printing also observes pointer addresses. Even small print helpers cannot be modeled as raw no-callback functions.

`printbool`, `printcomplex128`, `printcomplex64`, `printeface`, `printfloat32`, `printfloat64`, `printhex`, `printiface`, `printint`, `printlock`, `printnl`, `printpointer`, `printquoted`, `printslice`, `printsp`, `printstring`, `printuint`, `printuintptr`, `printunlock`.

### Legacy declaration

`throwinit` remains in the compiler builtin declarations, but has no runtime implementation in this tree; do not assign an implementation-derived contract.

`throwinit`.
