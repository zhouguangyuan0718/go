target triple = "x86_64-unknown-linux-goobj"

; A GC-leaf contract describes calls to the entry point. An unannotated
; implementation helper is still conservatively rewritten inside the body.
; CHECK-LABEL: define goabiinternal ptr @leaf_entry(
; CHECK: call goabiinternal token {{.*}}@llvm.experimental.gc.statepoint
; CHECK-SAME: @implementation_helper
; CHECK-SAME: "gc-live"(ptr %p)
; CHECK: [[LIVE:%[^ ]+]] = call coldcc ptr @llvm.experimental.gc.relocate
; CHECK: ret ptr [[LIVE]]
; CHECK-LABEL: define goabiinternal ptr @caller(
; CHECK-NOT: @llvm.experimental.gc.statepoint
; CHECK: call goabiinternal ptr @leaf_entry(ptr %p)
; CHECK-NOT: @llvm.experimental.gc.statepoint
; CHECK: ret ptr %p

declare goabiinternal void @implementation_helper()

define goabiinternal ptr @leaf_entry(ptr %p) "gc-leaf-function" gc "goallc" {
entry:
  call goabiinternal void @implementation_helper()
  ret ptr %p
}

define goabiinternal ptr @caller(ptr %p) gc "goallc" {
entry:
  %unused = call goabiinternal ptr @leaf_entry(ptr %p)
  ret ptr %p
}
