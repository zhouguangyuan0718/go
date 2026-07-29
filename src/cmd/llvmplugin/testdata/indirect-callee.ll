target triple = "x86_64-unknown-linux-goobj"

define goabiinternal void @indirect_callee(ptr %callee, i64 %arg) "go-stack-growth-statepoint" gc "goallc" {
entry:
  call goabiinternal void %callee(i64 %arg)
  ret void
}
