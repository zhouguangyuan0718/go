target triple = "aarch64-unknown-linux-goobj"

declare i64 @llvm.read_register.i64(metadata)

define goabiinternal i64 @aarch64_regtmp_range(i64 %v) #0 gc "goallc" {
entry:
  %before = add i64 %v, 1
  %tmp = call i64 @llvm.read_register.i64(metadata !0)
  %after = add i64 %before, %tmp
  ret i64 %after
}

attributes #0 = { "go-async-unsafe" }

!0 = !{!"x27"}

; OBJVIEW-LABEL: "name": "aarch64_regtmp_range"
; OBJVIEW: "size": [[#SIZE:]]
; OBJVIEW: "kind": "unsafe_point"
; OBJVIEW: "start": 0
; OBJVIEW-NEXT: "end": 4
; OBJVIEW-NEXT: "value": -2
; OBJVIEW: "start": 4
; OBJVIEW-NEXT: "end": [[#SIZE]]
; OBJVIEW-NEXT: "value": -1
