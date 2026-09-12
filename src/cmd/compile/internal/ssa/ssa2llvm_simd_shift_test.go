// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/types"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/goallc/go-llvm"
)

// Use Go's unbounded shift-count semantics as the oracle, independently of
// the lowering's safe-count normalization and result selection.
func ordinarySIMDShiftReference(x, count uint64, bits int, signed, right bool) uint64 {
	mask := ^uint64(0) >> (64 - bits)
	x &= mask
	if !right {
		return x << count & mask
	}
	if signed {
		// First sign-extend the original lane to int64. Narrow negative
		// lanes must not accidentally become positive int64 values.
		s := int64(x<<(64-bits)) >> (64 - bits)
		return uint64(s>>count) & mask
	}
	return x >> count
}

func ordinarySIMDShiftCounts(bits int) []uint64 {
	mask := ^uint64(0) >> (64 - bits)
	counts := make([]uint64, 256)
	for i := range counts {
		counts[i] = uint64(i)
	}
	for i := 8; i < bits; i++ {
		counts = append(counts, uint64(1)<<i, mask^(uint64(1)<<i))
	}
	return append(counts, mask)
}

func TestLLVMGeneratedSIMDOrdinaryShifts(t *testing.T) {
	oldTypes, oldModule := type2lTypes, CurrentModule
	type2lTypes = make(map[*types.Type]llvm.Type)
	defer func() { type2lTypes, CurrentModule = oldTypes, oldModule }()

	pattern := regexp.MustCompile(`^(ShiftAllLeft|ShiftAllRight|ShiftLeft|ShiftRight)(Int|Uint)(8|16|32|64)x([0-9]+)$`)
	recipes := map[string]goALLCSIMDLowering{
		"ShiftAllLeft":  goALLCSIMDLowerShiftAllLeft,
		"ShiftAllRight": goALLCSIMDLowerShiftAllRight,
		"ShiftLeft":     goALLCSIMDLowerShiftLeft,
		"ShiftRight":    goALLCSIMDLowerShiftRight,
	}
	families := make(map[string]int)
	opCount, archCount := 0, 0
	for op := Op(0); int(op) < len(goALLCSIMDOpcodeIndex); op++ {
		match := pattern.FindStringSubmatch(op.String())
		if match == nil {
			continue
		}
		opCount++
		family, kind := match[1], strings.ToLower(match[2])
		families[family]++
		bits, _ := strconv.Atoi(match[3])
		n, _ := strconv.Atoi(match[4])
		scalar := strings.HasPrefix(family, "ShiftAll")
		right := strings.HasSuffix(family, "Right")
		signed := kind == "int"
		info, ok := goALLCSIMDInfo(op)
		if !ok || info.lowering != recipes[family] {
			t.Errorf("%s has lowering %v (present=%v), want %v", op, info.lowering, ok, recipes[family])
			continue
		}
		wantLane := goALLCSIMDLaneUint
		if signed {
			wantLane = goALLCSIMDLaneInt
		}
		if info.lane != wantLane || int(info.laneBits) != bits {
			t.Fatalf("%s descriptor does not describe its data lanes", op)
		}
		if bits != 8 && ((!scalar && bits == 16) || (signed && right && bits == 64)) && info.amd64.cpuProfile != goCPUProfileX86AVX512 {
			t.Fatalf("%s must retain its AVX512 capability requirement", op)
		}

		// The shared scalar-count APIs cover all 128-bit integer shapes on
		// arm64. Other ordinary-shift shapes are amd64 APIs; 8-bit shifts
		// are arm64-only in this batch.
		archs := []string{"amd64"}
		if bits == 8 {
			archs = []string{"arm64"}
		} else if scalar && bits*n == 128 {
			archs = append(archs, "arm64")
		}
		for _, arch := range archs {
			archCount++
			for _, carrier := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/carrier=%v", op, arch, carrier), func(t *testing.T) {
					kinds := map[string]map[int]types.Kind{
						"int":  {8: types.TINT8, 16: types.TINT16, 32: types.TINT32, 64: types.TINT64},
						"uint": {8: types.TUINT8, 16: types.TUINT16, 32: types.TUINT32, 64: types.TUINT64},
					}
					vector := llvmTestSIMDType("shift-data", types.Types[kinds[kind][bits]], int64(n))
					countType := llvmTestSIMDType("shift-counts", types.Types[kinds["uint"][bits]], int64(n))
					integer := GlobalCtxt.IntType(bits)
					resultType := llvm.VectorType(integer, n)
					if carrier {
						vector = map[int]*types.Type{128: types.TypeVec128, 256: types.TypeVec256, 512: types.TypeVec512}[bits*n]
						countType = vector
					}
					countBits := bits
					if scalar {
						countType = types.Types[types.TUINT64]
						countBits = 64
					}
					countMask := ^uint64(0) >> (64 - countBits)
					module := GlobalCtxt.NewModule("ordinary-shift")
					CurrentModule = module
					builder := GlobalCtxt.NewBuilder()
					defer module.Dispose()
					defer builder.Dispose()
					fn := llvm.AddFunction(module, "shift", llvm.FunctionType(resultType, nil, false))
					builder.SetInsertPointAtEnd(llvm.AddBasicBlock(fn, "entry"))
					xArg := &Value{ID: 1, Op: OpArg, Type: vector}
					countArg := &Value{ID: 2, Op: OpArg, Type: countType}
					v := &Value{ID: 3, Op: op, Type: vector, Args: []*Value{xArg, countArg}}
					mask := ^uint64(0) >> (64 - bits)
					sign := uint64(1) << (bits - 1)
					data := []uint64{0, 1, mask, sign, sign - 1, sign + 1, mask - 1, 0xa55aa55aa55aa55a & mask}
					var result llvm.Value
					// Rotate actual edge values through every lane and cross them
					// with every count. Lane zero visits the exact control while
					// adjacent lanes mix in-range and out-of-range vector counts.
					for phase := range data {
						x := make([]uint64, n)
						xValues := make([]llvm.Value, n)
						for i := range x {
							x[i] = data[(phase+i)%len(data)]
							xValues[i] = llvm.ConstInt(integer, x[i], false)
						}
						xValue := llvm.ConstBitCast(llvm.ConstVector(xValues, false), getLLVMType(vector))
						for _, control := range ordinarySIMDShiftCounts(countBits) {
							counts := make([]uint64, n)
							var countValue llvm.Value
							if scalar {
								countValue = llvm.ConstInt(GlobalCtxt.Int64Type(), control, false)
								for i := range counts {
									counts[i] = control
								}
							} else {
								values := make([]llvm.Value, n)
								for i := range counts {
									counts[i] = (control + uint64(i)) & countMask
									values[i] = llvm.ConstInt(integer, counts[i], false)
								}
								countValue = llvm.ConstBitCast(llvm.ConstVector(values, false), getLLVMType(countType))
							}
							// Constant folding runs under each descriptor's exact
							// capability floor. The codegen fixtures independently
							// check source guards and FMV for dynamic operands.
							ctx := &LLVMFuncContext{
								F:               &Func{Config: &Config{arch: arch}, Entry: &Block{CPUfeatures: CPUavx | CPUavx2 | CPUavx512}},
								CPUFeatureFloor: info.archInfo(arch).cpuProfile,
								Vs:              map[ID]llvm.Value{1: xValue, 2: countValue},
								b:               builder,
							}
							result = ctx.GenLV(v)
							if result.Type() != resultType || result.IsAConstant().IsNil() {
								t.Fatalf("phase=%d count=%#x did not fold to natural lane vector: %s", phase, control, result.String())
							}
							for i, value := range x {
								want := ordinarySIMDShiftReference(value, counts[i], bits, signed, right)
								got := llvm.ConstExtractElement(result, llvm.ConstInt(GlobalCtxt.Int32Type(), uint64(i), false))
								if got.IsAConstantInt().IsNil() || got.ZExtValue() != want {
									t.Fatalf("phase=%d lane=%d x=%#x count=%#x got=%s want=%#x", phase, i, value, counts[i], got.String(), want)
								}
							}
						}
					}
					builder.CreateRet(result)
					if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
						t.Fatal(err)
					}
					for _, forbidden := range []string{" poison", " undef", "@llvm.x86.", "@llvm.aarch64."} {
						if strings.Contains(module.String(), forbidden) {
							t.Fatalf("unexpected %s in IR", forbidden)
						}
					}
				})
			}
		}
	}
	if opCount != 76 || archCount != 88 {
		t.Errorf("tested %d ordinary shift descriptors across %d architecture instances, want 76 and 88", opCount, archCount)
	}
	for family, want := range map[string]int{"ShiftAllLeft": 20, "ShiftAllRight": 20, "ShiftLeft": 18, "ShiftRight": 18} {
		if families[family] != want {
			t.Errorf("tested %d %s descriptors, want %d", families[family], family, want)
		}
	}
}
