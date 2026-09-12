; A source-local requirement must survive folding to an argument or constant,
; then disappear after verification so it cannot block ordinary optimization.
; The producer before the guard has no CPU requirement of its own.

target triple = "x86_64-unknown-linux-goobj"
@runtime.goallcCPUFeatures = external global i64
@runtime.x86HasAVX2 = external global i8
@runtime.x86HasAVX512 = external global i8

declare void @llvm.sideeffect()

define <4 x i32> @folded_identity(<4 x i32> %x, <4 x i32> %y) #0 {
entry:
  %producer = add <4 x i32> %x, %y
  %flag = load i8, ptr @runtime.x86HasAVX2, !goallc.cpu.guard !1
  %enabled = icmp ne i8 %flag, 0
  br i1 %enabled, label %feature, label %fallback
feature:
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !3
  %identity = shl <4 x i32> %producer, zeroinitializer
  ret <4 x i32> %identity
fallback:
  ret <4 x i32> %producer
}

define <4 x i32> @folded_constant(<4 x i32> %x) #0 {
entry:
  %flag = load i8, ptr @runtime.x86HasAVX2, !goallc.cpu.guard !1
  %enabled = icmp ne i8 %flag, 0
  br i1 %enabled, label %feature, label %fallback
feature:
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !3
  %zero = shl <4 x i32> zeroinitializer, <i32 1, i32 7, i32 15, i32 31>
  ret <4 x i32> %zero
fallback:
  ret <4 x i32> %x
}

; A strong guard supplies AVX2 instructions without forcing the independently
; controlled AVX2 boolean true. Every result folds, but each live operation
; still has to pass requirement verification.
define i32 @folded_mixed() #1 {
entry:
  %high = load i8, ptr @runtime.x86HasAVX512, !goallc.cpu.guard !2
  %high.enabled = icmp ne i8 %high, 0
  br i1 %high.enabled, label %high.path, label %low.path
high.path:
  %low = load i8, ptr @runtime.x86HasAVX2, !goallc.cpu.guard !1
  %low.enabled = icmp ne i8 %low, 0
  br i1 %low.enabled, label %both, label %only.high
both:
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !3
  %zero = shl <4 x i32> zeroinitializer, <i32 1, i32 7, i32 15, i32 31>
  %both.result = extractelement <4 x i32> %zero, i32 0
  ret i32 %both.result
only.high:
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !3
  %one = lshr <4 x i32> <i32 1, i32 1, i32 1, i32 1>, zeroinitializer
  %high.result = extractelement <4 x i32> %one, i32 0
  ret i32 %high.result
low.path:
  %other = load i8, ptr @runtime.x86HasAVX2, !goallc.cpu.guard !1
  %other.enabled = icmp ne i8 %other, 0
  br i1 %other.enabled, label %only.low, label %fallback
only.low:
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !3
  %two = shl <4 x i32> <i32 1, i32 1, i32 1, i32 1>, <i32 1, i32 1, i32 1, i32 1>
  %low.result = extractelement <4 x i32> %two, i32 0
  ret i32 %low.result
fallback:
  ret i32 3
}

; Non-FMV functions also consume anchors once their ABI floor proves them.
; Plain sideeffect calls, including one with a normal requires attachment,
; retain their semantics and must not be erased with the dedicated anchor.
define void @floor_anchors() #2 {
  call void @llvm.sideeffect()
  call void @llvm.sideeffect(), !goallc.cpu.requires !1
  call void @llvm.sideeffect(), !goallc.cpu.requires !1, !goallc.cpu.require-anchor !3
  ret void
}

attributes #0 = { "goallc.cpu.multiversion"="x86.avx2" }
attributes #1 = { "goallc.cpu.multiversion"="x86.avx2,x86.avx512" }
attributes #2 = { "goallc.cpu.feature-floor"="x86.avx2" }
!goallc.cpu.config = !{!0}
!0 = !{!"goallc.cpu.v1", !"amd64", !"v1"}
!1 = !{!"x86.avx2"}
!2 = !{!"x86.avx512"}
!3 = !{}

; CHECK-LABEL: define void @floor_anchors()
; CHECK: call void @llvm.sideeffect()
; CHECK-NEXT: call void @llvm.sideeffect(), !goallc.cpu.requires
; CHECK-NEXT: ret void

; CHECK-LABEL: define internal <4 x i32> @"folded_identity<goallc.fmv.baseline>"
; CHECK: %producer = add <4 x i32> %x, %y{{$}}
; CHECK-NOT: call void @llvm.sideeffect
; CHECK-NOT: shl
; CHECK: ret <4 x i32> %producer
; CHECK-LABEL: define internal <4 x i32> @"folded_identity<goallc.fmv.avx2>"
; CHECK: %producer = add <4 x i32> %x, %y{{$}}
; CHECK-NOT: call void @llvm.sideeffect
; CHECK-NOT: shl
; CHECK: ret <4 x i32> %producer

; CHECK-LABEL: define internal <4 x i32> @"folded_constant<goallc.fmv.baseline>"
; CHECK-NOT: call void @llvm.sideeffect
; CHECK: ret <4 x i32> %x
; CHECK-LABEL: define internal <4 x i32> @"folded_constant<goallc.fmv.avx2>"
; CHECK-NOT: call void @llvm.sideeffect
; CHECK-NOT: shl
; CHECK: ret <4 x i32> zeroinitializer

; CHECK-LABEL: define internal i32 @"folded_mixed<goallc.fmv.baseline>"
; CHECK-NOT: call void @llvm.sideeffect
; CHECK: ret i32 3
; CHECK-LABEL: define internal i32 @"folded_mixed<goallc.fmv.avx2>"
; CHECK-NOT: call void @llvm.sideeffect
; CHECK: ret i32 2
; CHECK-LABEL: define internal i32 @"folded_mixed<goallc.fmv.avx512>"
; CHECK-NOT: call void @llvm.sideeffect
; CHECK: ret i32 1
; CHECK-LABEL: define internal i32 @"folded_mixed<goallc.fmv.avx2-avx512>"
; CHECK-NOT: call void @llvm.sideeffect
; CHECK: ret i32 0
; CHECK-LABEL: define internal i32 @"folded_mixed<goallc.fmv.resolve>"
; CHECK: and i64 %features, 512
; CHECK: and i64 %features, 1024
; CHECK: and i64 %features, 1536
; CHECK-NOT: goallc.cpu.require-anchor
