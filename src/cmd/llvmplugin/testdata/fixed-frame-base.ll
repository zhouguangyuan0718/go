target triple = "aarch64-apple-darwin-goobj"

%result_storage = type { ptr, i64, ptr }
%quad_pointers = type { ptr, ptr, ptr, ptr }

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

; A store selected between distinct pointer slots cannot kill either slot on
; every path. Both surrounding statepoints therefore retain the byval
; contents. The optimized encoding/json/v2 reproducer has this shape.
; IR-LABEL: define goabiinternal void @fixed_frame_byval_phi_offsets(
; IR-COUNT-2: "deopt"({{.*}}ptr %input{{.*}}i64 1{{.*}}i64 4{{.*}}i64 15{{.*}}"gc-live"({{.*}}ptr %input)
; IR-NOT: ; (%input, %input)

; An offset which cannot be enumerated is also conservative: the store has no
; must-def pointer slots, rather than rejecting the function.
; IR-LABEL: define goabiinternal void @fixed_frame_byval_unknown_offset(
; IR-COUNT-2: "deopt"({{.*}}ptr %input{{.*}}i64 1{{.*}}i64 4{{.*}}i64 15{{.*}}"gc-live"({{.*}}ptr %input)
; IR-NOT: ; (%input, %input)

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
declare void @observe_pointer(ptr) #1

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

define goabiinternal void @fixed_frame_byval_phi_offsets(
    ptr byval(%quad_pointers) align 8 %input, i1 %condition,
    ptr %replacement) #0 gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  br i1 %condition, label %left, label %right

left:
  br label %merge

right:
  br label %merge

merge:
  %offset = phi i64 [ 0, %left ], [ 16, %right ]
  %address = getelementptr i8, ptr %input, i64 %offset
  store ptr %replacement, ptr %address, align 8
  call goabiinternal void @safepoint()
  %value = load ptr, ptr %input, align 8
  call void @observe_pointer(ptr %value)
  ret void
}

define goabiinternal void @fixed_frame_byval_unknown_offset(
    ptr byval(%quad_pointers) align 8 %input, i64 %offset,
    ptr %replacement) #0 gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  %address = getelementptr i8, ptr %input, i64 %offset
  store ptr %replacement, ptr %address, align 8
  call goabiinternal void @safepoint()
  %value = load ptr, ptr %input, align 8
  call void @observe_pointer(ptr %value)
  ret void
}

attributes #0 = { "frame-pointer"="non-leaf" "go_results_tuple" }
attributes #1 = { "gc-leaf-function" }
