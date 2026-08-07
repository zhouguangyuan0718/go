target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @callee()
declare goabiinternal ptr @make_pointer()
declare void @llvm.donothing()

; A named result used only by panic recovery is not ordinarily live through
; LLVM's explicit CFG. The marker makes the pointer live at calls after its
; definition, but not at the call that produces it.
define goabiinternal ptr @defer_result_live() #0 gc "goallc" {
entry:
  %result = call goabiinternal ptr @make_pointer()
  call void @llvm.donothing() [ "go.defer.result.live"(ptr %result) ]
  call goabiinternal void @callee()
  call goabiinternal void @callee()
  ret ptr null
}

attributes #0 = { "go-stack-growth-statepoint" }
