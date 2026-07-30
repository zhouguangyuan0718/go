; Copyright 2026 The Go Authors. All rights reserved.
; Use of this source code is governed by a BSD-style
; license that can be found in the LICENSE file.

%pair = type { ptr, i64, ptr }

declare goabiinternal void @safepoint()
declare goabiinternal void @consume_pair(ptr byval(%pair) align 8)

define goabiinternal void @non_transient_byval_source(%pair %value)
    "go-stack-growth-statepoint" gc "goallc" {
entry:
  %value.byval = alloca %pair, align 8
  store %pair %value, ptr %value.byval, align 8
  call goabiinternal void @safepoint()
  call goabiinternal void @consume_pair(
      ptr byval(%pair) align 8 %value.byval)
  ret void
}
