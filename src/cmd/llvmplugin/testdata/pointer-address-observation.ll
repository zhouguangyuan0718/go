target triple = "aarch64-apple-darwin-goobj"

declare i64 @__goallc$pointer.address(ptr)
declare goabiinternal void @callee()

define goabiinternal i1 @observe(ptr %pointer)
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  %before = call i64 @__goallc$pointer.address(ptr %pointer)
  call goabiinternal void @callee()
  %after = call i64 @__goallc$pointer.address(ptr %pointer)
  %same = icmp eq i64 %before, %after
  ret i1 %same
}
