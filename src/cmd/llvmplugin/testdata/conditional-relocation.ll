target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @callee()

define goabiinternal i64 @conditional_safepoint(ptr %p, i1 %take_call) {
entry:
  br i1 %take_call, label %call, label %skip

call:
  call goabiinternal void @callee()
  br label %merge

skip:
  br label %merge

merge:
  %value = load i64, ptr %p, align 8
  ret i64 %value
}
