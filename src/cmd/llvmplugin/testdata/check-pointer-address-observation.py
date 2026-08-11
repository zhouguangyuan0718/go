#!/usr/bin/env python3

import argparse
import re
import subprocess
import sys


def fail(message):
    print(f"pointer-address observation check failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--opt", required=True)
    parser.add_argument("--llc", required=True)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--input", required=True)
    args = parser.parse_args()

    optimized = subprocess.run(
        [args.opt, "-S", "-passes=default<O2>", args.input, "-o", "-"],
        capture_output=True,
        text=True,
    )
    if optimized.returncode != 0:
        fail(f"opt failed:\n{optimized.stdout}{optimized.stderr}")
    if optimized.stdout.count("call i64 @llvm.go.pointer.address.i64.p0") != 2:
        fail("O2 did not preserve both physical-address observations")
    if "ret i1 true" in optimized.stdout:
        fail("O2 folded the address comparison")

    rewritten = subprocess.run(
        [
            args.llc,
            f"-load-pass-plugin={args.plugin}",
            "-goallc-pass-plugin-emit-ir",
            "-filetype=null",
            "-o",
            "-",
            "-",
        ],
        input=optimized.stdout,
        capture_output=True,
        text=True,
    )
    if rewritten.returncode != 0:
        fail(f"llc failed:\n{rewritten.stdout}{rewritten.stderr}")
    ir = rewritten.stdout + rewritten.stderr
    if "llvm.go.pointer.address" in ir:
        fail("statepoint rewrite left a pointer-address observation behind")
    patterns = [
        r"%before\.lowered = ptrtoint ptr %pointer to i64",
        r"%pointer\.relocated = call coldcc ptr @llvm\.experimental\.gc\.relocate",
        r"%after\.lowered = ptrtoint ptr %pointer\.relocated to i64",
        r"%same = icmp eq i64 %before\.lowered, %after\.lowered",
    ]
    for pattern in patterns:
        if not re.search(pattern, ir):
            fail(f"rewritten IR does not match {pattern!r}")


if __name__ == "__main__":
    main()
