target triple = "x86_64-unknown-linux-goobj"

declare goabiinternal void @first_callee()
declare goabiinternal void @second_callee()

define goabiinternal i64 @different_pointer_sets_across_calls(ptr %p, ptr %q) "go-stack-growth-statepoint" gc "goallc" {
entry:
  call goabiinternal void @first_callee()
  %first = load i64, ptr %p, align 8
  call goabiinternal void @second_callee()
  %second = load i64, ptr %q, align 8
  %sum = add i64 %first, %second
  ret i64 %sum
}
