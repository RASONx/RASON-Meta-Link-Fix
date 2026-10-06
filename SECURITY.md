# Security

## Reporting a security issue

Please do not include passwords, tokens, crash dumps containing personal information, or other secrets in a public issue.

If you believe the application performs an unexpected registry, process, service, file, or network operation, open a GitHub issue with the exact application version and steps to reproduce.

## Scope

The intended system changes are deliberately narrow:

- `HKCU\Software\Oculus\RemoteHeadset\NumSlices`
- the application's own backup/status data under `HKCU\Software\RASON\MetaLinkFix`
- controlled restart of Meta's `OVRService`

The application is not intended to modify NVIDIA or Meta program binaries and contains no telemetry or updater.
