#!/usr/bin/env python3

import argparse
import struct
import subprocess
import sys


BLOCK_COUNT = 19
SYMBOL_SIZE = 21
RELOC_SIZE = 23
R_CALLIND = 10
BLK_DATA_IDX = 13
BLK_DATA = 16


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
        data_index_offset = offsets[BLK_DATA_IDX] + 4 * symbol_index
        data_start, data_end = struct.unpack_from(
            "<II", data, data_index_offset
        )
        result[name] = {
            "data": data[
                offsets[BLK_DATA] + data_start : offsets[BLK_DATA] + data_end
            ],
            "relocations": relocs,
        }
    return result


def is_x86_indirect_call(function_data, offset):
    # Skip the instruction prefixes that may precede the FF /2 opcode. The
    # relocation still points at the beginning of the complete instruction.
    index = offset
    legacy_prefixes = {
        0x26,
        0x2E,
        0x36,
        0x3E,
        0x64,
        0x65,
        0x66,
        0x67,
        0xF2,
        0xF3,
    }
    while index < len(function_data) and index - offset < 15:
        byte = function_data[index]
        if byte in legacy_prefixes or 0x40 <= byte <= 0x4F:
            index += 1
            continue
        break
    if index + 1 >= len(function_data) or function_data[index] != 0xFF:
        return False
    modrm = function_data[index + 1]
    return modrm & 0x38 == 0x10


def is_aarch64_indirect_call(function_data, offset):
    if offset < 0 or offset + 4 > len(function_data) or offset % 4 != 0:
        return False
    instruction = struct.unpack_from("<I", function_data, offset)[0]
    # BLR Xn, ignoring the register field in bits [9:5].
    return instruction & 0xFFFFFC1F == 0xD63F0000


def is_indirect_call(function_data, offset, triple):
    if offset < 0 or offset >= len(function_data):
        return False
    if triple.startswith("x86_64-"):
        return is_x86_indirect_call(function_data, offset)
    if triple.startswith("aarch64-"):
        return is_aarch64_indirect_call(function_data, offset)
    fail(f"unsupported triple {triple}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--llc", required=True)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--input", required=True)
    parser.add_argument("--triple", required=True)
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
    symbols = read_object(args.output)

    expected_marker_fields = {
        "size": 0,
        "type": R_CALLIND,
        "addend": 0,
        "pkg": 0,
        "target": 0,
    }
    marker_offsets = []
    for function_name in ("indirect_callee", "memory_indirect_callee"):
        indirect = symbols.get(function_name)
        if indirect is None:
            fail(f"missing {function_name} symbol")
        relocations = indirect["relocations"]
        markers = [
            reloc for reloc in relocations if reloc["type"] == R_CALLIND
        ]
        if len(markers) != 1:
            fail(f"{function_name} has unexpected R_CALLIND markers: {markers}")
        marker = markers[0]
        marker_fields = {
            key: marker[key] for key in expected_marker_fields
        }
        if marker_fields != expected_marker_fields:
            fail(f"{function_name} has unexpected R_CALLIND marker: {marker}")
        if not is_indirect_call(
            indirect["data"], marker["offset"], args.triple
        ):
            code = indirect["data"][marker["offset"] : marker["offset"] + 8]
            fail(
                f"{function_name} R_CALLIND offset {marker['offset']} does not "
                f"point at an indirect call instruction: {code.hex()}"
            )
        marker_offsets.append(marker["offset"])
        direct = [
            reloc
            for reloc in relocations
            if reloc["type"] == args.direct_reloc
        ]
        if len(direct) != 1:
            fail(
                f"{function_name} has {len(direct)} direct call relocations, "
                "want 1"
            )

    direct_function = symbols.get("call_only_pointer_argument")
    if direct_function is None:
        fail("missing call_only_pointer_argument symbol")
    direct_relocations = direct_function["relocations"]
    if any(reloc["type"] == R_CALLIND for reloc in direct_relocations):
        fail("direct call function has an R_CALLIND marker")
    direct_calls = [
        reloc
        for reloc in direct_relocations
        if reloc["type"] == args.direct_reloc
    ]
    if len(direct_calls) != 2:
        fail(
            "direct call function has "
            f"{len(direct_calls)} direct call relocations, want 2"
        )

    print(
        f"{args.triple}: register and memory-loaded calls have R_CALLIND "
        f"offsets={marker_offsets} pointing at indirect call instructions, "
        "size=0 target=0:0; "
        f"direct calls remain type={args.direct_reloc}"
    )


if __name__ == "__main__":
    try:
        main()
    except (OSError, RuntimeError, subprocess.CalledProcessError, ValueError) as error:
        print(f"check-call-relocations: {error}", file=sys.stderr)
        sys.exit(1)
