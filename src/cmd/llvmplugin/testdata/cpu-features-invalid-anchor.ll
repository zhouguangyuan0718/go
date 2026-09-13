; Validate the marker contract before guard simplification can erase a
; malformed anchor together with an identity operation.
; CHECK: error: !goallc.cpu.require-anchor must mark a void llvm.sideeffect() call without arguments or operand bundles

target triple = "x86_64-unknown-linux-goobj"
@runtime.goallcCPUFeatures = external global i64

define <4 x i32> @bad_anchor(<4 x i32> %x) #0 {
  %identity = shl <4 x i32> %x, zeroinitializer, !goallc.cpu.requires !1, !goallc.cpu.require-anchor !2
  ret <4 x i32> %identity
}

attributes #0 = { "goallc.cpu.multiversion"="x86.avx2" }
!goallc.cpu.config = !{!0}
!0 = !{!"goallc.cpu.v1", !"amd64", !"v1"}
!1 = !{!"x86.avx2"}
!2 = !{}
