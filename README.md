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

Use the **Releases** section on this repository and download the latest version.

For v1.0.1:
- `RASON-Meta-Link-Fix-v1.0.1.exe`
- `RASON-Meta-Link-Fix-v1.0.1.zip`
- `SHA256.txt`

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
