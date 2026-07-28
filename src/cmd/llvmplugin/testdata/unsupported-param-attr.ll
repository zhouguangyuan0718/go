target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @closure_callee(ptr nest)

define goabiinternal void @unsupported_param_attr(ptr %context) {
entry:
  call goabiinternal void @closure_callee(ptr nest %context)
  ret void
}
