#!/usr/bin/env python3
"""Embed a Windows icon and application manifest directly into a PE32+ executable.

This keeps the Go build dependency-free while giving the final EXE a normal Explorer icon
and an explicit requireAdministrator manifest. The script only appends a .rsrc section and
updates the PE resource directory/section metadata.
"""
from __future__ import annotations

import argparse
import struct
from pathlib import Path

RT_ICON = 3
RT_GROUP_ICON = 14
RT_VERSION = 16
RT_MANIFEST = 24
LANG_EN_US = 0x0409


def align(v: int, a: int) -> int:
    return (v + a - 1) & ~(a - 1)


def parse_ico(data: bytes):
    if len(data) < 6:
        raise ValueError("ICO is too small")
    reserved, typ, count = struct.unpack_from("<HHH", data, 0)
    if reserved != 0 or typ != 1 or count < 1:
        raise ValueError("Invalid ICO header")
    entries = []
    off = 6
    for i in range(count):
        if off + 16 > len(data):
            raise ValueError("Truncated ICO directory")
        width, height, colors, reserved2, planes, bitcount, size, image_off = struct.unpack_from("<BBBBHHII", data, off)
        if image_off + size > len(data):
            raise ValueError("ICO image points beyond end of file")
        entries.append({
            "width": width,
            "height": height,
            "colors": colors,
            "reserved": reserved2,
            "planes": planes,
            "bitcount": bitcount,
            "size": size,
            "data": data[image_off:image_off + size],
        })
        off += 16
    return entries


def group_icon_blob(entries, first_icon_id: int = 1) -> bytes:
    out = bytearray(struct.pack("<HHH", 0, 1, len(entries)))
    for idx, ent in enumerate(entries):
        out += struct.pack(
            "<BBBBHHIH",
            ent["width"], ent["height"], ent["colors"], ent["reserved"],
            ent["planes"], ent["bitcount"], ent["size"], first_icon_id + idx,
        )
    return bytes(out)


def _wstr(s: str) -> bytes:
    return (s + "\0").encode("utf-16le")


def _pad4(buf: bytearray):
    while len(buf) % 4:
        buf.append(0)


def _version_block(key: str, value: bytes, value_length: int, value_type: int, children: list[bytes] | None = None) -> bytes:
    children = children or []
    b = bytearray(b"\0\0\0\0\0\0")
    b += _wstr(key)
    _pad4(b)
    b += value
    _pad4(b)
    for child in children:
        b += child
        _pad4(b)
    struct.pack_into("<HHH", b, 0, len(b), value_length, value_type)
    return bytes(b)


def version_info_blob() -> bytes:
    # VS_FIXEDFILEINFO for version 1.0.1.0.
    fixed = struct.pack(
        "<13I",
        0xFEEF04BD, 0x00010000,
        0x00010000, 0x00010000,
        0x00010000, 0x00010000,
        0x0000003F, 0x00000000,
        0x00040004, 0x00000001, 0x00000000,
        0x00000000, 0x00000000,
    )
    strings = {
        "CompanyName": "RASON",
        "FileDescription": "RASON Meta Link Fix",
        "FileVersion": "1.0.1.0",
        "InternalName": "RASON Meta Link Fix",
        "OriginalFilename": "RASON-Meta-Link-Fix.exe",
        "ProductName": "RASON Meta Link Fix",
        "ProductVersion": "1.0.1.0",
        "Comments": "Unofficial community workaround for Meta Horizon Link NVIDIA freezes",
        "LegalCopyright": "© 2026 RASON",
    }
    string_children = []
    for key, text in strings.items():
        value = _wstr(text)
        string_children.append(_version_block(key, value, len(text) + 1, 1))
    table = _version_block("040904B0", b"", 0, 1, string_children)
    string_file = _version_block("StringFileInfo", b"", 0, 1, [table])
    translation = struct.pack("<HH", LANG_EN_US, 1200)
    var = _version_block("Translation", translation, len(translation), 0)
    var_file = _version_block("VarFileInfo", b"", 0, 1, [var])
    return _version_block("VS_VERSION_INFO", fixed, len(fixed), 0, [string_file, var_file])


def build_resource_blob(section_rva: int, ico_data: bytes, manifest_data: bytes) -> bytes:
    icons = parse_ico(ico_data)
    # Resource leaves: (type, id, language, bytes)
    leaves = []
    for i, ent in enumerate(icons, 1):
        leaves.append((RT_ICON, i, LANG_EN_US, ent["data"]))
    leaves.append((RT_GROUP_ICON, 1, LANG_EN_US, group_icon_blob(icons, 1)))
    leaves.append((RT_VERSION, 1, LANG_EN_US, version_info_blob()))
    leaves.append((RT_MANIFEST, 1, LANG_EN_US, manifest_data))

    # Group by type, then ID. All entries are numeric IDs.
    by_type: dict[int, dict[int, tuple[int, bytes]]] = {}
    for typ, rid, lang, payload in leaves:
        by_type.setdefault(typ, {})[rid] = (lang, payload)

    # First allocate all directory structures.
    buf = bytearray()

    def alloc(n: int, alignment: int = 4) -> int:
        pos = align(len(buf), alignment)
        if pos > len(buf):
            buf.extend(b"\0" * (pos - len(buf)))
        buf.extend(b"\0" * n)
        return pos

    def write_dir_header(pos: int, num_ids: int):
        struct.pack_into("<IIHHHH", buf, pos, 0, 0, 0, 0, 0, num_ids)

    root_types = sorted(by_type)
    root_pos = alloc(16 + 8 * len(root_types), 4)
    write_dir_header(root_pos, len(root_types))

    type_dirs = {}
    id_dirs = {}
    data_entry_positions = {}

    for typ in root_types:
        ids = sorted(by_type[typ])
        tpos = alloc(16 + 8 * len(ids), 4)
        write_dir_header(tpos, len(ids))
        type_dirs[typ] = tpos
        for rid in ids:
            ipos = alloc(16 + 8, 4)  # one language entry per ID
            write_dir_header(ipos, 1)
            id_dirs[(typ, rid)] = ipos
            dpos = alloc(16, 4)
            data_entry_positions[(typ, rid)] = dpos

    # Fill directory entries.
    for i, typ in enumerate(root_types):
        struct.pack_into("<II", buf, root_pos + 16 + 8 * i, typ, 0x80000000 | type_dirs[typ])
        ids = sorted(by_type[typ])
        tpos = type_dirs[typ]
        for j, rid in enumerate(ids):
            struct.pack_into("<II", buf, tpos + 16 + 8 * j, rid, 0x80000000 | id_dirs[(typ, rid)])
            lang, _payload = by_type[typ][rid]
            ipos = id_dirs[(typ, rid)]
            struct.pack_into("<II", buf, ipos + 16, lang, data_entry_positions[(typ, rid)])

    # Append resource payloads and fill IMAGE_RESOURCE_DATA_ENTRY records.
    for typ in root_types:
        for rid in sorted(by_type[typ]):
            _lang, payload = by_type[typ][rid]
            payload_pos = align(len(buf), 4)
            if payload_pos > len(buf):
                buf.extend(b"\0" * (payload_pos - len(buf)))
            buf.extend(payload)
            dpos = data_entry_positions[(typ, rid)]
            struct.pack_into("<IIII", buf, dpos, section_rva + payload_pos, len(payload), 0, 0)

    return bytes(buf)


def embed(exe_path: Path, ico_path: Path, manifest_path: Path):
    pe = bytearray(exe_path.read_bytes())
    ico = ico_path.read_bytes()
    manifest = manifest_path.read_bytes()

    if pe[:2] != b"MZ":
        raise ValueError("Not a PE executable (missing MZ)")
    pe_off = struct.unpack_from("<I", pe, 0x3C)[0]
    if pe[pe_off:pe_off + 4] != b"PE\0\0":
        raise ValueError("Not a PE executable (missing PE signature)")
    file_hdr = pe_off + 4
    num_sections = struct.unpack_from("<H", pe, file_hdr + 2)[0]
    size_opt = struct.unpack_from("<H", pe, file_hdr + 16)[0]
    opt = file_hdr + 20
    magic = struct.unpack_from("<H", pe, opt)[0]
    if magic != 0x20B:
        raise ValueError(f"Expected PE32+ optional header, got 0x{magic:04x}")

    section_alignment = struct.unpack_from("<I", pe, opt + 32)[0]
    file_alignment = struct.unpack_from("<I", pe, opt + 36)[0]
    size_headers = struct.unpack_from("<I", pe, opt + 60)[0]
    number_rva_sizes = struct.unpack_from("<I", pe, opt + 108)[0]
    if number_rva_sizes < 3:
        raise ValueError("PE optional header has no resource data directory")
    data_dir = opt + 112
    section_table = opt + size_opt

    # Do not double-embed.
    old_rsrc_rva, old_rsrc_size = struct.unpack_from("<II", pe, data_dir + 2 * 8)
    if old_rsrc_rva or old_rsrc_size:
        raise ValueError("Executable already has a resource directory")

    sections = []
    for i in range(num_sections):
        sh = section_table + i * 40
        name = bytes(pe[sh:sh + 8]).split(b"\0", 1)[0].decode("ascii", "replace")
        vsize, va, raw_size, raw_ptr = struct.unpack_from("<IIII", pe, sh + 8)
        sections.append((name, vsize, va, raw_size, raw_ptr))

    # Ensure there is room for one more 40-byte section header inside SizeOfHeaders.
    new_sh = section_table + num_sections * 40
    if new_sh + 40 > size_headers:
        raise ValueError("No room for an additional PE section header")

    last_va_end = max(va + max(vsize, raw_size) for _, vsize, va, raw_size, _ in sections)
    new_va = align(last_va_end, section_alignment)
    blob = build_resource_blob(new_va, ico, manifest)
    raw_size = align(len(blob), file_alignment)
    raw_ptr = align(len(pe), file_alignment)

    # Preserve any overlay by padding then appending; Go builds normally have none.
    if raw_ptr > len(pe):
        pe.extend(b"\0" * (raw_ptr - len(pe)))
    pe.extend(blob)
    if len(blob) < raw_size:
        pe.extend(b"\0" * (raw_size - len(blob)))

    # New section header.
    name = b".rsrc\0\0\0"
    pe[new_sh:new_sh + 8] = name
    struct.pack_into("<IIIIIIHHI", pe, new_sh + 8,
                     len(blob), new_va, raw_size, raw_ptr,
                     0, 0, 0, 0, 0x40000040)

    # File header / optional header updates.
    struct.pack_into("<H", pe, file_hdr + 2, num_sections + 1)
    old_init = struct.unpack_from("<I", pe, opt + 8)[0]
    struct.pack_into("<I", pe, opt + 8, old_init + raw_size)
    struct.pack_into("<I", pe, opt + 56, align(new_va + len(blob), section_alignment))  # SizeOfImage
    struct.pack_into("<II", pe, data_dir + 2 * 8, new_va, len(blob))
    # Clear checksum; unsigned Go executables generally have 0 here already.
    struct.pack_into("<I", pe, opt + 64, 0)

    exe_path.write_bytes(pe)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("exe", type=Path)
    ap.add_argument("--icon", required=True, type=Path)
    ap.add_argument("--manifest", required=True, type=Path)
    args = ap.parse_args()
    embed(args.exe, args.icon, args.manifest)
    print(f"Embedded icon + manifest into {args.exe}")


if __name__ == "__main__":
    main()
