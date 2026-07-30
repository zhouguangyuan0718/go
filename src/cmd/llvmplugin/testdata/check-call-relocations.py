#!/usr/bin/env python3

import argparse
import struct
import subprocess
import sys


BLOCK_COUNT = 19
SYMBOL_SIZE = 21
RELOC_SIZE = 23
R_CALLIND = 10


def fail(message):
    raise RuntimeError(message)


def string_at(data, offset):
    size, string_offset = struct.unpack_from("<II", data, offset)
    return data[string_offset : string_offset + size].decode()


def read_symbols(data, start, end):
    return [
        string_at(data, offset)
        for offset in range(start, end, SYMBOL_SIZE)
    ]


def read_object(path):
    raw = open(path, "rb").read()
    magic = bytes([0]) + b"go120ld"
    base = raw.index(magic)
    data = raw[base:]
    offsets = struct.unpack_from(f"<{BLOCK_COUNT}I", data, 20)

    symbols = []
    for block in range(3, 7):
        symbols.extend(read_symbols(data, offsets[block], offsets[block + 1]))
    reloc_indexes = [
        struct.unpack_from("<I", data, offsets[11] + 4 * index)[0]
        for index in range((offsets[12] - offsets[11]) // 4)
    ]

    result = {}
    for symbol_index, name in enumerate(symbols):
        relocs = []
        for reloc_index in range(
            reloc_indexes[symbol_index], reloc_indexes[symbol_index + 1]
        ):
            offset = offsets[14] + reloc_index * RELOC_SIZE
            reloc_offset = struct.unpack_from("<i", data, offset)[0]
            size = data[offset + 4]
            reloc_type = struct.unpack_from("<H", data, offset + 5)[0]
            addend = struct.unpack_from("<q", data, offset + 7)[0]
            pkg_index, target_index = struct.unpack_from("<II", data, offset + 15)
            relocs.append(
                {
                    "offset": reloc_offset,
                    "size": size,
                    "type": reloc_type,
                    "addend": addend,
                    "pkg": pkg_index,
                    "target": target_index,
                }
            )
        result[name] = relocs
    return result


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--llc", required=True)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--input", required=True)
    parser.add_argument("--triple", required=True)
    parser.add_argument("--indirect-offset", type=int, required=True)
    parser.add_argument("--direct-reloc", type=int, required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    subprocess.run(
        [
            args.llc,
            f"-mtriple={args.triple}",
            f"-load-pass-plugin={args.plugin}",
            "-verify-machineinstrs",
            "-filetype=obj",
            "-o",
            args.output,
            args.input,
        ],
        check=True,
    )
    relocations = read_object(args.output)

    expected_marker = {
        "offset": args.indirect_offset,
        "size": 0,
        "type": R_CALLIND,
        "addend": 0,
        "pkg": 0,
        "target": 0,
    }
    for function_name in ("indirect_callee", "memory_indirect_callee"):
        indirect = relocations.get(function_name)
        if indirect is None:
            fail(f"missing {function_name} symbol")
        markers = [reloc for reloc in indirect if reloc["type"] == R_CALLIND]
        if markers != [expected_marker]:
            fail(f"{function_name} has unexpected R_CALLIND markers: {markers}")
        direct = [
            reloc for reloc in indirect if reloc["type"] == args.direct_reloc
        ]
        if len(direct) != 1:
            fail(
                f"{function_name} has {len(direct)} direct call relocations, "
                "want 1"
            )

    direct_function = relocations.get("call_only_pointer_argument")
    if direct_function is None:
        fail("missing call_only_pointer_argument symbol")
    if any(reloc["type"] == R_CALLIND for reloc in direct_function):
        fail("direct call function has an R_CALLIND marker")
    direct_calls = [
        reloc for reloc in direct_function if reloc["type"] == args.direct_reloc
    ]
    if len(direct_calls) != 2:
        fail(
            "direct call function has "
            f"{len(direct_calls)} direct call relocations, want 2"
        )

    print(
        f"{args.triple}: register and memory-loaded calls have R_CALLIND "
        f"offset={expected_marker['offset']} size=0 target=0:0; "
        f"direct calls remain type={args.direct_reloc}"
    )


if __name__ == "__main__":
    try:
        main()
    except (OSError, RuntimeError, subprocess.CalledProcessError, ValueError) as error:
        print(f"check-call-relocations: {error}", file=sys.stderr)
        sys.exit(1)
