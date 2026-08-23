// asmcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import _ "unsafe"

//go:linkname llvmLinknameExternal runtime.llvmLinknameExternal
func llvmLinknameExternal() int

//go:linkname llvmLinknameLocal runtime.llvmLinknameLocal
func llvmLinknameLocal() int {
	return 7
}

//go:linkname llvmLinknameVoid runtime.llvmLinknameVoid
//go:noinline
func llvmLinknameVoid() {
	llvmLinknameSink++
}

var llvmLinknameSink int

// LLVM-LABEL: define goabiinternal i64 @codegen.llvmLinknameCalls()
// LLVM: call goabiinternal i64 @"runtime.llvmLinknameExternal<linkname>"()
// LLVM: declare goabiinternal i64 @"runtime.llvmLinknameExternal<linkname>"()
// LLVM-NOT: @"runtime.llvmLinknameLocal<linkname>"
// LLVM-LABEL: define weak goabi0 void @"runtime.llvmLinknameLocal<ABI0>"(
// LLVM-SAME: ptr goret(i64) align 8 "goretindex"="0" [[RESULT_HOME:%[^)]+]])
// LLVM: [[RESULT:%.*]] = call goabiinternal i64 @runtime.llvmLinknameLocal()
// LLVM-NEXT: store i64 [[RESULT]], ptr [[RESULT_HOME]], align 8
// LLVM-LABEL: define goabiinternal i64 @runtime.llvmLinknameLocal()
// LLVM-LABEL: define weak goabi0 void @"runtime.llvmLinknameVoid<ABI0>"()
// LLVM: musttail call goabiinternal void @runtime.llvmLinknameVoid()
// LLVM-NEXT: ret void
// LLVM-OPT-LABEL: define goabiinternal i64 @codegen.llvmLinknameCalls()
// LLVM-OPT: call goabiinternal i64 @"runtime.llvmLinknameExternal<linkname>"()
// LLVM-OPT: declare goabiinternal i64 @"runtime.llvmLinknameExternal<linkname>"()
// LLVM-OPT-NOT: @"runtime.llvmLinknameLocal<linkname>"
// LLVM-OPT-LABEL: define weak goabi0 void @"runtime.llvmLinknameLocal<ABI0>"(
// LLVM-OPT-SAME: ptr {{.*}}goret(i64) align 8{{.*}} "goretindex"="0" [[OPT_RESULT_HOME:%[^)]+]])
// LLVM-OPT: store i64 7, ptr [[OPT_RESULT_HOME]], align 8
// LLVM-OPT-LABEL: define goabiinternal {{.*}}@runtime.llvmLinknameLocal()
// LLVM-OPT-LABEL: define weak goabi0 void @"runtime.llvmLinknameVoid<ABI0>"()
// LLVM-OPT: musttail call goabiinternal void @runtime.llvmLinknameVoid()
// LLVM-OPT-NEXT: ret void
func llvmLinknameCalls() int {
	llvmLinknameVoid()
	return llvmLinknameExternal() + llvmLinknameLocal()
}
