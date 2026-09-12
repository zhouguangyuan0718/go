// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package ssa

import (
	"cmd/compile/internal/abi"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
	"internal/buildcfg"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestLLVMCPUProfileBaselines(t *testing.T) {
	for _, test := range []struct {
		name, arch, profile string
		level               int
		lse, wantRequired   bool
	}{
		{"arm64-round", "arm64", goCPUProfileX86SSE41, 1, false, false},
		{"v1-round", "amd64", goCPUProfileX86SSE41, 1, false, true},
		{"v2-round", "amd64", goCPUProfileX86SSE41, 2, false, false},
		{"v1-fma", "amd64", goCPUProfileX86FMA, 1, false, true},
		{"v2-fma", "amd64", goCPUProfileX86FMA, 2, false, true},
		{"v3-fma", "amd64", goCPUProfileX86FMA, 3, false, false},
		{"v4-fma", "amd64", goCPUProfileX86FMA, 4, false, false},
		{"v1-popcnt", "amd64", goCPUProfileX86POPCNT, 1, false, true},
		{"v2-popcnt", "amd64", goCPUProfileX86POPCNT, 2, false, false},
		{"amd64-lse", "amd64", goCPUProfileARM64LSE, 1, false, false},
		{"arm64-v8.0", "arm64", goCPUProfileARM64LSE, 1, false, true},
		{"arm64-lse", "arm64", goCPUProfileARM64LSE, 1, true, false},
		{"v3-avx2", "amd64", goCPUProfileX86AVX2, 3, false, false},
		{"v4-avx512", "amd64", goCPUProfileX86AVX512, 4, false, false},
		{"v4-vbmi", "amd64", goCPUProfileX86AVX512VBMI, 4, false, true},
		{"v4-bitalg", "amd64", goCPUProfileX86AVX512BITALG, 4, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := llvmCPUProfileByName(test.profile)
			caps := llvmBaselineCPUCapabilities(test.arch, test.level, test.lse)
			required := p.arch == test.arch && caps&p.capabilities != p.capabilities
			if required != test.wantRequired {
				t.Errorf("required=%v, want %v", required, test.wantRequired)
			}
		})
	}
	avx2, avx512 := llvmCPUProfileByName(goCPUProfileX86AVX2), llvmCPUProfileByName(goCPUProfileX86AVX512)
	if avx512.capabilities&avx2.capabilities != avx2.capabilities || avx512.predicate&avx2.predicate != 0 {
		t.Fatal("AVX512 instruction capability was confused with the AVX2 predicate")
	}
}

func llvmCPUPlanBaselineForTest(t *testing.T) {
	t.Helper()
	amd64, arm64 := buildcfg.GOAMD64, buildcfg.GOARM64
	t.Cleanup(func() { buildcfg.GOAMD64, buildcfg.GOARM64 = amd64, arm64 })
	buildcfg.GOAMD64, buildcfg.GOARM64.LSE = 1, false
}

func llvmCPUPlanGuardedFunc(t *testing.T) fun {
	t.Helper()
	c := testConfig(t)
	pkg := types.NewPkg("internal/cpu", "cpu")
	x86 := types.NewStruct([]*types.Field{
		types.NewField(src.NoXPos, pkg.Lookup("HasAVX512"), types.Types[types.TBOOL]),
		types.NewField(src.NoXPos, pkg.Lookup("HasAVX2"), types.Types[types.TBOOL]),
	})
	types.CalcStructSize(x86)
	vec := llvmTestSIMDType("Int8x32", types.Types[types.TINT8], 32)
	f := c.Fun("entry",
		Bloc("entry",
			Valu("mem", OpInitMem, types.TypeMem, 0, nil),
			Valu("sb", OpSB, types.Types[types.TUINTPTR], 0, nil),
			Valu("addr", OpAddr, types.NewPtr(x86), 0, &obj.LSym{Name: "internal/cpu.X86"}, "sb"),
			Valu("hp", OpOffPtr, types.NewPtr(types.Types[types.TBOOL]), x86.Field(0).Offset, nil, "addr"),
			Valu("high", OpLoad, types.Types[types.TBOOL], 0, nil, "hp", "mem"),
			Valu("x", OpArg, vec, 0, nil), Valu("y", OpArg, vec, 0, nil),
			If("high", "body", "exit")),
		Bloc("body",
			Valu("lp", OpOffPtr, types.NewPtr(types.Types[types.TBOOL]), x86.Field(1).Offset, nil, "addr"),
			Valu("low", OpLoad, types.Types[types.TBOOL], 0, nil, "lp", "mem"),
			Valu("add", OpAddInt8x32, vec, 0, nil, "x", "y"),
			Valu("add2", OpAddInt8x32, vec, 0, nil, "add", "x"),
			Goto("exit")),
		Bloc("exit", Exit("mem")),
	)
	f.f.Entry.CPUfeatures = CPUavx
	return f
}

func TestLLVMCPUFeaturePlanStrongerGuard(t *testing.T) {
	llvmCPUPlanBaselineForTest(t)
	f := llvmCPUPlanGuardedFunc(t)
	p := llvmPlanCPUFeatures(f.f)
	if p.floor.profile != goCPUProfileX86AVX || p.floor.source != "entry-ssa" {
		t.Fatalf("floor=%+v", p.floor)
	}
	if got := strings.Join(p.profiles, ","); got != goCPUProfileX86AVX512 {
		t.Fatalf("profiles=%s", got)
	}
	for _, name := range []string{"add", "add2"} {
		if p.requirements[f.values[name].ID] != goCPUProfileX86AVX2 {
			t.Errorf("missing requirement for %s", name)
		}
	}
	if len(p.requirements) != 2 || len(p.guards) != 1 || p.guards[f.values["high"].ID] != goCPUProfileX86AVX512 {
		t.Fatalf("unexpected requirement/guard plan: %+v", p)
	}
	if _, marked := p.guards[f.values["low"].ID]; marked {
		t.Fatal("independent AVX2 predicate was selected by capability implication")
	}
	// Reordering the source storage must not change the immutable plan.
	slices.Reverse(f.f.Blocks)
	if other := llvmPlanCPUFeatures(f.f); !reflect.DeepEqual(p, other) {
		t.Fatalf("block-order-dependent plan: %+v vs %+v", p, other)
	}
}

func TestLLVMCPUFeaturePlanWideCalls(t *testing.T) {
	llvmCPUPlanBaselineForTest(t)
	for _, test := range []struct {
		name                 string
		registers, unguarded bool
	}{
		{"guarded-register-call", true, false},
		{"unguarded-register-call", true, true},
		{"unguarded-stack-only-call", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := llvmCPUPlanGuardedFunc(t)
			vec := llvmTestSIMDType("Uint8x64", types.Types[types.TUINT8], 64)
			registers := 0
			if test.registers {
				registers = 16
			}
			config := abi.NewABIConfig(registers, registers, 0, uint8(obj.ABIInternal))
			aux := StaticAuxCall(&obj.LSym{Name: "callee"}, config.ABIAnalyzeTypes(nil, []*types.Type{vec}))
			block := f.blocks["body"]
			if test.unguarded {
				block = f.blocks["exit"]
			}
			call := block.NewValue1A(src.NoXPos, OpStaticCall, aux.LateExpansionResultType(), aux, f.values["mem"])
			p := llvmPlanCPUFeatures(f.f)
			if test.unguarded && test.registers {
				if p.floor.profile != goCPUProfileX86AVX512 || p.floor.source != "wide-call" || p.floor.call != call.ID || len(p.profiles) != 0 || len(p.requirements) != 0 {
					t.Fatalf("unguarded call did not establish the whole-function floor: %+v", p)
				}
			} else {
				if p.floor.profile != goCPUProfileX86AVX || strings.Join(p.profiles, ",") != goCPUProfileX86AVX512 {
					t.Fatalf("unexpected floor/profiles: %+v", p)
				}
				want := ""
				if test.registers {
					want = goCPUProfileX86AVX512
				}
				if got := p.requirements[call.ID]; got != want {
					t.Errorf("call requirement=%q, want %q", got, want)
				}
			}
			slices.Reverse(f.f.Blocks)
			if other := llvmPlanCPUFeatures(f.f); !reflect.DeepEqual(p, other) {
				t.Fatal("wide-call planning depends on block order")
			}
		})
	}
}

func TestLLVMCPUFeaturePlanMidway(t *testing.T) {
	llvmCPUPlanBaselineForTest(t)
	f := llvmCPUPlanGuardedFunc(t)
	f.f.Name = "f@simd256"
	p := llvmPlanCPUFeatures(f.f)
	if p.floor.profile != goCPUProfileX86AVX2 || p.floor.source != "midway" || len(p.profiles) != 0 || len(p.requirements) != 0 {
		t.Fatalf("Midway AVX2 floor created a redundant dispatch: %+v", p)
	}
	if len(p.guards) != 0 {
		t.Fatal("entry capability incorrectly selected runtime predicates")
	}
}

func TestLLVMCPUFeaturePlanCompilerOps(t *testing.T) {
	llvmCPUPlanBaselineForTest(t)
	for _, test := range []struct {
		arch, symbol, profile string
		op                    Op
	}{
		{"amd64", "runtime.x86HasFMA", goCPUProfileX86FMA, OpFMA},
		{"amd64", "runtime.x86HasSSE41", goCPUProfileX86SSE41, OpFloor},
		{"amd64", "runtime.x86HasPOPCNT", goCPUProfileX86POPCNT, OpPopCount64},
		{"arm64", "runtime.arm64HasATOMICS", goCPUProfileARM64LSE, OpAtomicAdd64Variant},
	} {
		t.Run(test.profile, func(t *testing.T) {
			c := testConfigArch(t, test.arch)
			f := c.Fun("entry", Bloc("entry", Valu("mem", OpInitMem, types.TypeMem, 0, nil), Exit("mem")))
			b := f.f.Entry
			value := b.NewValue0(src.NoXPos, test.op, types.Types[types.TUINT64])
			var guard *Value
			if test.arch == "amd64" {
				guard = b.NewValue0A(src.NoXPos, OpHasCPUFeature, types.Types[types.TBOOL], &obj.LSym{Name: test.symbol})
			} else {
				addr := b.NewValue0A(src.NoXPos, OpAddr, types.NewPtr(types.Types[types.TBOOL]), &obj.LSym{Name: test.symbol})
				guard = b.NewValue2(src.NoXPos, OpLoad, types.Types[types.TBOOL], addr, f.values["mem"])
			}
			p := llvmPlanCPUFeatures(f.f)
			if p.requirements[value.ID] != test.profile || p.guards[guard.ID] != test.profile || strings.Join(p.profiles, ",") != test.profile {
				t.Fatalf("compiler-op plan: %+v", p)
			}
			// Baseline coverage removes the requirement and dispatch request,
			// but retains the existing compiler-generated guard marker policy.
			buildcfg.GOAMD64, buildcfg.GOARM64.LSE = 4, true
			p = llvmPlanCPUFeatures(f.f)
			buildcfg.GOAMD64, buildcfg.GOARM64.LSE = 1, false
			if len(p.requirements) != 0 || len(p.profiles) != 0 || p.guards[guard.ID] != test.profile {
				t.Fatalf("baseline compiler-op plan: %+v", p)
			}
		})
	}
}

func TestLLVMCPUFeaturePlanGeneratedCoverage(t *testing.T) {
	for op := Op(0); int(op) < len(goALLCSIMDOpcodeIndex); op++ {
		info, ok := goALLCSIMDInfo(op)
		if !ok {
			continue
		}
		for _, arch := range []string{"amd64", "arm64"} {
			got, kind := llvmCPURequirement(&Value{Op: op}, arch)
			if want := info.archInfo(arch).cpuProfile; got != want || kind != llvmCPUGenerated {
				t.Errorf("%s/%s: requirement=%s kind=%d, want %s", op, arch, got, kind, want)
			}
		}
	}
	for _, op := range []Op{OpRound, OpAtomicAdd64, OpAdd64} {
		for _, arch := range []string{"amd64", "arm64"} {
			if got, _ := llvmCPURequirement(&Value{Op: op}, arch); got != "" {
				t.Errorf("%s/%s got an extra requirement %s", op, arch, got)
			}
		}
	}
}
