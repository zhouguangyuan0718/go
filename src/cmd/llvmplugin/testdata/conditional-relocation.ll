target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @callee()

define goabiinternal i64 @conditional_safepoint(ptr %p, i1 %take_call) "go-stack-growth-statepoint" gc "goallc" {
entry:
  br i1 %take_call, label %call, label %skip

call:
  call goabiinternal void @callee()
  br label %merge

skip:
  br label %merge

merge:
  %first = load i64, ptr %p, align 8
  %second = load i64, ptr %p, align 8
  %sum = add i64 %first, %second
  ret i64 %sum
}

define goabiinternal i64 @conditional_phi_edge_use(
    ptr %p, ptr %q, i1 %take_call)
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  br i1 %take_call, label %call, label %skip

call:
  call goabiinternal void @callee()
  br label %merge

skip:
  br label %merge

merge:
  %selected = phi ptr [ %p, %call ], [ %q, %skip ]
  %value = load i64, ptr %selected, align 8
  ret i64 %value
}
