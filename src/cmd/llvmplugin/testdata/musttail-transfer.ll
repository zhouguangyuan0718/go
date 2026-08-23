target triple = "aarch64-unknown-linux-gnu"

; CHECK-LABEL: define goabi0 void @"abi0_tail<ABI0>"
; CHECK: musttail call goabiinternal void @target()
; CHECK-NEXT: ret void
; CHECK-NOT: llvm.experimental.gc.statepoint

declare goabiinternal void @target()

define goabi0 void @"abi0_tail<ABI0>"() "gc-leaf-function" "go-nosplit" gc "goallc" {
entry:
  musttail call goabiinternal void @target()
  ret void
}
