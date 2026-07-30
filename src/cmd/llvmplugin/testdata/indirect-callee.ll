target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @consume_pointer(ptr)

define goabiinternal void @indirect_callee(ptr %callee, i64 %arg) #0 gc "goallc" {
entry:
  call goabiinternal void %callee(i64 %arg)
  ret void
}

define goabiinternal void @memory_indirect_callee(ptr %callee_slot, i64 %arg) #0 gc "goallc" {
entry:
  %callee = load ptr, ptr %callee_slot, align 8
  call goabiinternal void %callee(i64 %arg)
  ret void
}

define goabiinternal void @call_only_pointer_argument(ptr %value) #0 gc "goallc" {
entry:
  call goabiinternal void @consume_pointer(ptr %value)
  ret void
}

attributes #0 = { "frame-pointer"="non-leaf" "go-stack-growth-statepoint" }
