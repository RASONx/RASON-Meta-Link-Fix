# Changelog

## v1.0.1

- Core Meta Link workaround unchanged: `NumSlices=1` / Sliced Encoding off.
- Fixed the UI appearing frozen / "Not responding" when Apply, Restore or OVRService actions are used: service work now runs off the Win32 message/UI thread.
- Status detection now runs asynchronously.
- Added hard timeouts around `nvidia-smi` and `sc.exe` calls so a stuck helper command cannot block the interface indefinitely.
- Closing during an active service operation now hides the window immediately and exits after the short critical service operation finishes safely.
- Replaced the shield-style/generic look with a cleaner VR/link application icon.
- Icon is embedded in the PE resource section and appears on the EXE in Windows Explorer.
- Added a `requireAdministrator` Windows manifest so UAC elevation happens at launch.
- Added Windows version metadata: RASON Meta Link Fix 1.0.1.0.
- Common Controls v6 + PerMonitorV2 DPI manifest settings.
- German/English auto-language, manual language toggle and startup safety notice retained.
- Existing backup format remains compatible and reversible.
- No additional VR/Windows tweaks introduced.

## v1.0.0

- Added the bilingual Windows UI around the already-tested `NumSlices` workaround.
- Added backup/restore, diagnostics and explicit disclosure of all changes.
