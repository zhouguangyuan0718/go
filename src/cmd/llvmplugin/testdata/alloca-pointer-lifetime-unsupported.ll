target triple = "x86_64-unknown-linux-goobj"

declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture)
declare void @llvm.fake.use(...)
declare goabiinternal void @safepoint()
declare goabiinternal void @observe(ptr)

define goabiinternal void @locals_pointer_alloca_with_lifetime() "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca ptr, align 8
  call goabiinternal void @safepoint()
  call void @llvm.lifetime.start.p0(i64 8, ptr %slot)
  store ptr null, ptr %slot, align 8
  call goabiinternal void @safepoint()
  call void (...) @llvm.fake.use(ptr %slot)
  call goabiinternal void @safepoint()
  ret void
}

define goabiinternal void @stack_object_alloca_with_lifetime() "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca ptr, align 8
  call goabiinternal void @safepoint()
  call void @llvm.lifetime.start.p0(i64 8, ptr %slot)
  store ptr null, ptr %slot, align 8
  call goabiinternal void @observe(ptr %slot)
  call goabiinternal void @safepoint()
  ret void
}

define goabiinternal void @loop_reinitialized_pointer_alloca(i1 %again) "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca ptr, align 8
  br label %loop

loop:
  call void @llvm.lifetime.start.p0(i64 8, ptr %slot)
  store ptr null, ptr %slot, align 8
  call goabiinternal void @safepoint()
  call void (...) @llvm.fake.use(ptr %slot)
  br i1 %again, label %loop, label %exit

exit:
  call goabiinternal void @safepoint()
  ret void
}
