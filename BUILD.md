# Build

RASON Meta Link Fix v1.0.1

## Source layout

- `src/main.go` — application source
- `embed_windows_resources.py` — Windows PE resource embedder
- `app.manifest` — UAC / Common Controls / DPI manifest
- `go.mod` — Go module definition
- The exact `.ico` / `.png` branding assets used by v1.0.1 are included in the downloadable v1.0.1 release ZIP.

## Environment

- Go 1.23.x
- Target: `windows/amd64`
- CGO disabled
- Python 3 for PE resource embedding

## Commands

```bash
gofmt -w src/main.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet ./src

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w -H=windowsgui" \
  -o RASON-Meta-Link-Fix.exe ./src

# For the exact published branding, extract assets/RASON-Meta-Link-Fix.ico from the v1.0.1 release ZIP first.
python embed_windows_resources.py RASON-Meta-Link-Fix.exe \
  --icon assets/RASON-Meta-Link-Fix.ico \
  --manifest app.manifest
```

## Static checks performed for v1.0.1

- Successful Windows PE32+ x86-64 GUI build.
- `go vet` succeeded for the Windows target.
- PE resource directory contains RT_ICON, RT_GROUP_ICON, RT_VERSION and RT_MANIFEST.
- Embedded manifest contains `requestedExecutionLevel=requireAdministrator`.
- Embedded file/product version is `1.0.1.0`.
- Source audit confirms the Meta Link workaround writes only `NumSlices` under `HKCU\Software\Oculus\RemoteHeadset`; additional registry writes are limited to the application's own backup/status key under `HKCU\Software\RASON\MetaLinkFix`.
- No `net/http`, socket, telemetry, updater or download code is used by the application.
- Apply / Restore / OVRService restart execute in background goroutines rather than in the Win32 message/UI thread.
- `nvidia-smi` and `sc.exe` helper invocations use explicit timeouts.
- A second clean build with the same inputs was byte-identical.

## Runtime limitation

The build environment used for the release could not attach a physical Quest headset or execute the native Windows GUI. The underlying `NumSlices=1` workaround was separately validated on an affected Quest 3 / NVIDIA Meta Link setup. v1.0.1 changes only the application shell, responsiveness and branding around that workaround.
