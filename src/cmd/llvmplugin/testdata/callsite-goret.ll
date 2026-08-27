target triple = "x86_64-unknown-linux-goobj"

%slice = type { ptr, i64, i64 }

; OBJVIEW-LABEL: TEXT call_slice_abi0(SB)
; OBJVIEW-NOT: FUNCDATA_StackObjects
; OBJVIEW: R_CALL{{(ARM64)?}}:copySlice
; OBJVIEW-NEXT: {{.*}}ordinary safepoint{{.*}}LocalsPointerMaps={{0+}}
; OBJVIEW-NOT: FUNCDATA_StackObjects
;
; OBJVIEW-LABEL: TEXT live_value_call_slice_abi0(SB)
; OBJVIEW: R_CALL{{(ARM64)?}}:copySlice
; OBJVIEW-NEXT: {{.*}}ordinary safepoint{{.*}}LocalsPointerMaps={{0+}}
; OBJVIEW: R_CALL{{(ARM64)?}}:safepoint
; OBJVIEW-NEXT: {{.*}}ordinary safepoint{{.*}}LocalsPointerMaps={{[01]*1[01]*}}
;
; OBJVIEW-LABEL: TEXT observed_call_slice_abi0(SB)
; OBJVIEW: FUNCDATA_StackObjects independent object[0]
; OBJVIEW: R_CALL{{(ARM64)?}}:copySlice
; OBJVIEW-NEXT: {{.*}}ordinary safepoint{{.*}}LocalsPointerMaps={{0+}}
; OBJVIEW: R_CALL{{(ARM64)?}}:safepoint
; OBJVIEW-NEXT: {{.*}}ordinary safepoint{{.*}}LocalsPointerMaps={{[01]*1[01]*}}

; IR-LABEL: define goabiinternal %slice @call_slice_abi0(
; The ABI0 call defines the complete result carrier. Its pre-call contents are
; not a caller root and require neither zeroing nor pointer-map metadata.
; IR: %argument = alloca %slice, align 8
; IR: %result = alloca %slice, align 8
; IR-NOT: llvm.stackcoloring.no_merge
; IR-NOT: llvm.memset.inline
; IR: call goabi0 token {{.*}} @llvm.experimental.gc.statepoint
; IR-SAME: ptr byval(%slice) align 8 %argument
; IR-SAME: ptr goret(%slice) align 8 "goretindex"="0" %result
; IR-SAME: i32 0, i32 0)
; IR-NOT: "deopt"(
; IR-NOT: "gc-live"(

; MIR-LABEL: name: call_slice_abi0
; The slice argument is forwarded into the outgoing ABI0 frame. Only the
; statepoint goret lowering's post-call result carrier remains a local object.
; MIR: stack:
; MIR-NOT: name: argument
; MIR: name: result
; MIR: entry_values:
; MIR: STATEPOINT {{.*}}@"copySlice<ABI0>"
; The target copies the ABI0 output frame into the result FrameIndex before
; ending the call sequence, and the continuation reloads that FrameIndex.
; No result virtual register crosses the statepoint continuation boundary.
; MIR: {{(MOV64mr|STR(Q|X)ui)}} {{.*}}%stack.{{[0-9]+}}.result
; MIR: ADJCALLSTACKUP
; MIR: bb.1.entry.statepoint.cont:
; MIR: {{(MOV64rm|LDRXui)}} %stack.{{[0-9]+}}.result

declare goabi0 void @"copySlice<ABI0>"(
    ptr byval(%slice) align 8,
    ptr goret(%slice) align 8 "goretindex"="0")

define goabiinternal %slice @call_slice_abi0(%slice %value) #0 gc "goallc" {
entry:
  %argument = alloca %slice, align 8
  %result = alloca %slice, align 8
  %base = extractvalue %slice %value, 0
  store ptr %base, ptr %argument, align 8
  %len.address = getelementptr inbounds i8, ptr %argument, i64 8
  %len = extractvalue %slice %value, 1
  store i64 %len, ptr %len.address, align 8
  %cap.address = getelementptr inbounds i8, ptr %argument, i64 16
  %cap = extractvalue %slice %value, 2
  store i64 %cap, ptr %cap.address, align 8
  call void @llvm.lifetime.start.p0(ptr %result)
  call goabi0 void @"copySlice<ABI0>"(
      ptr byval(%slice) align 8 %argument,
      ptr goret(%slice) align 8 "goretindex"="0" %result)
  %result.base = load ptr, ptr %result, align 8
  %result.len.address = getelementptr inbounds i8, ptr %result, i64 8
  %result.len = load i64, ptr %result.len.address, align 8
  %result.cap.address = getelementptr inbounds i8, ptr %result, i64 16
  %result.cap = load i64, ptr %result.cap.address, align 8
  %ret.base = insertvalue %slice poison, ptr %result.base, 0
  %ret.len = insertvalue %slice %ret.base, i64 %result.len, 1
  %ret.cap = insertvalue %slice %ret.len, i64 %result.cap, 2
  ret %slice %ret.cap
}

; IR-LABEL: define goabiinternal %slice @live_value_call_slice_abi0(
; The carrier address is dead at its defining call, but a pointer loaded from
; the returned slice is ordinary SSA data and remains explicitly tracked when
; it crosses a later statepoint.
; IR: %live.result = alloca %slice, align 8
; IR: call goabi0 token {{.*}} @llvm.experimental.gc.statepoint
; IR-SAME: ptr goret(%slice) align 8 "goretindex"="0" %live.result
; IR-SAME: i32 0, i32 0)
; IR: %live.result.base = load ptr, ptr %live.result
; IR: call goabiinternal token {{.*}} @llvm.experimental.gc.statepoint
; IR-SAME: ptr elementtype(void ()) @safepoint
; IR-SAME: "gc-live"(ptr %live.result.base)
; IR: call coldcc ptr @llvm.experimental.gc.relocate

; MIR-LABEL: name: live_value_call_slice_abi0
; MIR: STATEPOINT {{.*}}@"copySlice<ABI0>"
; MIR: STATEPOINT {{.*}}@safepoint

declare goabiinternal void @safepoint()

define goabiinternal %slice @live_value_call_slice_abi0(%slice %value) #0 gc "goallc" {
entry:
  %live.argument = alloca %slice, align 8
  %live.result = alloca %slice, align 8
  %base = extractvalue %slice %value, 0
  store ptr %base, ptr %live.argument, align 8
  %len.address = getelementptr inbounds i8, ptr %live.argument, i64 8
  %len = extractvalue %slice %value, 1
  store i64 %len, ptr %len.address, align 8
  %cap.address = getelementptr inbounds i8, ptr %live.argument, i64 16
  %cap = extractvalue %slice %value, 2
  store i64 %cap, ptr %cap.address, align 8
  call void @llvm.lifetime.start.p0(ptr %live.result)
  call goabi0 void @"copySlice<ABI0>"(
      ptr byval(%slice) align 8 %live.argument,
      ptr goret(%slice) align 8 "goretindex"="0" %live.result)
  %live.result.base = load ptr, ptr %live.result, align 8
  %live.result.len.address = getelementptr inbounds i8, ptr %live.result, i64 8
  %live.result.len = load i64, ptr %live.result.len.address, align 8
  %live.result.cap.address = getelementptr inbounds i8, ptr %live.result, i64 16
  %live.result.cap = load i64, ptr %live.result.cap.address, align 8
  call goabiinternal void @safepoint()
  %ret.base = insertvalue %slice poison, ptr %live.result.base, 0
  %ret.len = insertvalue %slice %ret.base, i64 %live.result.len, 1
  %ret.cap = insertvalue %slice %ret.len, i64 %live.result.cap, 2
  ret %slice %ret.cap
}

; IR-LABEL: define goabiinternal %slice @observed_call_slice_abi0(
; An independently observable result address is not a pure goret carrier.
; Preserve the fixed-frame root and its rematerialization across both the
; defining call (contents not live yet) and a later safepoint (contents live).
; IR: %observed.result = alloca %slice, align 8, !llvm.stackcoloring.no_merge
; IR: call goabi0 token {{.*}} @llvm.experimental.gc.statepoint
; IR-SAME: ptr goret(%slice) align 8 "goretindex"="0" %observed.result
; IR-SAME: "deopt"({{.*}}ptr %observed.result, i64 0, i64 24, i64 8, i64 8, i64 0,
; IR-SAME: "gc-live"(ptr %observed.result)
; IR: call goabiinternal token {{.*}} @llvm.experimental.gc.statepoint
; IR-SAME: ptr elementtype(void ()) @safepoint
; IR-SAME: "deopt"({{.*}}ptr %observed.result, i64 0, i64 24, i64 8, i64 8, i64 1,
; IR-SAME: "gc-live"(ptr %observed.result)

declare goabiinternal void @observe(ptr)

define goabiinternal %slice @observed_call_slice_abi0(%slice %value) #0 gc "goallc" {
entry:
  %observed.argument = alloca %slice, align 8
  %observed.result = alloca %slice, align 8
  %base = extractvalue %slice %value, 0
  store ptr %base, ptr %observed.argument, align 8
  %len.address = getelementptr inbounds i8, ptr %observed.argument, i64 8
  %len = extractvalue %slice %value, 1
  store i64 %len, ptr %len.address, align 8
  %cap.address = getelementptr inbounds i8, ptr %observed.argument, i64 16
  %cap = extractvalue %slice %value, 2
  store i64 %cap, ptr %cap.address, align 8
  call void @llvm.lifetime.start.p0(ptr %observed.result)
  call goabi0 void @"copySlice<ABI0>"(
      ptr byval(%slice) align 8 %observed.argument,
      ptr goret(%slice) align 8 "goretindex"="0" %observed.result)
  call goabiinternal void @safepoint()
  call goabiinternal void @observe(ptr %observed.result)
  %result.base = load ptr, ptr %observed.result, align 8
  %result.len.address = getelementptr inbounds i8, ptr %observed.result, i64 8
  %result.len = load i64, ptr %result.len.address, align 8
  %result.cap.address = getelementptr inbounds i8, ptr %observed.result, i64 16
  %result.cap = load i64, ptr %result.cap.address, align 8
  %ret.base = insertvalue %slice poison, ptr %result.base, 0
  %ret.len = insertvalue %slice %ret.base, i64 %result.len, 1
  %ret.cap = insertvalue %slice %ret.len, i64 %result.cap, 2
  ret %slice %ret.cap
}

declare void @llvm.lifetime.start.p0(ptr nocapture) #1

attributes #0 = { "frame-pointer"="non-leaf" }
attributes #1 = { mustprogress nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
