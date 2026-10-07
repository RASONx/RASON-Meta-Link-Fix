# RASON Meta Link Fix

[English](README.md) · **Deutsch**

Ein kleines, transparentes Windows-Tool für einen bestimmten **Meta Horizon Link / NVIDIA PC-VR Freeze- und Stutter-Workaround**.

> **Inoffizielles Community-Tool.** Dieses Projekt steht in keiner Verbindung zu Meta oder NVIDIA und wird von beiden Unternehmen weder unterstützt noch offiziell empfohlen.

## Was der Fix macht

Der Fix nimmt genau eine gezielte Meta-Link-Änderung vor:

`HKCU\Software\Oculus\RemoteHeadset\NumSlices = 1`

Damit wird **Sliced Encoding** in Meta Link deaktiviert. Anschließend startet die Anwendung den Meta-Dienst `OVRService` kontrolliert neu, damit die Einstellung übernommen wird.

Auf dem betroffenen Quest-3-/NVIDIA-System, auf dem dieses Projekt praktisch getestet wurde, verschwanden zuvor häufige Link-Freezes und Stutter nach Anwendung dieses Workarounds, während der aktuelle NVIDIA-Treiber installiert blieb.

Das ist ein erfolgreicher Praxistest auf einem tatsächlich betroffenen System, aber **keine Garantie dafür, dass damit jedes Meta-Link-Problem auf jedem PC gelöst wird**.

## Einsatzbereich: Stabilitäts-Workaround, kein Air-Link-Tuning

Dieses Projekt ist als **Stabilitäts-Workaround für betroffene Meta-Link-Nutzer** gedacht, insbesondere für Systeme mit Freezes oder starkem Stuttering unter neueren NVIDIA-Treibern.

Es ist **kein** Tuning für Air-Link-Bitrate, Latenz oder Bildqualität. Das Deaktivieren von Sliced Encoding ist ein Kompromiss: Bei manchen Air-Link-Setups kann es für Performance oder Bitratenverhalten sinnvoller sein, Sliced Encoding aktiviert zu lassen.

Wenn dein Air-Link-Setup mit aktiviertem Sliced Encoding besser funktioniert, ist dieses Tool kein Upgrade. Nutze **Original wiederherstellen** und behalte die Konfiguration, die auf deinem System am besten läuft.

## Manuelle Alternative / warum überhaupt eine App?

Sliced Encoding kann auch direkt im **Oculus Debug Tool (ODT)** deaktiviert oder derselbe Registry-Wert manuell gesetzt werden. Dieses Projekt behauptet ausdrücklich nicht, einen geheimen Treiber-Patch zu enthalten.

Die Anwendung ist ein kleiner reversibler Wrapper um diesen Workaround: Sie sichert den vorherigen Wert, setzt und überprüft die Einstellung, kann den Originalzustand wiederherstellen, startet `OVRService` neu und zeigt klar an, ob der Workaround aktiv ist.

## Quellcode

Der Quellcode der veröffentlichten v1.0.1-Anwendung liegt direkt in diesem Repository:

- [`src/main.go`](src/main.go) — Windows-App, Oberfläche, Backup/Restore und Meta-Link-Service-Steuerung
- [`embed_windows_resources.py`](embed_windows_resources.py) — Einbettung von Icon, Manifest und Versionsressourcen in die EXE
- [`app.manifest`](app.manifest) — Windows-UAC-/DPI-Manifest
- [`go.mod`](go.mod) — Go-Moduldefinition

Diese Text-Quelldateien wurden aus dem v1.0.1-Releaseprojekt übernommen und anhand ihrer Git-Blob-Hashes verifiziert. Das herunterladbare v1.0.1-ZIP enthält zusätzlich die Branding-Assets, die für die veröffentlichte EXE verwendet wurden.

## Was der Fix NICHT macht

- Er patcht oder ersetzt **keine NVIDIA-Treiber**.
- Er verändert **keine Meta- oder NVIDIA-DLLs** und keine signierten Binärdateien.
- Er blendet die Meta-Warnmeldung nicht einfach aus und nennt das dann „Fix“.
- Er setzt keine allgemeinen oder fragwürdigen Windows-„Optimierungen“.
- Er enthält keine Telemetrie, keinen Updater und keine Netzwerkkomponenten.

Die Anwendung speichert lediglich eigene Backup-/Statusinformationen unter:

`HKCU\Software\RASON\MetaLinkFix`

Dadurch kann der ursprüngliche Zustand wiederhergestellt werden.

## Funktionen

- Windows x64
- Startet über normales Windows-UAC mit Administratorrechten
- Automatische Sprache: Deutsch auf einem deutschen Windows, sonst Englisch
- Sprache zusätzlich manuell zwischen DE / EN umschaltbar
- Klare Anzeige **FIX AKTIV / INAKTIV**
- Fix anwenden / Original wiederherstellen
- Kontrollierter Neustart von `OVRService`
- Diagnose von NVIDIA-Treiber und Meta Link, soweit verfügbar
- Hintergrundoperationen, damit die Oberfläche nicht einfriert
- Offener Quellcode
- Keine Telemetrie

## Download

**Aktuelle Version: [RASON Meta Link Fix v1.0.1](https://github.com/RASONx/RASON-Meta-Link-Fix/releases/tag/v1.0.1)**

Direkte Downloads:
- [RASON-Meta-Link-Fix-v1.0.1.exe](https://github.com/RASONx/RASON-Meta-Link-Fix/releases/download/v1.0.1/RASON-Meta-Link-Fix-v1.0.1.exe)
- [RASON-Meta-Link-Fix-v1.0.1.zip](https://github.com/RASONx/RASON-Meta-Link-Fix/releases/download/v1.0.1/RASON-Meta-Link-Fix-v1.0.1.zip)
- Prüfsummen: [`SHA256.txt`](SHA256.txt)

## Verwendung

1. Eine laufende Meta-Link-Sitzung beenden.
2. `RASON-Meta-Link-Fix.exe` starten.
3. Die Windows-UAC-Abfrage bestätigen.
4. Den Hinweis beim Start lesen und bestätigen.
5. **Fix anwenden** anklicken.
6. Die Anwendung `OVRService` neu starten lassen.
7. Meta Horizon Link wieder öffnen und das gewohnte PC-VR-Spiel testen.

Falls der Workaround auf deinem System nicht hilft, kannst du mit **Original wiederherstellen** zum vorherigen Zustand zurückkehren.

## Warum Administratorrechte?

Der Registry-Wert selbst liegt im Benutzerbereich. Für einen zuverlässigen Neustart des Meta-Dienstes `OVRService` werden jedoch Administratorrechte benötigt. Deshalb fordert die EXE über ein normales Windows-Manifest UAC-Rechte an.

## Sicherheits- und Risikohinweis

Die Verwendung erfolgt auf eigene Gefahr. Das Tool verändert eine von Meta Link verwendete Registry-Einstellung und startet einen Meta-Dienst neu. Die Änderung ist bewusst reversibel aufgebaut, trotzdem kann keine vollständige Kompatibilität mit jeder Version von Meta Horizon Link, Windows, NVIDIA-Treibern oder Quest garantiert werden.

Lade die EXE ausschließlich von der offiziellen **Releases-Seite dieses Repositorys** herunter und vergleiche bei Bedarf den SHA-256-Hash.

## Aus dem Quellcode bauen

Benötigt werden:

- Go 1.23+
- Python 3
- Windows-x64 als Build-Ziel

Das Projekt verwendet absichtlich keine externen Go-Abhängigkeiten. Der genaue Ablauf steht in `BUILD.md`.

## Version

Aktuelle Version: **v1.0.1**

v1.0.1 verändert den bereits praktisch bestätigten `NumSlices=1`-Workaround nicht. Verbessert wurden die Anwendung selbst: keine blockierenden UI-Aktionen mehr, saubereres Branding, UAC-Manifest, zweisprachige Oberfläche, Diagnosefunktionen und ein sichereres Verhalten beim Schließen.

## Lizenz

MIT. Siehe [LICENSE](LICENSE).
