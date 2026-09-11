// runoutput

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Many distinct pointer-containing stack homes can give a single statepoint
// more than 65535 SelectionDAG operands. In particular, each constant in the
// alloca layout protocol lowers to both a stack-map tag and its value.
package main

import "fmt"

func main() {
	const count = 2850
	fmt.Print(`package main

import "runtime"

//go:noinline
func seed() string { return string([]byte("abcdefghij")) }

//go:noinline
func check(values []*string) {
	runtime.GC()
	for i, p := range values {
		if len(*p) != 1+i%10 || (*p)[0] != 'a' {
			panic("lost stack-home contents")
		}
		// The homes must remain distinct, including ones with equal values.
		*p = ""
	}
}

func main() {
	s := seed()
`)
	for i := 0; i < count; i++ {
		fmt.Printf("v%d := s[:%d]\n", i, 1+i%10)
	}
	fmt.Print("check([]*string{\n")
	for i := 0; i < count; i++ {
		fmt.Printf("&v%d,\n", i)
	}
	fmt.Print("})\n}\n")
}
