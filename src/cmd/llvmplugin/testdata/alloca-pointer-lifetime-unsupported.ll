target triple = "x86_64-unknown-linux-goobj"

declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture)
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture)
declare void @llvm.memset.inline.p0.i64(ptr nocapture writeonly, i8, i64, i1 immarg)
declare void @llvm.fake.use(...)
declare goabiinternal void @safepoint()
declare goabiinternal void @observe(ptr)
declare goabiinternal void @observe_slice({ ptr, i64, i64 })

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

define goabiinternal void @preinitialized_pointer_alloca() "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca [2 x ptr], align 8
  call void @llvm.lifetime.start.p0(i64 16, ptr %slot)
  call void @llvm.memset.inline.p0.i64(ptr align 8 %slot, i8 0, i64 16, i1 false)
  call goabiinternal void @safepoint()
  call void (...) @llvm.fake.use(ptr %slot)
  ret void
}

define goabiinternal void @phi_edge_pointer_alloca(
    i1 %use_stack, ptr %other) "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca ptr, align 8
  br i1 %use_stack, label %initialize, label %external

initialize:
  call void @llvm.lifetime.start.p0(i64 8, ptr %slot)
  store ptr null, ptr %slot, align 8
  br label %merge

external:
  br label %merge

merge:
  %selected = phi ptr [ %slot, %initialize ], [ %other, %external ]
  call goabiinternal void @safepoint()
  call void (...) @llvm.fake.use(ptr %selected)
  ret void
}

define goabiinternal void @hoisted_aggregate_pointer_alloca() "go-stack-growth-statepoint" gc "goallc" {
entry:
  %slot = alloca [2 x ptr], align 8
  %slice = insertvalue { ptr, i64, i64 } poison, ptr %slot, 0
  %slice.len = insertvalue { ptr, i64, i64 } %slice, i64 2, 1
  %slice.cap = insertvalue { ptr, i64, i64 } %slice.len, i64 2, 2
  call goabiinternal void @safepoint()
  call void @llvm.lifetime.start.p0(i64 16, ptr %slot)
  store ptr null, ptr %slot, align 8
  %second = getelementptr inbounds ptr, ptr %slot, i64 1
  store ptr null, ptr %second, align 8
  call goabiinternal void @observe_slice({ ptr, i64, i64 } %slice.cap)
  call void @llvm.lifetime.end.p0(i64 16, ptr %slot)
  call goabiinternal void @safepoint()
  ret void
}
