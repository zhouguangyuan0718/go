// asmcheck

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// LLVM-DAG: define goabiinternal void @codegen.store_shifted(ptr %t, i32 %x)
// LLVM-DAG: lshr i32 %x, 8
// LLVM-DAG: lshr i32 %x, 16
// LLVM-DAG: store i16 {{%.*}}, ptr {{%.*}}, align 2
// LLVM-DAG: define goabiinternal void @codegen.store_const(ptr %t)
// LLVM-DAG: store i8 1, ptr {{%.*}}, align 1
// LLVM-DAG: store i8 2, ptr {{%.*}}, align 1
// LLVM-DAG: store i16 3, ptr {{%.*}}, align 2

type tile1 struct {
	a uint16
	b uint16
	c uint32
}

func store_tile1(t *tile1) {
	// amd64:`MOVQ`
	t.a, t.b, t.c = 1, 1, 1
}

type tile2 struct {
	a, b, c, d, e int8
}

func store_tile2(t *tile2) {
	// amd64:`MOVW`
	t.a, t.b = 1, 1
	// amd64:`MOVW`
	t.d, t.e = 1, 1
}

type tile3 struct {
	a, b uint8
	c    uint16
}

func store_shifted(t *tile3, x uint32) {
	// amd64:`MOVL`
	// ppc64:`MOVHBR`
	t.a = uint8(x)
	t.b = uint8(x >> 8)
	t.c = uint16(x >> 16)
}

func store_const(t *tile3) {
	// 0x00030201
	// amd64:`MOVL \$197121`
	// 0x01020003
	// ppc64:`MOVD \$16908291`
	t.a, t.b, t.c = 1, 2, 3
}
