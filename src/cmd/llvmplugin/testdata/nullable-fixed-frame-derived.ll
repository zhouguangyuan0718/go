target triple = "x86_64-unknown-linux-goobj"

; IR-LABEL: define goabiinternal i8 @nullable_fixed_frame_derived(
; IR: %base = phi ptr [ null, %nil ], [ %frame, %present ]
; IR-NOT: phi ptr [ inttoptr (i64 2240 to ptr)
; IR: @llvm.experimental.gc.statepoint{{.*}}@safepoint
; IR-SAME: "gc-live"(ptr %base)
; IR: %base.relocated{{.*}} = call {{.*}}ptr @llvm.experimental.gc.relocate
; IR: %field.derived.remat{{.*}} = getelementptr i8, ptr %base.relocated{{.*}}, i64 2240

; MIR-LABEL: name: nullable_fixed_frame_derived
; MIR: stack:
; MIR: name: frame{{.*}}size: 2248
; MIR-NOT: MOV{{.*}}2240
; MIR: STATEPOINT{{.*}}@safepoint{{.*}}1, 8, %stack.1
; MIR: MOV8rm{{.*}}2240

; MIR-AARCH64-LABEL: name: nullable_fixed_frame_derived
; MIR-AARCH64: stack:
; MIR-AARCH64: name: frame{{.*}}size: 2248
; MIR-AARCH64-NOT: MOVi32imm 2240
; MIR-AARCH64: STATEPOINT{{.*}}@safepoint{{.*}}1, 8, %stack.1
; MIR-AARCH64: LDRBBui{{.*}}2240

; OBJVIEW-LABEL: TEXT nullable_fixed_frame_derived(SB)
; OBJVIEW: size=78 args=8 locals=2256
; OBJVIEW: FUNCDATA_LocalsPointerMaps count=2 bits=282
; OBJVIEW-SAME: map[0]={{0+}}
; OBJVIEW-SAME: map[1]=1{{0+}}
; OBJVIEW: CALL {{.*}}R_CALL:safepoint
; OBJVIEW-SAME: PCDATA_StackMapIndex=1
; OBJVIEW-SAME: LocalsPointerMaps=1{{0+}}
; OBJVIEW: MOVQ 0(SP), AX
; OBJVIEW-NEXT: {{.*}}MOVZX 0x8c0(AX), AX

declare goabiinternal void @safepoint()

; Match SROA's optimized form for a field address made through a pointer that
; selects either nil or a fixed stack object. The integer 2240 must remain an
; address offset, never an independently tracked Go pointer.
define goabiinternal i8 @nullable_fixed_frame_derived(i1 %use_frame)
    gc "goallc" {
entry:
  %frame = alloca [281 x i64], align 8
  br i1 %use_frame, label %present, label %nil

nil:
  br label %merge

present:
  %frame.field = getelementptr i8, ptr %frame, i64 2240
  br label %merge

merge:
  %field = phi ptr [ inttoptr (i64 2240 to ptr), %nil ],
                   [ %frame.field, %present ]
  %base = phi ptr [ null, %nil ], [ %frame, %present ]
  call goabiinternal void @safepoint()
  %value = load i8, ptr %field, align 1
  ret i8 %value
}
