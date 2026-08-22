target triple = "x86_64-unknown-linux-goobj"

; OBJVIEW-LABEL: "name": "different_pointer_sets_across_calls"
; OBJVIEW: "size": [[#SIZE:]]
; OBJVIEW: "kind": "unsafe_point"
; OBJVIEW: "start": 0
; OBJVIEW-NEXT: "end": [[#STACK_CHECK_END:]]
; OBJVIEW-NEXT: "value": -2
; OBJVIEW: "start": [[#STACK_CHECK_END]]
; OBJVIEW-NEXT: "end": [[#MORESTACK_BEGIN:]]
; OBJVIEW-NEXT: "value": -1
; OBJVIEW: "start": [[#MORESTACK_BEGIN]]
; OBJVIEW-NEXT: "end": [[#MORESTACK_END:]]
; OBJVIEW-NEXT: "value": -2
; OBJVIEW: "start": [[#MORESTACK_END]]
; OBJVIEW-NEXT: "end": [[#SIZE]]
; OBJVIEW-NEXT: "value": -1
; OBJVIEW: "kind": "stack_map_index"
; OBJVIEW: "start": 0
; OBJVIEW-NEXT: "end": [[#FIRST_START:]]
; OBJVIEW-NEXT: "value": -1
; OBJVIEW: "start": [[#FIRST_START]]
; OBJVIEW-NEXT: "end": [[#SECOND_START:]]
; OBJVIEW-NEXT: "value": 1
; OBJVIEW: "start": [[#SECOND_START]]
; OBJVIEW-NEXT: "end": [[#ENTRY_DEPTH_START:]]
; OBJVIEW-NEXT: "value": 2
; OBJVIEW: "start": [[#ENTRY_DEPTH_START]]
; OBJVIEW-NEXT: "end": [[#SIZE]]
; OBJVIEW-NEXT: "value": -1
; OBJVIEW: "kind": "locals_pointer_maps"
; OBJVIEW: "count": 3
; OBJVIEW: "index": 0
; OBJVIEW-NEXT: "set_bits": null
; OBJVIEW: "index": 1
; OBJVIEW: "set_bits": [
; OBJVIEW-NEXT: [[#FIRST_BIT:]],
; OBJVIEW-NEXT: [[#SECOND_BIT:]]
; OBJVIEW: "index": 2
; OBJVIEW: "set_bits": [
; OBJVIEW-NEXT: [[#SECOND_MAP_BIT:]]
; OBJVIEW: "call_offset": [[#FIRST_START+1]]
; OBJVIEW: "stack_map_index": 1
; OBJVIEW: "call_offset": [[#SECOND_START+1]]
; OBJVIEW: "stack_map_index": 2
; OBJVIEW: "stack_map_index": -1

declare goabiinternal void @first_callee()
declare goabiinternal void @second_callee()

define goabiinternal i64 @different_pointer_sets_across_calls(ptr %p, ptr %q) gc "goallc" {
entry:
  call goabiinternal void @first_callee()
  %first = load i64, ptr %p, align 8
  call goabiinternal void @second_callee()
  %second = load i64, ptr %q, align 8
  %sum = add i64 %first, %second
  ret i64 %sum
}
