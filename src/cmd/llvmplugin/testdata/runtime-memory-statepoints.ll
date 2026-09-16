target triple = "x86_64-unknown-linux-goobj"

; Logical memory effects must not imply GC leaf. Both a read-only hash and a
; memory-free conversion can split the Go stack; live pointers must relocate.
declare goabiinternal i64 @runtime.memhash(ptr readonly captures(none), i64, i64) nounwind willreturn memory(read)
declare goabiinternal i32 @runtime.fint32to32(i32) nounwind willreturn memory(none)

; CHECK-LABEL: define goabiinternal i64 @hash_caller(
; CHECK: call goabiinternal token {{.*}}@llvm.experimental.gc.statepoint
; CHECK-SAME: @runtime.memhash
; CHECK-SAME: "gc-live"(
; CHECK: call coldcc ptr @llvm.experimental.gc.relocate
; CHECK: load i64, ptr %{{.*relocated.*}}
define goabiinternal i64 @hash_caller(ptr %data, ptr %live, i64 %seed) gc "goallc" {
entry:
  %hash = call goabiinternal i64 @runtime.memhash(ptr %data, i64 %seed, i64 8)
  %v = load i64, ptr %live
  %r = add i64 %hash, %v
  ret i64 %r
}

; CHECK-LABEL: define goabiinternal i32 @numeric_caller(
; CHECK: call goabiinternal token {{.*}}@llvm.experimental.gc.statepoint
; CHECK-SAME: @runtime.fint32to32
; CHECK-SAME: "gc-live"(ptr %live)
; CHECK: [[LIVE:%.*]] = call coldcc ptr @llvm.experimental.gc.relocate
; CHECK: load i32, ptr [[LIVE]]
define goabiinternal i32 @numeric_caller(i32 %x, ptr %live) gc "goallc" {
entry:
  %bits = call goabiinternal i32 @runtime.fint32to32(i32 %x)
  %v = load i32, ptr %live
  %r = add i32 %bits, %v
  ret i32 %r
}

; The same contract on a definition must not suppress its internal safepoints.
; CHECK-LABEL: define goabiinternal i32 @numeric_definition(
; CHECK: call goabiinternal token {{.*}}@llvm.experimental.gc.statepoint
; CHECK-SAME: @runtime.fint32to32
; CHECK: call i32 @llvm.experimental.gc.result
; CHECK: ret i32
define goabiinternal i32 @numeric_definition(i32 %x) nounwind willreturn memory(none) gc "goallc" {
entry:
  %bits = call goabiinternal i32 @runtime.fint32to32(i32 %x)
  ret i32 %bits
}
