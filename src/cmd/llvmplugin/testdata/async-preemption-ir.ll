target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @callee()
declare i64 @llvm.go.pointer.address.i64.p0(ptr)
declare ptr @llvm.go.pointer.from.address.p0.i64(i64)

define goabiinternal i64 @ordinary_loop(i64 %n) #0 gc "goallc" {
entry:
  br label %loop
loop:
  %i = phi i64 [ 0, %entry ], [ %next, %loop ]
  %next = add nuw i64 %i, 1
  %done = icmp eq i64 %next, %n
  br i1 %done, label %exit, label %loop
exit:
  ret i64 %next
}

define goabiinternal ptr @ordinary_pointer_roundtrip(ptr %p) #0 gc "goallc" {
entry:
  %address = call i64 @llvm.go.pointer.address.i64.p0(ptr %p)
  %pointer = call ptr @llvm.go.pointer.from.address.p0.i64(i64 %address)
  ret ptr %pointer
}

define goabiinternal void @ordinary_call() #0 gc "goallc" {
entry:
  call goabiinternal void @callee()
  ret void
}

define goabiinternal i64 @ordinary_alloca(i64 %v) #0 gc "goallc" {
entry:
  %slot = alloca i64, align 8
  store volatile i64 %v, ptr %slot, align 8
  %result = load volatile i64, ptr %slot, align 8
  ret i64 %result
}

define goabiinternal i64 @ordinary_atomic(ptr %p) #0 gc "goallc" {
entry:
  %v = load atomic i64, ptr %p seq_cst, align 8
  ret i64 %v
}

define goabiinternal <2 x i64> @ordinary_vector(<2 x i64> %v) #0 gc "goallc" {
entry:
  %result = add <2 x i64> %v, <i64 1, i64 1>
  ret <2 x i64> %result
}

attributes #0 = { "go-async-unsafe" }

; Async-preemption analysis is read-only and happens after final machine
; lowering. The pre-isel IR pipeline preserves the fail-closed attribute and
; does not encode safe/unsafe regions in IR.
; CHECK-LABEL: define goabiinternal i64 @ordinary_loop
; CHECK-SAME: #[[ATTR:[0-9]+]] gc "goallc"
; CHECK-LABEL: define goabiinternal ptr @ordinary_pointer_roundtrip
; CHECK-SAME: #[[ATTR]] gc "goallc"
; CHECK: ptrtoint ptr %p to i64
; CHECK: inttoptr i64 %address.lowered to ptr
; CHECK-LABEL: define goabiinternal void @ordinary_call
; CHECK-SAME: #[[ATTR]] gc "goallc"
; CHECK-LABEL: define goabiinternal i64 @ordinary_alloca
; CHECK-SAME: #[[ATTR]] gc "goallc"
; CHECK-LABEL: define goabiinternal i64 @ordinary_atomic
; CHECK-SAME: #[[ATTR]] gc "goallc"
; CHECK-LABEL: define goabiinternal <2 x i64> @ordinary_vector
; CHECK-SAME: #[[ATTR]] gc "goallc"
; CHECK: attributes #[[ATTR]] = { "go-async-unsafe" }
