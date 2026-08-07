#!/usr/bin/env python3

import argparse
import re
import subprocess
import sys


def fail(message):
    print(f"defer result liveness check failed: {message}", file=sys.stderr)
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
    if '"go.defer.result.live"(ptr %result)' not in optimized.stdout:
        fail("O2 did not preserve the named defer result marker")

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
    if "go.defer.result.live" in ir:
        fail("statepoint rewrite left the frontend marker behind")
    if len(re.findall(r'"gc-live"\(ptr ', ir)) != 2:
        fail("the two calls after the result definition are not both GC-live")
    if len(re.findall(r"%result\.relocated", ir)) < 2:
        fail("missing relocation chain for the named defer result")


if __name__ == "__main__":
    main()
