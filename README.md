# RASON Meta Link Fix

[English](#rason-meta-link-fix) · [Deutsch](README.de.md)

A small, transparent Windows utility for a specific Meta Horizon Link / NVIDIA PC-VR freeze and stutter workaround.

> **Unofficial community tool.** This project is not affiliated with, endorsed by, or supported by Meta or NVIDIA.

## What it does

The fix applies one targeted Meta Link configuration change:

`HKCU\Software\Oculus\RemoteHeadset\NumSlices = 1`

This disables Meta Link **Sliced Encoding** and then restarts `OVRService` in a controlled way so the setting is picked up.

On the affected Quest 3 + NVIDIA system used to validate this project, frequent Link freezes and stutter stopped after applying this workaround while keeping the current NVIDIA driver installed.

That is a real-world validation of this workaround on one affected system, **not a guarantee that it will solve every Meta Link issue on every PC**.

## Scope: stability workaround, not Air Link tuning

This project is intended as a **stability workaround for affected Meta Link users**, specifically for systems experiencing freezes or severe stutter with newer NVIDIA drivers.

It is **not** an Air Link bitrate, latency, or image-quality optimization. Disabling Sliced Encoding is a trade-off: on some Air Link setups, leaving Sliced Encoding enabled may be preferable for performance or bitrate behavior.

If your Air Link setup works better with Sliced Encoding enabled, do not treat this tool as an upgrade. Use **Restore Original** and keep the configuration that works best for your system.

## Manual alternative / why an app?

You can also disable Sliced Encoding manually in **Oculus Debug Tool (ODT)** or by setting the same registry value yourself. This project does not claim to contain a secret driver patch.

The utility exists as a small reversible wrapper around that workaround: it backs up the previous value, applies and verifies the setting, can restore the original state, restarts `OVRService`, and clearly shows whether the workaround is active.

## Source code

The source for the published v1.0.1 application is available directly in this repository:

- [`src/main.go`](src/main.go) — Windows application, UI, backup/restore and Meta Link service handling
- [`embed_windows_resources.py`](embed_windows_resources.py) — PE icon/manifest/version resource embedding
- [`app.manifest`](app.manifest) — Windows UAC/DPI manifest
- [`go.mod`](go.mod) — Go module definition

These text source files were copied from the v1.0.1 release project and verified against their Git blob hashes. The downloadable v1.0.1 ZIP also contains the branding assets used for the published executable.

## What it does NOT do

- It does **not** patch or replace NVIDIA drivers.
- It does **not** modify Meta/NVIDIA DLLs or signed binaries.
- It does **not** hide the Meta warning banner as a fake “fix”.
- It does **not** apply unrelated Windows “optimization” tweaks.
- It does **not** install telemetry, an updater, or network components.

The app also stores only its own backup/status information under:

`HKCU\Software\RASON\MetaLinkFix`

so the original setting can be restored.

## Features

- Windows x64
- Runs elevated via normal Windows UAC
- Automatic UI language: German on German Windows, English otherwise
- Manual DE / EN switch
- Clear **FIX ACTIVE / INACTIVE** status
- Apply / restore workflow
- Controlled `OVRService` restart
- NVIDIA driver / Meta Link status diagnostics where available
- Background operations so the UI stays responsive
- Transparent source code
- No telemetry

## Download

**Latest release: [RASON Meta Link Fix v1.0.1](https://github.com/RASONx/RASON-Meta-Link-Fix/releases/tag/v1.0.1)**

Direct downloads:
- [RASON-Meta-Link-Fix-v1.0.1.exe](https://github.com/RASONx/RASON-Meta-Link-Fix/releases/download/v1.0.1/RASON-Meta-Link-Fix-v1.0.1.exe)
- [RASON-Meta-Link-Fix-v1.0.1.zip](https://github.com/RASONx/RASON-Meta-Link-Fix/releases/download/v1.0.1/RASON-Meta-Link-Fix-v1.0.1.zip)
- Checksums: [`SHA256.txt`](SHA256.txt)

## Usage

1. Close an active Meta Link session.
2. Start `RASON-Meta-Link-Fix.exe`.
3. Approve the Windows UAC prompt.
4. Read and accept the startup notice.
5. Click **Apply Fix**.
6. Let the app restart `OVRService`.
7. Start Meta Horizon Link again and test your usual PC-VR workload.

If the workaround does not help your system, use **Restore Original**.

## Why administrator rights?

Writing the user-level registry value itself does not require administrator rights, but restarting the Meta `OVRService` reliably does. The executable therefore requests elevation through a standard Windows manifest.

## Safety / risk notice

Use this utility at your own risk. It changes a documented Windows registry location used by Meta Link and restarts a Meta service. The app is designed to be reversible, but neither the author nor contributors can guarantee compatibility with every Meta Horizon Link, Windows, NVIDIA driver, or Quest version.

Before running any executable downloaded from the internet, verify that you downloaded it from this repository's official Releases page and compare its SHA-256 checksum.

## Build from source

Requirements:

- Go 1.23+
- Python 3
- Windows target build support

The project intentionally avoids third-party Go dependencies. See `BUILD.md` for the exact build process.

## Version

Current release: **v1.0.1**

v1.0.1 keeps the validated `NumSlices=1` workaround unchanged and improves the application shell: responsive background operations, cleaner branding, UAC manifest, bilingual UI, diagnostics, and safer close behavior.

## License

MIT. See [LICENSE](LICENSE).

---

### Deutsch

Eine vollständige deutsche Beschreibung und Anleitung findest du in **[README.de.md](README.de.md)**.
