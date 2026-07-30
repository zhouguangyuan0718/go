target triple = "x86_64-unknown-linux-goobj"

; X86 has nine integer-class parameter registers. The final typed byval
; parameter is therefore a complete Go ABI stack object at argp+0. Its pointee
; is live at the synthetic morestack statepoint and must be encoded in
; FUNCDATA_ArgsPointerMaps, not FUNCDATA_LocalsPointerMaps.
define goabiinternal ptr @x86_byval_stack_pointer(
    i64 %a0, i64 %a1, i64 %a2, i64 %a3, i64 %a4,
    i64 %a5, i64 %a6, i64 %a7, i64 %a8,
    ptr byval(ptr) align 8 %pointer.byval)
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  %buf = alloca [5000 x i8], align 16
  %slot = getelementptr inbounds [5000 x i8], ptr %buf, i64 0, i64 4999
  store volatile i8 1, ptr %slot, align 1
  %pointer = load ptr, ptr %pointer.byval, align 8
  ret ptr %pointer
}
