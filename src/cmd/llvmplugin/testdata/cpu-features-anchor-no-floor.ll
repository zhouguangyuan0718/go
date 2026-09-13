; An explicit anchor outside FMV still needs baseline or ABI-floor coverage.
; CHECK: error: GoALLC CPU requirement x86.avx2 survives in function no_floor without the required target features

target triple = "x86_64-unknown-linux-goobj"
declare void @llvm.sideeffect()

define void @no_floor() {
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !2
  ret void
}

!goallc.cpu.config = !{!0}
!0 = !{!"goallc.cpu.v1", !"amd64", !"v1"}
!1 = !{!"x86.avx2"}
!2 = !{}
