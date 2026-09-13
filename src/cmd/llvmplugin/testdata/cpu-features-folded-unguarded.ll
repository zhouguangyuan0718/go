; An identity shift disappearing during local simplification must not erase
; the source operation's CPU contract. A requested stronger clone does not
; legalize this operation in the baseline clone.
; CHECK: error: GoALLC CPU requirement x86.avx2 survives in function folded_bad<goallc.fmv.baseline> without the required target features

target triple = "x86_64-unknown-linux-goobj"
@runtime.goallcCPUFeatures = external global i64
declare void @llvm.sideeffect()

define <4 x i32> @folded_bad(<4 x i32> %x) #0 {
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !2
  %identity = shl <4 x i32> %x, zeroinitializer
  ret <4 x i32> %identity
}

attributes #0 = { "goallc.cpu.multiversion"="x86.avx512" }
!goallc.cpu.config = !{!0}
!0 = !{!"goallc.cpu.v1", !"amd64", !"v1"}
!1 = !{!"x86.avx2"}
!2 = !{}
