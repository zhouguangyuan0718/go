target triple = "x86_64-unknown-linux-goobj"

; IR-LABEL: define goabiinternal ptr @pair_across_call(
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %value.leaf.0)

; IR-LABEL: define goabiinternal ptr @triple_across_call(
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %value.leaf.1, ptr %value.leaf.2)

; IR-LABEL: define goabiinternal void @nested_across_call(
; IR: extractvalue %nested %value, 1, 1, 0

; IR-LABEL: define goabiinternal ptr @fixed_vector_across_call(
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(<2 x ptr> %value)
; IR: call coldcc <2 x ptr> @llvm.experimental.gc.relocate.v2p0

; IR-LABEL: define goabiinternal ptr @nested_fixed_vector_across_call(
; IR: %value.leaf.0 = extractvalue %vector_pair %value, 0
; IR: insertvalue %vector_pair {{.*}}<2 x ptr> %value.leaf.0.relocated, 0

; IR-LABEL: define goabiinternal ptr @insertvalue_across_call(
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %value.leaf.0)
; IR: insertvalue %pair {{.*}}ptr %value.leaf.0.relocated

; IR-LABEL: define goabiinternal ptr @inserted_pointer_also_live(
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %pointer)
; IR: %pointer.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: insertvalue %pair {{.*}}ptr %pointer.relocated
; IR: ret ptr %pointer.relocated

; IR-LABEL: define goabiinternal ptr @inserted_call_result_also_live(
; IR: [[CALL_RESULT:%.*]] = call ptr @llvm.experimental.gc.result.p0
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr [[CALL_RESULT]])
; IR: [[RELOCATED:%.*]] = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: insertvalue %pair {{.*}}ptr [[RELOCATED]]
; IR: ret ptr [[RELOCATED]]

; IR-LABEL: define goabiinternal ptr @inserted_pointer_in_call_argument(
; IR: @llvm.experimental.gc.statepoint{{.*}}@consume_slice{{.*}}{ ptr, i64, i64 } %slice{{.*}}"gc-live"(ptr %pointer, ptr %value.leaf.0)
; IR: %pointer.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: %value.leaf.0.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: insertvalue %pair {{.*}}ptr %value.leaf.0.relocated
; IR: ret ptr %pointer.relocated

; IR-LABEL: define goabiinternal ptr @inserted_leaf_in_call_argument(
; IR: @llvm.experimental.gc.statepoint{{.*}}@consume_pair{{.*}}"gc-live"(ptr %pointer, ptr %value.leaf.0)
; IR: %pointer.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: %value.leaf.0.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: insertvalue %pair {{.*}}ptr %value.leaf.0.relocated
; IR: ret ptr %pointer.relocated

; IR-LABEL: define goabiinternal ptr @inserted_call_result_leaf_in_call_argument(
; IR: @llvm.experimental.gc.statepoint{{.*}}@consume_pair{{.*}}"gc-live"(ptr %value.leaf.0)
; IR: %value.leaf.0.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: insertvalue %pair {{.*}}ptr %value.leaf.0.relocated
; IR: ret ptr %value.leaf.0.relocated

; IR-LABEL: define goabiinternal ptr @nested_direct_pointer_leaf_alias(
; IR: [[SOURCE_LEAF:%source\.leaf\.1\.0]] = extractvalue %container %source, 1, 0
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr [[SOURCE_LEAF]]
; IR: [[SOURCE_RELOCATED:%source\.leaf\.1\.0\.relocated[^ ]*]] = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: %copy.rebuilt = insertvalue %pair poison, ptr [[SOURCE_RELOCATED]], 0

; IR-LABEL: define goabiinternal ptr @inserted_pointer_used_by_derived(
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %value.leaf.0, ptr %pointer)
; IR: %value.leaf.0.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0
; IR: %pointer.relocated = call coldcc ptr @llvm.experimental.gc.relocate.p0

; IR-LABEL: define goabiinternal ptr @partial_insertvalue_across_call(
; IR: %partial.leaf.0 = extractvalue %reflect_value %partial, 0
; IR-NOT: %partial.leaf.1
; IR: @llvm.experimental.gc.statepoint{{.*}}"gc-live"(ptr %partial.leaf.0)

; IR-LABEL: define goabiinternal ptr @phi_across_call(
; IR: %value.leaf.0 = extractvalue %pair %value, 0

; IR-LABEL: define goabiinternal ptr @multiple_calls(
; IR: %value.leaf.0.relocated2

; IR-LABEL: define goabiinternal ptr @aggregate_call_result(
; IR: call %pair @llvm.experimental.gc.result.{{[^(]+}}

; IR-LABEL: define goabiinternal void @aggregate_current_call_argument(
; IR-NOT: extractvalue
; IR-NOT: "gc-live"
; IR: @llvm.experimental.gc.statepoint{{.*}}@consume_pair{{.*}}%pair %value
; IR: ret void

; IR-LABEL: define goabiinternal void @aggregate_load_store(
; IR: %value.leaf.0 = extractvalue %pair %value, 0

; IR-LABEL: define goabiinternal void @frozen_aggregate(
; IR: %value = freeze %pair poison

%pair = type { ptr, i64 }
%triple = type { i64, ptr, ptr }
%reflect_value = type { ptr, ptr, i64 }
%nested = type { i64, [2 x { ptr, i32 }] }
%vector_pair = type { <2 x ptr>, i64 }
%container = type { i64, %pair }

declare goabiinternal void @safepoint()
declare goabiinternal void @consume_pair(%pair)
declare goabiinternal %pair @make_pair(ptr, i64)
declare goabiinternal ptr @make_pointer()
declare goabiinternal void @consume_slice({ ptr, i64, i64 })
declare goabiinternal void @leaf_consume_pair(%pair) #0
declare goabiinternal void @leaf_consume_nested(%nested) #0
declare goabiinternal void @leaf_consume_vector_pair(%vector_pair) #0
declare goabiinternal void @leaf_consume_container(%container) #0

define goabiinternal ptr @pair_across_call(%pair %value) gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  %pointer = extractvalue %pair %value, 0
  ret ptr %pointer
}

define goabiinternal ptr @triple_across_call(%triple %value) gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  %first = extractvalue %triple %value, 1
  %second = extractvalue %triple %value, 2
  %same = icmp eq ptr %first, %second
  %result = select i1 %same, ptr %first, ptr %second
  ret ptr %result
}

define goabiinternal void @nested_across_call(%nested %value) gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_nested(%nested %value)
  ret void
}

define goabiinternal ptr @fixed_vector_across_call(ptr %source) gc "goallc" {
entry:
  %value = load <2 x ptr>, ptr %source, align 8
  call goabiinternal void @safepoint()
  %result = extractelement <2 x ptr> %value, i32 1
  ret ptr %result
}

define goabiinternal ptr @nested_fixed_vector_across_call(ptr %source, i64 %number) gc "goallc" {
entry:
  %vector.value = load <2 x ptr>, ptr %source, align 8
  %with_vector = insertvalue %vector_pair poison, <2 x ptr> %vector.value, 0
  %value = insertvalue %vector_pair %with_vector, i64 %number, 1
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_vector_pair(%vector_pair %value)
  %vector = extractvalue %vector_pair %value, 0
  %result = extractelement <2 x ptr> %vector, i32 1
  ret ptr %result
}

define goabiinternal ptr @insertvalue_across_call(ptr %pointer, i64 %number) gc "goallc" {
entry:
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_pair(%pair %value)
  %result = extractvalue %pair %value, 0
  ret ptr %result
}

; The aggregate leaf and the scalar originally inserted into it are both live
; after the safepoint. They retain distinct pre-rewrite SSA identities but must
; share one gc-live root and one relocated definition at this call.
define goabiinternal ptr @inserted_pointer_also_live(ptr %pointer, i64 %number) gc "goallc" {
entry:
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_pair(%pair %value)
  ret ptr %pointer
}

; Keep the same coalescing valid when the shared root is an earlier call result.
; Rewriting that earlier call replaces and erases its original Value identity.
define goabiinternal ptr @inserted_call_result_also_live(i64 %number) gc "goallc" {
entry:
  %pointer = call goabiinternal ptr @make_pointer()
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_pair(%pair %value)
  ret ptr %pointer
}

; A source that also supplies the current call's aggregate argument retains a
; distinct live-through identity. Go calling conventions may place the
; argument and live-through values in different machine locations.
define goabiinternal ptr @inserted_pointer_in_call_argument(ptr %pointer, i64 %number) gc "goallc" {
entry:
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  %slice_with_pointer = insertvalue { ptr, i64, i64 } poison, ptr %pointer, 0
  %slice_with_length = insertvalue { ptr, i64, i64 } %slice_with_pointer, i64 1, 1
  %slice = insertvalue { ptr, i64, i64 } %slice_with_length, i64 1, 2
  call goabiinternal void @consume_slice({ ptr, i64, i64 } %slice)
  call goabiinternal void @leaf_consume_pair(%pair %value)
  ret ptr %pointer
}

; An ordinary scalar source and its scalarized call-argument leaf retain
; distinct roots because the source may represent a separate ABI carrier.
define goabiinternal ptr @inserted_leaf_in_call_argument(ptr %pointer, i64 %number) gc "goallc" {
entry:
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  call goabiinternal void @consume_pair(%pair %value)
  call goabiinternal void @leaf_consume_pair(%pair %value)
  ret ptr %pointer
}

; A direct pointer call result has no aggregate ABI carrier identity. When its
; scalarized leaf supplies the current call argument, retain that leaf as the
; single root and share its relocation with the original result.
define goabiinternal ptr @inserted_call_result_leaf_in_call_argument(i64 %number) gc "goallc" {
entry:
  %pointer = call goabiinternal ptr @make_pointer()
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  call goabiinternal void @consume_pair(%pair %value)
  call goabiinternal void @leaf_consume_pair(%pair %value)
  ret ptr %pointer
}

; Scalarizing %source replaces and erases %source_pointer after %copy has
; recorded it as direct pointer provenance. The alias chain must be retargeted
; to %source's new leaf instead of retaining a dangling Value pointer which a
; later aggregate leaf can reuse.
define goabiinternal ptr @nested_direct_pointer_leaf_alias(%container %source, %pair %other, i64 %number) gc "goallc" {
entry:
  %source_pointer = extractvalue %container %source, 1, 0
  %copy_with_pointer = insertvalue %pair poison, ptr %source_pointer, 0
  %copy = insertvalue %pair %copy_with_pointer, i64 %number, 1
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_container(%container %source)
  call goabiinternal void @leaf_consume_pair(%pair %other)
  call goabiinternal void @leaf_consume_pair(%pair %copy)
  ret ptr %source_pointer
}

; A source added only as the relocation base for a derived address is not an
; independently live scalar root and cannot replace the aggregate leaf.
define goabiinternal ptr @inserted_pointer_used_by_derived(ptr %pointer, i64 %number) gc "goallc" {
entry:
  %with_pointer = insertvalue %pair zeroinitializer, ptr %pointer, 0
  %value = insertvalue %pair %with_pointer, i64 %number, 1
  %derived = getelementptr i8, ptr %pointer, i64 8
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_pair(%pair %value)
  %result = load ptr, ptr %derived
  ret ptr %result
}

; Scalarizing a partially initialized aggregate must not manufacture SSA
; pointer roots for poison leaves. Those leaves can be overwritten after the
; safepoint without ever being observed, as happens while reflect.Value is
; assembled for a later call.
define goabiinternal ptr @partial_insertvalue_across_call(ptr %pointer, i64 %number) gc "goallc" {
entry:
  %partial = insertvalue %reflect_value poison, ptr %pointer, 0
  call goabiinternal void @safepoint()
  %with_data = insertvalue %reflect_value %partial, ptr @partial_insertvalue_across_call, 1
  %value = insertvalue %reflect_value %with_data, i64 %number, 2
  %leaf = extractvalue %reflect_value %value, 0
  ret ptr %leaf
}

define goabiinternal ptr @phi_across_call(i1 %choose, %pair %left_value, %pair %right_value) gc "goallc" {
entry:
  br i1 %choose, label %left, label %right

left:
  br label %merge

right:
  br label %merge

merge:
  %value = phi %pair [ %left_value, %left ], [ %right_value, %right ]
  call goabiinternal void @safepoint()
  %result = extractvalue %pair %value, 0
  ret ptr %result
}

define goabiinternal ptr @select_across_call(i1 %choose, %pair %left_value, %pair %right_value) gc "goallc" {
entry:
  %value = select i1 %choose, %pair %left_value, %pair %right_value
  call goabiinternal void @safepoint()
  %result = extractvalue %pair %value, 0
  ret ptr %result
}

define goabiinternal ptr @phi_edge_use(%pair %value) gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  br label %merge

merge:
  %carried = phi %pair [ %value, %entry ]
  %result = extractvalue %pair %carried, 0
  ret ptr %result
}

define goabiinternal ptr @multiple_calls(%pair %value) gc "goallc" {
entry:
  call goabiinternal void @safepoint()
  call goabiinternal void @safepoint()
  %result = extractvalue %pair %value, 0
  ret ptr %result
}

define goabiinternal ptr @aggregate_call_result(ptr %pointer) gc "goallc" {
entry:
  %value = call goabiinternal %pair @make_pair(ptr %pointer, i64 7)
  call goabiinternal void @safepoint()
  %result = extractvalue %pair %value, 0
  ret ptr %result
}

define goabiinternal void @aggregate_current_call_argument(%pair %value) gc "goallc" {
entry:
  call goabiinternal void @consume_pair(%pair %value)
  ret void
}

define goabiinternal void @aggregate_load_store(ptr %source, ptr %destination) gc "goallc" {
entry:
  %value = load %pair, ptr %source, align 8
  call goabiinternal void @safepoint()
  store %pair %value, ptr %destination, align 8
  ret void
}

define goabiinternal ptr @alloca_derived_leaf(i64 %number) gc "goallc" {
entry:
  %slot = alloca i64, align 8
  store i64 %number, ptr %slot, align 8
  %value = insertvalue %pair poison, ptr %slot, 0
  call goabiinternal void @safepoint()
  %result = extractvalue %pair %value, 0
  ret ptr %result
}

define goabiinternal void @frozen_aggregate() gc "goallc" {
entry:
  %value = freeze %pair poison
  call goabiinternal void @safepoint()
  call goabiinternal void @leaf_consume_pair(%pair %value)
  ret void
}

attributes #0 = { "gc-leaf-function" }
