target triple = "x86_64-unknown-linux-goobj"

declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture)
declare void @llvm.fake.use(...)
declare goabiinternal void @safepoint()

define goabiinternal void @path_dependent_pointer_alloca_lifetime(i1 %start) "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca ptr, align 8
  br i1 %start, label %live, label %join

live:
  call void @llvm.lifetime.start.p0(i64 8, ptr %slot)
  br label %join

join:
  call goabiinternal void @safepoint()
  call void (...) @llvm.fake.use(ptr %slot)
  ret void
}
