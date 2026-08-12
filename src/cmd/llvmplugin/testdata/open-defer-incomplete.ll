target triple = "x86_64-unknown-linux-goobj"

; ERROR: GoALLC open-coded defer function is missing frame state metadata

declare goabiinternal void @safepoint()

define goabiinternal void @open_defer_missing_slots() #0 gc "goallc" {
entry:
  %bits = alloca i8, align 1, !goallc.open_defer_bits !0
  store volatile i8 0, ptr %bits, align 1
  call goabiinternal void @safepoint()
  ret void
}

attributes #0 = { "go-stack-growth-statepoint" }

!0 = !{}
