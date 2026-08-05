#!/usr/bin/env python3

import argparse
import re
import subprocess
import sys


def fail(message):
    print(f"function marker inline check failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def function_body(ir, name):
    match = re.search(
        rf"define [^@]*@{re.escape(name)}\([^{{]*\) [^{{]*\{{\n(.*?)\n\}}",
        ir,
        re.DOTALL,
    )
    if not match:
        fail(f"missing function {name}")
    return match.group(1)


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
    if re.search(r"define [^@]*@callee\(", optimized.stdout):
        fail("O2 did not remove the fully inlined callee")
    body = function_body(optimized.stdout, "caller")
    if not re.search(r"call void @llvm\.sideeffect\(\), !goobj\.marker_reloc", body):
        fail("O2 did not clone the marker into caller")

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
    if re.search(r"call void @llvm\.sideeffect", ir):
        fail("statepoint rewrite left a function marker intrinsic behind")
    if re.search(r"!\{ptr @callee, ptr @target, i32 24, i64 96\}", ir):
        fail("materialized a marker relocation for the deleted callee")
    pattern = r"!\{ptr @caller, ptr @target, i32 24, i64 96\}"
    if not re.search(pattern, ir):
        fail("missing materialized marker relocation for caller")


if __name__ == "__main__":
    main()
