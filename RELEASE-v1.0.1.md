# RASON Meta Link Fix v1.0.1

## English

This release packages the Meta Horizon Link / NVIDIA freeze workaround that was validated on an affected Quest 3 + NVIDIA PC-VR setup.

### What the fix does

- Sets `HKCU\Software\Oculus\RemoteHeadset\NumSlices = 1`
- This disables Meta Link **Sliced Encoding**
- Restarts Meta's `OVRService` so the setting is applied
- Backs up the previous `NumSlices` state and can restore it

### What v1.0.1 improves

- Fixes the app itself freezing / showing "Not responding" during service operations
- Moves service and status work off the UI thread
- Adds hard command timeouts
- Adds clean close behavior
- Adds a cleaner application icon and Windows metadata
- Runs elevated via standard UAC
- Automatic German/English UI based on Windows language
- Keeps the actual Meta Link workaround unchanged

### Important

This is an **unofficial community workaround** and is not affiliated with Meta or NVIDIA. It does not patch NVIDIA drivers or Meta binaries and it does not simply hide Meta's driver warning.

On the affected system used for validation, frequent Meta Link freezes and stutter stopped after applying the workaround while keeping the current NVIDIA driver. This does **not** guarantee that it fixes every Link issue on every PC.

Use at your own risk.

### Downloads

Attach these files to the GitHub Release:

- `RASON-Meta-Link-Fix-v1.0.1.exe`
- `RASON-Meta-Link-Fix-v1.0.1.zip`

Verify the checksums in `SHA256.txt`.

---

## Deutsch

Diese Version enthält den Meta-Horizon-Link-/NVIDIA-Freeze-Workaround, der auf einem tatsächlich betroffenen Quest-3-/NVIDIA-PC-VR-System erfolgreich getestet wurde.

### Was der Fix macht

- Setzt `HKCU\Software\Oculus\RemoteHeadset\NumSlices = 1`
- Dadurch wird **Sliced Encoding** in Meta Link deaktiviert
- Startet Metas `OVRService` neu, damit die Einstellung übernommen wird
- Sichert den vorherigen `NumSlices`-Zustand und kann ihn wiederherstellen

### Was v1.0.1 verbessert

- Behebt das Einfrieren / "Keine Rückmeldung" der Fix-App bei Service-Aktionen
- Service- und Statusarbeiten laufen nicht mehr im UI-Thread
- Harte Timeouts für Hilfsprozesse
- Sauberes Verhalten beim Schließen
- Seriöseres App-Icon und Windows-Metadaten
- Startet über normales Windows-UAC mit Administratorrechten
- Automatische deutsche/englische Oberfläche passend zur Windows-Sprache
- Der eigentliche Meta-Link-Workaround bleibt unverändert

### Wichtig

Dies ist ein **inoffizieller Community-Workaround** und steht in keiner Verbindung zu Meta oder NVIDIA. Das Tool patcht keine NVIDIA-Treiber oder Meta-Binaries und blendet Metas Treiberwarnung nicht einfach aus.

Auf dem betroffenen Testsystem verschwanden die zuvor häufigen Meta-Link-Freezes und Stutter nach Anwendung des Workarounds, obwohl der aktuelle NVIDIA-Treiber installiert blieb. Das ist **keine Garantie** für jedes System.

Nutzung auf eigene Gefahr.

### Downloads

An den GitHub-Release anhängen:

- `RASON-Meta-Link-Fix-v1.0.1.exe`
- `RASON-Meta-Link-Fix-v1.0.1.zip`

Prüfsummen stehen in `SHA256.txt`.
