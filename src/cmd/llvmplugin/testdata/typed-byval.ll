target triple = "x86_64-unknown-linux-goobj"

%pointer_pair = type { ptr, ptr }

; IR-LABEL: define goabiinternal void @pure_ssa_call_slot(
; A pure SSA value needs an ordinary source alloca for byval. The existing
; pointer-alloca machinery describes it at the statepoint; there is no
; byval-call-slot classifier or initializer analysis.
; IR: call goabiinternal token {{.*}} @llvm.experimental.gc.statepoint
; IR-SAME: ptr byval(ptr) align 8 %argument.home.address
; IR-SAME: i64 1095520067{{.*}}ptr %argument.home

; MIR-LABEL: name: pure_ssa_call_slot
; The ordinary register-argument home optimization may coalesce the source
; alloca with %argument's fixed home. The byval call still performs a normal
; store into its outgoing stack argument slot.
; MIR: stack: []
; MIR: ADJCALLSTACKDOWN64 80
; MIR: MOV64mr %{{[0-9]+}}, 1, $noreg, 0, $noreg
; MIR: STATEPOINT {{.*}}@consume_pointer

; IR-LABEL: define goabiinternal ptr @incoming_pointer_home(
; A pointer value loaded from the incoming byval home follows the generic
; statepoint spill path. The home is not reused as a spill for that value.
; IR: %value = load ptr, ptr %value.home, align 8
; IR-NOT: "deopt"(
; IR: "gc-live"(ptr %value)

; MIR-LABEL: name: incoming_pointer_home
; MIR: fixedStack:
; MIR: - { id: [[POINTER_OBJECT_HOME:[0-9]+]], type: default, offset: 0, size: 8
; MIR: stack:
; MIR-NEXT: - { id: [[POINTER_SPILL:[0-9]+]], name: '', type: default, offset: 0, size: 8
; MIR: MOV64mr %stack.[[POINTER_SPILL]]
; MIR: STATEPOINT
; MIR-SAME: 1, 8, %stack.[[POINTER_SPILL]], 0
; MIR: MOV64rm %stack.[[POINTER_SPILL]]

; IR-LABEL: define goabiinternal ptr @incoming_aggregate_home(
; IR-NOT: "deopt"(
; IR: "gc-live"(ptr %second)

; MIR-LABEL: name: incoming_aggregate_home
; MIR: stack:
; MIR-NEXT: - { id: [[AGGREGATE_SPILL:[0-9]+]], name: '', type: default, offset: 0, size: 8
; MIR: MOV64mr %stack.[[AGGREGATE_SPILL]]
; MIR: STATEPOINT
; MIR-SAME: 1, 8, %stack.[[AGGREGATE_SPILL]], 0
; MIR: MOV64rm %stack.[[AGGREGATE_SPILL]]

; IR-LABEL: define goabiinternal ptr @passed_to_mutator_incoming_pointer_home(
; Taking the address of a Go parameter and passing it on makes the incoming
; byval home address-observable. The object's pointer layout is retained, while
; the separately loaded pointer still follows the generic spill path.
; IR: %value = load ptr, ptr %value.home, align 8
; IR: elementtype(void (ptr)) @mutate
; IR: "deopt"({{.*}}ptr %value.home
; IR-SAME: "gc-live"(ptr %value)

; MIR-LABEL: name: passed_to_mutator_incoming_pointer_home
; MIR: stack:
; MIR-NEXT: - { id: 0, name: '', type: default, offset: 0, size: 8
; MIR: MOV64mr %stack.0
; MIR: STATEPOINT
; MIR-SAME: @mutate
; MIR-SAME: 1, 8, %stack.0, 0
; MIR: STATEPOINT
; MIR-SAME: @safepoint
; MIR-SAME: 1, 8, %stack.0, 0

; OBJVIEW-LABEL: "name": "incoming_pointer_home"
; OBJVIEW-NOT: "kind": "stack_objects"
; OBJVIEW-LABEL: "name": "incoming_aggregate_home"
; OBJVIEW-NOT: "kind": "stack_objects"
; OBJVIEW-LABEL: "name": "passed_to_mutator_incoming_pointer_home"
; OBJVIEW: "kind": "stack_objects"
; OBJVIEW: "offset": 0
; OBJVIEW: "size": 8
; OBJVIEW: "ptr_bytes": 8

declare goabiinternal void @consume_pointer(
    i64, i64, i64, i64, i64, i64, i64, i64, i64,
    ptr byval(ptr) align 8)

declare goabiinternal void @safepoint()

declare goabiinternal void @mutate(ptr)

define goabiinternal void @pure_ssa_call_slot(ptr %argument) #0 gc "goallc" {
entry:
  %argument.home = alloca ptr, align 8
  store ptr %argument, ptr %argument.home, align 8
  call goabiinternal void @consume_pointer(
      i64 0, i64 1, i64 2, i64 3, i64 4, i64 5, i64 6, i64 7, i64 8,
      ptr byval(ptr) align 8 %argument.home)
  ret void
}

define goabiinternal ptr @incoming_pointer_home(
    i64 %a0, i64 %a1, i64 %a2, i64 %a3, i64 %a4, i64 %a5, i64 %a6,
    i64 %a7, i64 %a8, ptr byval(ptr) align 8 %value.home) #0 gc "goallc" {
entry:
  %value = load ptr, ptr %value.home, align 8
  call goabiinternal void @safepoint()
  ret ptr %value
}

define goabiinternal ptr @incoming_aggregate_home(
    i64 %a0, i64 %a1, i64 %a2, i64 %a3, i64 %a4, i64 %a5, i64 %a6,
    i64 %a7, i64 %a8, ptr byval(%pointer_pair) align 8 %value.home) #0 gc "goallc" {
entry:
  %second.home = getelementptr inbounds %pointer_pair, ptr %value.home, i32 0, i32 1
  %second = load ptr, ptr %second.home, align 8
  call goabiinternal void @safepoint()
  ret ptr %second
}

define goabiinternal ptr @passed_to_mutator_incoming_pointer_home(
    i64 %a0, i64 %a1, i64 %a2, i64 %a3, i64 %a4, i64 %a5, i64 %a6,
    i64 %a7, i64 %a8, ptr byval(ptr) align 8 %value.home) #0 gc "goallc" {
entry:
  %value = load ptr, ptr %value.home, align 8
  call goabiinternal void @mutate(ptr %value.home)
  call goabiinternal void @safepoint()
  ret ptr %value
}

attributes #0 = { "frame-pointer"="non-leaf" }
