target triple = "aarch64-apple-darwin-goobj"

%result_storage = type { ptr, i64, ptr }

; IR-LABEL: define goabiinternal {{.*}} @fixed_frame_goret_base(
; IR-NOT: @llvm.experimental.gc.relocate
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %result)
; IR: store %result_storage zeroinitializer, ptr %result, align 8
; IR: "deopt"({{.*}}ptr %result{{.*}}i64 1{{.*}}i64 3{{.*}}i64 5{{.*}}"gc-live"(ptr %result)
; IR-NOT: @llvm.experimental.gc.relocate

; IR-LABEL: define goabiinternal void @fixed_frame_byval_contents(
; IR: "deopt"({{.*}}ptr %input{{.*}}i64 1{{.*}}i64 3{{.*}}i64 5{{.*}}"gc-live"(ptr %input)
; IR: load ptr, ptr %input, align 8
; IR-NOT: "deopt"({{.*}}ptr %input

; MIR-LABEL: name: fixed_frame_goret_base
; MIR: fixedStack:
; MIR: type: default
; MIR: stack: []
; MIR: bb.0.entry:
; MIR: STATEPOINT{{.*}}2, 1, 0, %fixed-stack.0, 0, 2, 0, csr_aarch64_go,
; MIR: bb.1.entry.statepoint.cont:
; MIR: [[RESULT:%[0-9]+]]:gpr64 = COPY $xzr
; MIR: STRXui [[RESULT]], %fixed-stack.0, 0
; MIR: STATEPOINT{{.*}}2, 1, 0, %fixed-stack.0, 0, 2, 0, csr_aarch64_go,
; MIR: bb.2.entry.statepoint.cont.statepoint.cont:

; OBJVIEW-LABEL: "name": "fixed_frame_goret_base"
; OBJVIEW: "kind": "args_pointer_maps"
; OBJVIEW: "count": 2
; OBJVIEW: "num_bits": 3
; OBJVIEW: "set_bits": null
; OBJVIEW: "set_bits": [
; OBJVIEW-NEXT: 0,
; OBJVIEW-NEXT: 2
; OBJVIEW: "kind": "locals_pointer_maps"
; OBJVIEW: "count": 2
; OBJVIEW: "set_bits": null

; OBJVIEW-LABEL: "name": "fixed_frame_byval_contents"
; OBJVIEW: "kind": "args_pointer_maps"
; OBJVIEW: "count": 2
; OBJVIEW: "num_bits": 3
; OBJVIEW: "set_bits": [
; OBJVIEW-NEXT: 0,
; OBJVIEW-NEXT: 2
; OBJVIEW: "set_bits": null

declare goabiinternal void @safepoint()

define goabiinternal { i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64 } @fixed_frame_goret_base(
    ptr goret(%result_storage) align 8 "goretindex"="15" %result) #0 gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  store %result_storage zeroinitializer, ptr %result, align 8
  call goabiinternal void @safepoint()
  ret { i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64, i64 } zeroinitializer
}

define goabiinternal void @fixed_frame_byval_contents(
    ptr byval(%result_storage) align 8 %input) #0 gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  %value = load ptr, ptr %input, align 8
  call goabiinternal void @safepoint()
  ret void
}

attributes #0 = { "frame-pointer"="non-leaf" "go_results_tuple" }
