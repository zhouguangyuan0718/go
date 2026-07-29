target triple = "aarch64-apple-darwin-goobj"

define goabiinternal ptr @aarch64_pointer_and_code_live(
    ptr %callee, ptr %pointer) #0 gc "goallc" {
entry:
  call goabiinternal void %callee()
  %value = load i8, ptr %pointer, align 1
  %used = icmp ne i8 %value, 0
  %result = select i1 %used, ptr %pointer, ptr %pointer
  ret ptr %result
}

attributes #0 = { "frame-pointer"="non-leaf" "go-stack-growth-statepoint" }
