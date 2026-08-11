target triple = "aarch64-apple-darwin-goobj"

declare i64 @llvm.go.pointer.address.i64.p0(ptr)
declare goabiinternal void @callee()

define goabiinternal i1 @observe(ptr %pointer)
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  %before = call i64 @llvm.go.pointer.address.i64.p0(ptr %pointer)
  call goabiinternal void @callee()
  %after = call i64 @llvm.go.pointer.address.i64.p0(ptr %pointer)
  %same = icmp eq i64 %before, %after
  ret i1 %same
}
