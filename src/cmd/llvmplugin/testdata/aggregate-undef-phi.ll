target triple = "x86_64-unknown-linux-goobj"

%read_result = type { i64, { ptr, ptr } }

; IR-LABEL: define goabiinternal i64 @partial_phi_across_call(
; IR-NOT: %value.leaf.1.0
; IR-NOT: %value.leaf.1.1
; IR: @llvm.experimental.gc.statepoint{{.*}}@safepoint
; IR-NOT: "gc-live"

; MIR-LABEL: name: partial_phi_across_call
; MIR: stack: []
; MIR-NOT: %stack.
; MIR: STATEPOINT{{.*}}@safepoint{{.*}}csr_64_goabiinternal
; MIR-NOT: %stack.

; MIR-AARCH64-LABEL: name: partial_phi_across_call
; MIR-AARCH64: stack: []
; MIR-AARCH64-NOT: %stack.
; MIR-AARCH64: STATEPOINT{{.*}}@safepoint{{.*}}csr_aarch64_go
; MIR-AARCH64-NOT: %stack.

; OBJVIEW-LABEL: TEXT partial_phi_across_call(SB) llvm-ir
; OBJVIEW: FUNCDATA_LocalsPointerMaps count=1 bits=1 map[0]=0
; OBJVIEW: R_CALL:safepoint{{.*}}PCDATA_StackMapIndex=0{{.*}}LocalsPointerMaps=0

; An aggregate PHI can carry a live integer while every pointer leaf is still
; undef. Scalarization must not turn those uninitialized leaves into GC roots.
declare goabiinternal void @safepoint()

define goabiinternal i64 @partial_phi_across_call(i1 %choose, i64 %left_number, i64 %right_number) gc "goallc" {
entry:
  br i1 %choose, label %left, label %right

left:
  %left_value = insertvalue %read_result undef, i64 %left_number, 0
  br label %merge

right:
  %right_value = insertvalue %read_result undef, i64 %right_number, 0
  br label %merge

merge:
  %value = phi %read_result [ %left_value, %left ], [ %right_value, %right ]
  call goabiinternal void @safepoint()
  %number = extractvalue %read_result %value, 0
  ret i64 %number
}
