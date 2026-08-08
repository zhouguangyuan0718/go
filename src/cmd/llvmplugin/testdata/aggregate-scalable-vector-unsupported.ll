target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @safepoint()

define goabiinternal ptr @scalable_vector(ptr %source) "go-stack-growth-statepoint" gc "goallc" {
entry:
  %value = load <vscale x 2 x ptr>, ptr %source, align 8
  call goabiinternal void @safepoint()
  %result = extractelement <vscale x 2 x ptr> %value, i32 0
  ret ptr %result
}
