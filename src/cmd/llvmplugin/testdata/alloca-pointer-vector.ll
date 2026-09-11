target triple = "x86_64-unknown-linux-goobj"
target datalayout = "e-m:e-p270:32:32-p271:32:32-p272:64:64-i64:64-i128:128-f80:128-v128:64:64-v256:64:64-v512:64:64-n8:16:32:64-S64"

; IR-LABEL: define goabiinternal ptr @pointer_vector_alloca(
; IR: %slot = alloca <5 x ptr>, align 8
; IR: "deopt"({{.*}}i64 1195461697, ptr %slot, i64 65, i64 31, i64 1095519299, i64 5)
; IR-SAME: "gc-live"(ptr %slot)

; IR-LABEL: define goabiinternal ptr @partially_initialized_pointer_vector(
; IR: %slot = alloca <5 x ptr>, align 8
; IR-NEXT: call void @llvm.memset.inline.p0.i64(ptr align 8 %slot, i8 0, i64 64, i1 false)

; IR-LABEL: define goabiinternal ptr @aggregate_with_non_pointer_vector(
; IR: %slot = alloca %mixed, align 8

; OBJVIEW-LABEL: "name": "pointer_vector_alloca"
; OBJVIEW: "kind": "locals_pointer_maps"
; OBJVIEW: "num_bits": 8
; OBJVIEW: "index": 1
; OBJVIEW-NEXT: "set_bits": [
; OBJVIEW-NEXT: 0,
; OBJVIEW-NEXT: 1,
; OBJVIEW-NEXT: 2,
; OBJVIEW-NEXT: 3,
; OBJVIEW-NEXT: 4

declare goabiinternal void @safepoint()

define goabiinternal ptr @pointer_vector_alloca(
    ptr %p0, ptr %p1, ptr %p2, ptr %p3, ptr %p4) gc "goallc" {
entry:
  %slot = alloca <5 x ptr>, align 8
  %value.0 = insertelement <5 x ptr> poison, ptr %p0, i32 0
  %value.1 = insertelement <5 x ptr> %value.0, ptr %p1, i32 1
  %value.2 = insertelement <5 x ptr> %value.1, ptr %p2, i32 2
  %value.3 = insertelement <5 x ptr> %value.2, ptr %p3, i32 3
  %value.4 = insertelement <5 x ptr> %value.3, ptr %p4, i32 4
  store <5 x ptr> %value.4, ptr %slot, align 8
  call goabiinternal void @safepoint()
  %loaded = load <5 x ptr>, ptr %slot, align 8
  %result = extractelement <5 x ptr> %loaded, i32 4
  ret ptr %result
}

define goabiinternal ptr @partially_initialized_pointer_vector(
    ptr %p0, ptr %p1, ptr %p2, ptr %p3) gc "goallc" {
entry:
  %slot = alloca <5 x ptr>, align 8
  %value.0 = insertelement <5 x ptr> poison, ptr %p0, i32 0
  %value.1 = insertelement <5 x ptr> %value.0, ptr %p1, i32 1
  %value.2 = insertelement <5 x ptr> %value.1, ptr %p2, i32 2
  %value.3 = insertelement <5 x ptr> %value.2, ptr %p3, i32 3
  store <5 x ptr> %value.3, ptr %slot, align 8
  call goabiinternal void @safepoint()
  %loaded = load <5 x ptr>, ptr %slot, align 8
  %result = extractelement <5 x ptr> %loaded, i32 3
  ret ptr %result
}

%mixed = type { ptr, <4 x i32> }

define goabiinternal ptr @aggregate_with_non_pointer_vector(
    ptr %p) gc "goallc" {
entry:
  %slot = alloca %mixed, align 8
  %value.0 = insertvalue %mixed poison, ptr %p, 0
  %value.1 = insertvalue %mixed %value.0, <4 x i32> zeroinitializer, 1
  store %mixed %value.1, ptr %slot, align 8
  call goabiinternal void @safepoint()
  %loaded = load %mixed, ptr %slot, align 8
  %result = extractvalue %mixed %loaded, 0
  ret ptr %result
}
