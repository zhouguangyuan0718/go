; Standard target features supply capabilities but never Go predicates.
target triple = "x86_64-unknown-linux-gnu"
@runtime.goallcCPUFeatures = external global i64
@runtime.x86HasAVX2 = external global i8

define <8 x i32> @implied(<8 x i32> %x) #0 {
  %v = add <8 x i32> %x, %x, !goallc.cpu.requires !1
  ret <8 x i32> %v
}
; CHECK-LABEL: define internal <8 x i32> @"implied<goallc.fmv.baseline>"
; CHECK: add <8 x i32>

define <8 x i32> @cpu_default(<8 x i32> %x) #1 {
  %v = add <8 x i32> %x, %x, !goallc.cpu.requires !1
  ret <8 x i32> %v
}
; CHECK-LABEL: define internal <8 x i32> @"cpu_default<goallc.fmv.baseline>"
; CHECK: add <8 x i32>

define i8 @predicate() #0 {
  %g = load i8, ptr @runtime.x86HasAVX2, !goallc.cpu.guard !1
  ret i8 %g
}
; CHECK-LABEL: define internal i8 @"predicate<goallc.fmv.baseline>"
; CHECK: ret i8 0
; CHECK-LABEL: define internal i8 @"predicate<goallc.fmv.avx2>"
; CHECK: ret i8 1

define <8 x i32> @reenable(<8 x i32> %x) #2 {
  %g = load i8, ptr @runtime.x86HasAVX2, !goallc.cpu.guard !1
  %enabled = icmp ne i8 %g, 0
  br i1 %enabled, label %fast, label %slow
fast:
  %v = add <8 x i32> %x, %x, !goallc.cpu.requires !1
  ret <8 x i32> %v
slow:
  ret <8 x i32> %x
}
; CHECK-LABEL: define internal <8 x i32> @"reenable<goallc.fmv.baseline>"
; CHECK-NOT: add <8 x i32>
; CHECK: ret <8 x i32> %x
; CHECK-LABEL: define internal <8 x i32> @"reenable<goallc.fmv.avx2>"
; CHECK: add <8 x i32>

; LLVM supplies AVX implicitly and honors the final positive override.
attributes #0 = { "target-features"="-avx2,+avx2" "goallc.cpu.multiversion"="x86.avx2" }
attributes #1 = { "target-cpu"="x86-64-v3" "goallc.cpu.multiversion"="x86.avx2" }
attributes #2 = { "target-features"="+avx,+avx2,-avx2" "goallc.cpu.multiversion"="x86.avx2" }
!goallc.cpu.config = !{!0}
!0 = !{!"goallc.cpu.v1", !"amd64", !"v1"}
!1 = !{!"x86.avx2"}
