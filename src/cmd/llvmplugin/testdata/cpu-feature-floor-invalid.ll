; A later negative feature overrides both the CPU default and an earlier +avx2.
; CHECK: error: GoALLC CPU requirement x86.avx2 survives in function bad<goallc.fmv.baseline> without the required target features

target triple = "x86_64-unknown-linux-gnu"

@runtime.goallcCPUFeatures = external global i64

define <8 x i32> @bad(<8 x i32> %x) #0 {
entry:
  %v = add <8 x i32> %x, %x, !goallc.cpu.requires !1
  ret <8 x i32> %v
}

attributes #0 = { "target-cpu"="x86-64-v3" "target-features"="+avx2,-avx2" "goallc.cpu.multiversion"="x86.fma" }

!goallc.cpu.config = !{!0}
!0 = !{!"goallc.cpu.v1", !"amd64", !"v1"}
!1 = !{!"x86.avx2"}
