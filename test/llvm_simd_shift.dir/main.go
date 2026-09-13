// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
)

var enabledCases, fallbackCases, skippedCases int
var checkedScalarCounts, checkedVectorCounts, checkedLanes int
var constantEnabledCases, constantFallbackCases, constantSkippedCases, checkedConstantVectors int

func main() {
	checkArch()
	if cases := enabledCases + fallbackCases + skippedCases; cases != archShiftCases {
		panic(fmt.Sprintf("ordinary shifts: covered %d cases, want %d", cases, archShiftCases))
	}
	checkConstantArch()
	if cases := constantEnabledCases + constantFallbackCases + constantSkippedCases; cases != archConstantShiftCases {
		panic(fmt.Sprintf("ordinary shift constants: covered %d cases, want %d", cases, archConstantShiftCases))
	}
	if os.Getenv("GOALLC_SIMD_SHIFT_TRACE") == "1" {
		fmt.Printf("ordinary shifts: enabled=%d fallback=%d skipped=%d vectors=%d scalar-counts=%d vector-counts=%d lanes=%d\n",
			enabledCases, fallbackCases, skippedCases, checkedScalarCounts+checkedVectorCounts,
			checkedScalarCounts, checkedVectorCounts, checkedLanes)
		fmt.Printf("ordinary shift constants: enabled=%d fallback=%d skipped=%d vectors=%d\n",
			constantEnabledCases, constantFallbackCases, constantSkippedCases, checkedConstantVectors)
	}
}
