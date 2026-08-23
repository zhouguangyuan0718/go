// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
)

var llvmAsyncPreemptReady uint32

//go:noinline
func llvmAsyncPreemptSpin() {
	for {
	}
}

func main() {
	runtime.GOMAXPROCS(1)
	debug.SetGCPercent(-1)
	runtime.GC()

	go func() {
		atomic.StoreUint32(&llvmAsyncPreemptReady, 1)
		llvmAsyncPreemptSpin()
	}()
	for atomic.LoadUint32(&llvmAsyncPreemptReady) == 0 {
		runtime.Gosched()
	}

	// The spinning goroutine has no call or synchronous safe point. Stopping it
	// for GC requires the LLVM function's asynchronous safe-point metadata.
	runtime.GC()
}
