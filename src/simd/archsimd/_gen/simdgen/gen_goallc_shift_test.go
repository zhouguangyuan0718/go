// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"testing"
)

func goALLCShiftTestPointer[T any](value T) *T { return &value }

func goALLCShiftTestVector(name string) Operand {
	base, bits, lanes, _ := goALLCLaneFromGoType(&name)
	width := bits * lanes
	return Operand{Class: "vreg", Go: &name, Base: &base, ElemBits: &bits, Lanes: &lanes, Bits: &width}
}

func goALLCShiftTestScalar(name string) Operand {
	return Operand{Class: "greg", Go: &name, Bits: goALLCShiftTestPointer(64)}
}

func goALLCShiftTestNativeScalar(name string, bits int) Operand {
	operand := goALLCShiftTestVector(name)
	operand.TreatLikeAScalarOfSize = &bits
	return operand
}

func goALLCShiftTestOperation(lowering, data string, count Operand) Operation {
	return Operation{
		rawOperation: rawOperation{Out: []Operand{goALLCShiftTestVector(data)}, LLVMLowering: &lowering},
		Go:           "Shift",
		In:           []Operand{goALLCShiftTestVector(data), count},
	}
}

func checkGoALLCShiftValidation(t *testing.T, op, genericOp Operation, lowering, wantPanic string) {
	t.Helper()
	defer func() {
		err := recover()
		if wantPanic == "" {
			if err != nil {
				t.Fatalf("valid shift rejected: %v", err)
			}
		} else if err == nil || !strings.Contains(fmt.Sprint(err), wantPanic) {
			t.Fatalf("validation panic = %v, want diagnostic containing %q", err, wantPanic)
		}
	}()
	validateGoALLCLowering(op, genericOp, lowering, PureVregIn, OneVregOut, NoMask, NoImm)
}

func TestGoALLCShiftShape(t *testing.T) {
	for _, lowering := range []string{"shift-all-left", "shift-all-right", "shift-left", "shift-right"} {
		for _, base := range []string{"Int", "Uint"} {
			for _, bits := range []int{8, 16, 32, 64} {
				for _, width := range []int{128, 256, 512} {
					all := strings.HasPrefix(lowering, "shift-all-")
					if bits == 8 && (!all || width != 128) {
						continue // Byte shifts are scalar-count ARM64 operations.
					}
					name := fmt.Sprintf("%s%dx%d", base, bits, width/bits)
					t.Run(lowering+"/"+name, func(t *testing.T) {
						counts := []Operand{goALLCShiftTestVector(fmt.Sprintf("Uint%dx%d", bits, width/bits))}
						if all {
							counts = []Operand{
								goALLCShiftTestScalar("uint64"),
								goALLCShiftTestNativeScalar("Uint64x2", 64),
								goALLCShiftTestNativeScalar("Int8x16", 64),
							}
						}
						for _, count := range counts {
							op := goALLCShiftTestOperation(lowering, name, count)
							checkGoALLCShiftValidation(t, op, op, lowering, "")
						}
					})
				}
			}
		}
	}
}

func TestGoALLCShiftRejectsMalformedOperands(t *testing.T) {
	for _, lowering := range []string{"shift-all-left", "shift-all-right", "shift-left", "shift-right"} {
		for _, test := range []struct {
			name, diagnostic string
			mutate           func(*Operation)
		}{
			{"missing-count", "has 1 inputs, want 2", func(op *Operation) { op.In = op.In[:1] }},
			{"extra-count", "has 3 inputs, want 2", func(op *Operation) { op.In = append(op.In, op.In[1]) }},
			{"missing-result", "has 0 outputs, want 1", func(op *Operation) { op.Out = nil }},
			{"extra-result", "has 2 outputs, want 1", func(op *Operation) { op.Out = append(op.Out, op.Out[0]) }},
			{"operand-order", "invalid operand order", func(op *Operation) { op.OperandOrder = goALLCShiftTestPointer("21Type1") }},
			{"commutative", "invalid operand order", func(op *Operation) { op.Commutative = true }},
			{"swapped-inputs", "integer data shape", func(op *Operation) { op.In[0], op.In[1] = op.In[1], op.In[0] }},
			{"float-data", "integer data shape", func(op *Operation) { op.In[0] = goALLCShiftTestVector("Float32x4") }},
			{"invalid-data-width", "integer data shape", func(op *Operation) { op.In[0] = goALLCShiftTestVector("Int32x3") }},
			{"missing-data-type", "integer data shape", func(op *Operation) { op.In[0].Go = nil }},
			{"scalar-data", "incompatible vector lanes", func(op *Operation) { op.In[0].TreatLikeAScalarOfSize = goALLCShiftTestPointer(64) }},
			{"memory-data", "incompatible vector lanes", func(op *Operation) { op.In[0].Class = "mem" }},
			{"missing-data-width", "incompatible vector lanes", func(op *Operation) { op.In[0].Bits = nil }},
			{"wrong-data-width", "incompatible vector lanes", func(op *Operation) { op.In[0].Bits = goALLCShiftTestPointer(256) }},
			{"result-signedness", "incompatible vector lanes", func(op *Operation) { op.Out[0] = goALLCShiftTestVector("Uint32x4") }},
			{"result-lane-width", "incompatible vector lanes", func(op *Operation) { op.Out[0] = goALLCShiftTestVector("Int64x2") }},
			{"result-lane-count", "incompatible vector lanes", func(op *Operation) { op.Out[0] = goALLCShiftTestVector("Int32x8") }},
			{"mask-result", "incompatible vector lanes", func(op *Operation) { op.Out[0].Class = "mask" }},
			{"scalar-result", "incompatible vector lanes", func(op *Operation) { op.Out[0].TreatLikeAScalarOfSize = goALLCShiftTestPointer(64) }},
			{"wrong-result-width", "incompatible vector lanes", func(op *Operation) { op.Out[0].Bits = goALLCShiftTestPointer(256) }},
			{"data-list", "invalid list/immediate", func(op *Operation) { op.In[0].ListNumber = goALLCShiftTestPointer(0) }},
			{"count-list", "invalid list/immediate", func(op *Operation) { op.In[1].ListNumber = goALLCShiftTestPointer(0) }},
			{"result-list", "invalid list/immediate", func(op *Operation) { op.Out[0].ListNumber = goALLCShiftTestPointer(0) }},
			{"constant-count", "invalid list/immediate", func(op *Operation) { op.In[1].Const = goALLCShiftTestPointer("1") }},
			{"immediate-offset", "invalid list/immediate", func(op *Operation) { op.In[1].ImmOffset = goALLCShiftTestPointer("0") }},
			{"immediate-limit", "invalid list/immediate", func(op *Operation) { op.In[1].ImmMax = goALLCShiftTestPointer(63) }},
		} {
			t.Run(lowering+"/"+test.name, func(t *testing.T) {
				count := goALLCShiftTestVector("Uint32x4")
				if strings.HasPrefix(lowering, "shift-all-") {
					count = goALLCShiftTestScalar("uint64")
				}
				op := goALLCShiftTestOperation(lowering, "Int32x4", count)
				genericOp := goALLCShiftTestOperation(lowering, "Int32x4", count)
				test.mutate(&genericOp)
				diagnostic := test.diagnostic
				if test.name == "swapped-inputs" && !strings.HasPrefix(lowering, "shift-all-") {
					diagnostic = "incompatible vector lanes"
				}
				checkGoALLCShiftValidation(t, op, genericOp, lowering, diagnostic)
			})
		}
	}
}

func TestGoALLCShiftRejectsMalformedCounts(t *testing.T) {
	for _, lowering := range []string{"shift-all-left", "shift-all-right", "shift-left", "shift-right"} {
		counts := []Operand{
			goALLCShiftTestVector("Int32x4"),
			goALLCShiftTestVector("Float32x4"),
			goALLCShiftTestVector("Uint16x8"),
			goALLCShiftTestVector("Uint32x8"),
			goALLCShiftTestScalar("int64"),
			goALLCShiftTestScalar("uint32"),
			goALLCShiftTestNativeScalar("Float64x2", 64),
			goALLCShiftTestNativeScalar("Uint64x2", 32),
		}
		diagnostic := "incompatible vector lanes"
		if strings.HasPrefix(lowering, "shift-all-") {
			diagnostic = "uint64 scalar count"
			counts = append(counts, goALLCShiftTestVector("Uint32x4"))
		} else {
			counts = append(counts, goALLCShiftTestScalar("uint64"), goALLCShiftTestNativeScalar("Int8x16", 64))
		}
		for i, count := range counts {
			t.Run(fmt.Sprintf("%s/%d", lowering, i), func(t *testing.T) {
				op := goALLCShiftTestOperation(lowering, "Int32x4", count)
				checkGoALLCShiftValidation(t, op, op, lowering, diagnostic)
			})
		}
		for _, test := range []struct {
			name   string
			mutate func(*Operand)
		}{
			{"missing-type", func(count *Operand) { count.Go = nil }},
			{"missing-width", func(count *Operand) { count.Bits = nil }},
			{"wrong-width", func(count *Operand) { count.Bits = goALLCShiftTestPointer(32) }},
			{"mask", func(count *Operand) { count.Class = "mask" }},
			{"immediate", func(count *Operand) { count.Class = "immediate" }},
			{"memory", func(count *Operand) { count.Class = "mem" }},
		} {
			t.Run(lowering+"/"+test.name, func(t *testing.T) {
				count := goALLCShiftTestVector("Uint32x4")
				if strings.HasPrefix(lowering, "shift-all-") {
					count = goALLCShiftTestNativeScalar("Int8x16", 64)
				}
				test.mutate(&count)
				op := goALLCShiftTestOperation(lowering, "Int32x4", count)
				checkGoALLCShiftValidation(t, op, op, lowering, diagnostic)
			})
		}
	}
}

func TestGoALLCShiftControlShape(t *testing.T) {
	for _, test := range []struct {
		name string
		in   inShape
		out  outShape
		mask maskShape
		imm  immShape
	}{
		{"mask-input", OneKmaskIn, OneVregOut, OneMask, NoImm},
		{"list-input", VlistIn, OneVregOut, NoMask, NoImm},
		{"scalar-output", PureVregIn, OneGregOut, NoMask, NoImm},
		{"immediate", PureVregIn, OneVregOut, NoMask, VarImm},
		{"mask", PureVregIn, OneVregOut, OneMask, NoImm},
	} {
		for _, lowering := range []string{"shift-all-left", "shift-all-right", "shift-left", "shift-right"} {
			t.Run(lowering+"/"+test.name, func(t *testing.T) {
				op := goALLCShiftTestOperation(lowering, "Int32x4", goALLCShiftTestScalar("uint64"))
				defer func() {
					if err := recover(); err == nil || !strings.Contains(fmt.Sprint(err), "unsupported shape") {
						t.Fatalf("validation panic = %v, want unsupported shape", err)
					}
				}()
				validateGoALLCLowering(op, op, lowering, test.in, test.out, test.mask, test.imm)
			})
		}
	}
}

func TestGoALLCShiftDescriptor(t *testing.T) {
	savedArch := *FlagArch
	defer func() { *FlagArch = savedArch }()
	for _, test := range []struct {
		arch, feature, profile string
	}{
		{"amd64", "AVX", "x86.avx"},
		{"amd64", "AVX2", "x86.avx2"},
		{"amd64", "AVX512F", "x86.avx512"},
		{"amd64", "AVX512BW", "x86.avx512"},
		{"arm64", "NEON", ""},
	} {
		t.Run(test.arch+"/"+test.feature, func(t *testing.T) {
			*FlagArch = test.arch
			op := goALLCShiftTestOperation("shift-all-right", "Int32x4", goALLCShiftTestNativeScalar("Int8x16", 64))
			op.CPUFeature = test.feature
			d := goALLCSIMDDescriptor(op, op, PureVregIn, OneVregOut, NoMask, NoImm)
			if d.Lowering != "shift-all-right" || d.Lane != "int" || d.LaneBits != 32 ||
				d.ResultLane != "" || d.ResultLaneBits != 0 {
				t.Fatalf("unexpected shift descriptor: %+v", d)
			}
			if len(d.Arch) != 1 || d.Arch[test.arch].CPUProfile != test.profile || d.Arch[test.arch].OperandOrder != "" {
				t.Fatalf("unexpected shift architecture descriptor: %+v", d.Arch)
			}
		})
	}
}
