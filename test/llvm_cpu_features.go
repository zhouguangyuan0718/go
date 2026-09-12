// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"math/bits"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
)

var llvmCPUCounter uint64

//go:noinline
func llvmCPUMath(x, y, z float64) (float64, float64, string) {
	return math.Floor(x), math.FMA(x, y, z), llvmCPUCallerName()
}

//go:noinline
func llvmCPUBits(x uint64) int {
	return bits.OnesCount64(x)
}

//go:noinline
func llvmCPUAtomic(delta uint64) (uint64, string) {
	return atomic.AddUint64(&llvmCPUCounter, delta), llvmCPUCallerName()
}

//go:noinline
func llvmCPUCallerName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}
	return fn.Name()
}

func llvmCPUCheck() {
	atomic.StoreUint64(&llvmCPUCounter, 0)
	floor, fma, mathName := llvmCPUMath(3.75, 2, 3)
	atomicValue, atomicName := llvmCPUAtomic(7)
	if floor != 3 || fma != 10.5 || llvmCPUBits(0xf0f0) != 8 || atomicValue != 7 ||
		mathName != "main.llvmCPUMath" || atomicName != "main.llvmCPUAtomic" {
		panic("bad CPU-feature multiversion result")
	}
	// Exercise fused rounding and exceptional values in both Go's hardware
	// and fallback paths, including children with individual features disabled.
	for _, test := range []struct{ x, y, z, want uint64 }{
		{0x3ff0000002000000, 0x3feffffffc000000, 0xbff0000000000000, 0xbc90000000000000},
		{0x7fefffffffffffff, 0x4000000000000000, 0xffefffffffffffff, 0x7fefffffffffffff},
		{0x0000000000000001, 0x3ff0000000000000, 0x8000000000000000, 0x0000000000000001},
		{0x8000000000000000, 0x4000000000000000, 0x8000000000000000, 0x8000000000000000},
		{0x7ff0000000000000, 0x4000000000000000, 0x3ff0000000000000, 0x7ff0000000000000},
	} {
		_, got, _ := llvmCPUMath(math.Float64frombits(test.x), math.Float64frombits(test.y), math.Float64frombits(test.z))
		if math.Float64bits(got) != test.want {
			panic("CPU-feature path changed fused rounding or exceptional value")
		}
	}
}

func llvmCPUChildEnv(godebug string) []string {
	env := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GODEBUG=") && !strings.HasPrefix(entry, "GOALLC_CPU_FEATURE_CHILD=") {
			env = append(env, entry)
		}
	}
	return append(env, "GODEBUG="+godebug, "GOALLC_CPU_FEATURE_CHILD=1")
}

func main() {
	llvmCPUCheck()
	if os.Getenv("GOALLC_CPU_FEATURE_CHILD") == "1" {
		return
	}

	var modes []string
	switch runtime.GOARCH {
	case "amd64":
		modes = []string{"cpu.all=off", "cpu.sse41=off", "cpu.fma=off", "cpu.popcnt=off"}
	case "arm64":
		modes = []string{"cpu.atomics=off", "cpu.atomics=on"}
	}
	for _, mode := range modes {
		cmd := exec.Command(os.Args[0])
		cmd.Env = llvmCPUChildEnv(mode)
		if output, err := cmd.CombinedOutput(); err != nil {
			panic("GODEBUG=" + mode + ": " + err.Error() + ": " + string(output))
		}
	}
}
