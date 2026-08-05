target triple = "x86_64-unknown-linux-goobj"

@target = external global i8
@llvm.compiler.used = appending global [1 x ptr] [ptr @target], section "llvm.metadata"

declare void @llvm.sideeffect()

define internal goabiinternal void @callee()
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  call void @llvm.sideeffect(), !goobj.marker_reloc !0
  ret void
}

define goabiinternal void @caller()
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  call goabiinternal void @callee()
  ret void
}

!0 = !{ptr @target, i32 24, i64 96}
