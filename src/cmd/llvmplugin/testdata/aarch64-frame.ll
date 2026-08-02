target triple = "aarch64-apple-darwin-goobj"

declare goabiinternal void @"runtime.GC"()

; Go indirect-call code words originate as uintptr values, not GC pointers.
; Materialize the callable pointer only at the call so EntryArgs contains the
; data pointer while ordinary statepoint liveness excludes the call-only code
; pointer.
define goabiinternal ptr @aarch64_pointer_and_code_live(
    i64 %callee_bits, ptr %pointer) #0 gc "goallc" {
entry:
  %callee = inttoptr i64 %callee_bits to ptr
  call goabiinternal void %callee()
  %value = load i8, ptr %pointer, align 1
  %used = icmp ne i8 %value, 0
  %result = select i1 %used, ptr %pointer, ptr %pointer
  ret ptr %result
}

define goabi0 ptr @aarch64_abi0_pointer_result(ptr %pointer) #0 gc "goallc" {
entry:
  %buf = alloca [8192 x i8], align 16
  %slot = getelementptr inbounds [8192 x i8], ptr %buf, i64 0, i64 8191
  store volatile i8 1, ptr %slot, align 1
  ret ptr %pointer
}

define goabiinternal ptr @aarch64_stack_pointer_arg(
    i64 %a0, i64 %a1, i64 %a2, i64 %a3,
    i64 %a4, i64 %a5, i64 %a6, i64 %a7,
    i64 %a8, i64 %a9, i64 %a10, i64 %a11,
    i64 %a12, i64 %a13, i64 %a14, i64 %a15,
    ptr %pointer) #0 gc "goallc" {
entry:
  %buf = alloca [8192 x i8], align 16
  %slot = getelementptr inbounds [8192 x i8], ptr %buf, i64 0, i64 8191
  store volatile i8 1, ptr %slot, align 1
  ret ptr %pointer
}

; Sub-word integer arguments occupy W registers, but their ABI homes retain
; their original one- and two-byte widths and offsets.
define goabiinternal void @aarch64_subword_homes(i8 %a, i16 %b) #0 gc "goallc" {
entry:
  call goabiinternal void @"runtime.GC"()
  ret void
}

; The i64 register argument's home starts beyond the 8-byte scaled-uimm12
; limit (32760), forcing the frameless morestack path to materialize SP+32776.
define goabiinternal i64 @aarch64_large_arg_home(
    [4096 x i64] %stackarg, i64 %regarg) #0 gc "goallc" {
entry:
  call goabiinternal void @"runtime.GC"()
  ret i64 %regarg
}

attributes #0 = { "frame-pointer"="non-leaf" "go-stack-growth-statepoint" }
