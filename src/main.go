//go:build windows

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	appName    = "RASON Meta Link Fix"
	appVersion = "1.0.1"

	WM_CREATE          = 0x0001
	WM_DESTROY         = 0x0002
	WM_COMMAND         = 0x0111
	WM_CLOSE           = 0x0010
	WM_SETFONT         = 0x0030
	WM_CTLCOLORSTATIC  = 0x0138
	WM_APP_WORK_DONE   = 0x8001
	WM_APP_STATUS_DONE = 0x8002

	WS_OVERLAPPED  = 0x00000000
	WS_CAPTION     = 0x00C00000
	WS_SYSMENU     = 0x00080000
	WS_MINIMIZEBOX = 0x00020000
	WS_VISIBLE     = 0x10000000
	WS_CHILD       = 0x40000000
	WS_TABSTOP     = 0x00010000
	BS_PUSHBUTTON  = 0x00000000
	BS_GROUPBOX    = 0x00000007
	SS_LEFT        = 0x00000000
	SS_CENTER      = 0x00000001

	SW_HIDE = 0
	SW_SHOW = 5

	COLOR_WINDOW = 5

	ID_APPLY   = 1001
	ID_RESTORE = 1002
	ID_RESTART = 1003
	ID_LAUNCH  = 1004
	ID_REFRESH = 1005
	ID_LOG     = 1006
	ID_LANG    = 1007

	MB_OK              = 0x00000000
	MB_YESNO           = 0x00000004
	MB_ICONINFORMATION = 0x00000040
	MB_ICONWARNING     = 0x00000030
	MB_ICONERROR       = 0x00000010
	MB_DEFBUTTON1      = 0x00000000
	IDYES              = 6

	TDCBF_YES_BUTTON = 0x0002
	TDCBF_NO_BUTTON  = 0x0004

	KEY_QUERY_VALUE    = 0x0001
	KEY_SET_VALUE      = 0x0002
	KEY_CREATE_SUB_KEY = 0x0004
	KEY_WOW64_64KEY    = 0x0100
	KEY_WOW64_32KEY    = 0x0200

	REG_SZ    = 1
	REG_DWORD = 4

	ERROR_SUCCESS        = 0
	ERROR_FILE_NOT_FOUND = 2

	TRANSPARENT = 1

	FW_NORMAL   = 400
	FW_SEMIBOLD = 600
	FW_BOLD     = 700

	IDI_APPLICATION = 32512
	IMAGE_ICON      = 1
)

const (
	HKEY_CURRENT_USER  = uintptr(0x80000001)
	HKEY_LOCAL_MACHINE = uintptr(0x80000002)

	linkKey = `Software\Oculus\RemoteHeadset`
	appKey  = `Software\RASON\MetaLinkFix`
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procPostMessageW     = user32.NewProc("PostMessageW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procEnableWindow     = user32.NewProc("EnableWindow")
	procSetWindowTextW   = user32.NewProc("SetWindowTextW")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
	procSendMessageW     = user32.NewProc("SendMessageW")
	procLoadIconW        = user32.NewProc("LoadIconW")
	procLoadImageW       = user32.NewProc("LoadImageW")
	procDestroyIcon      = user32.NewProc("DestroyIcon")
	procGetSysColorBrush = user32.NewProc("GetSysColorBrush")

	procCreateFontW  = gdi32.NewProc("CreateFontW")
	procDeleteObject = gdi32.NewProc("DeleteObject")
	procSetTextColor = gdi32.NewProc("SetTextColor")
	procSetBkMode    = gdi32.NewProc("SetBkMode")

	procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	procGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")

	procRegCreateKeyExW  = advapi32.NewProc("RegCreateKeyExW")
	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW   = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW  = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey      = advapi32.NewProc("RegCloseKey")

	procShellExecuteW = shell32.NewProc("ShellExecuteW")
	procIsUserAnAdmin = shell32.NewProc("IsUserAnAdmin")

	procTaskDialog = comctl32.NewProc("TaskDialog")
)

type POINT struct{ X, Y int32 }
type MSG struct {
	Hwnd     syscall.Handle
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type langPack struct {
	code               string
	windowTitle        string
	subtitle           string
	badgeAdmin         string
	sectionStatus      string
	statusActive       string
	statusInactive     string
	statusOther        string
	labelNvidia        string
	labelMeta          string
	labelMetaVersion   string
	labelTransport     string
	labelService       string
	labelSetting       string
	labelBackup        string
	backupYes          string
	backupNo           string
	exactChange        string
	exactChangeText    string
	safetyLine         string
	btnApply           string
	btnRestore         string
	btnRestart         string
	btnLaunch          string
	btnRefresh         string
	btnLog             string
	btnLanguage        string
	footerReady        string
	footerApplied      string
	footerRestored     string
	footerRefreshed    string
	footerWorking      string
	footerChecking     string
	notFound           string
	versionUnknown     string
	transportUnknown   string
	transportWired     string
	transportAir       string
	serviceRunning     string
	serviceStopped     string
	serviceStarting    string
	serviceStopping    string
	serviceUnknown     string
	disclaimerTitle    string
	disclaimerMain     string
	disclaimerBody     string
	adminRequiredTitle string
	adminRequiredText  string
	applyOkTitle       string
	applyOkText        string
	applyErrTitle      string
	applyErrText       string
	restoreAsk         string
	restoreOk          string
	restoreErr         string
	restartOk          string
	restartErr         string
	launchErr          string
	logErr             string
}

var de = langPack{
	code: "de", windowTitle: "RASON Meta Link Fix", subtitle: "Meta Horizon Link · NVIDIA Freeze Workaround", badgeAdmin: "Als Administrator ausgeführt",
	sectionStatus: "SYSTEM- UND FIXSTATUS", statusActive: "FIX AKTIV", statusInactive: "FIX NICHT AKTIV", statusOther: "ABWEICHENDE EINSTELLUNG",
	labelNvidia: "NVIDIA", labelMeta: "Meta Horizon Link", labelMetaVersion: "Meta-Version", labelTransport: "Verbindung", labelService: "OVRService", labelSetting: "Sliced Encoding", labelBackup: "Backup", backupYes: "Originalzustand gesichert", backupNo: "Noch kein Backup angelegt",
	exactChange: "WAS DIE APP ÄNDERT", exactChangeText: "HKCU\\Software\\Oculus\\RemoteHeadset\\NumSlices = 1  ·  danach Neustart von OVRService",
	safetyLine: "Keine Treiber-/DLL-Patches · kein Downgrade · keine Meta-Dateien verändert · kein Netzwerkzugriff",
	btnApply:   "Fix anwenden", btnRestore: "Original wiederherstellen", btnRestart: "OVRService neu starten", btnLaunch: "Meta Horizon Link öffnen", btnRefresh: "Status aktualisieren", btnLog: "Diagnose-Log öffnen", btnLanguage: "English",
	footerReady: "Bereit. Bestehende Link-Session vor Änderungen am besten trennen.", footerApplied: "Fix ist aktiv, verifiziert und OVRService wurde neu initialisiert.", footerRestored: "Originaleinstellung wurde wiederhergestellt.", footerRefreshed: "Status wurde aktualisiert.", footerWorking: "Vorgang läuft… Die Oberfläche bleibt bedienbar.", footerChecking: "Systemstatus wird geprüft…",
	notFound: "Nicht gefunden", versionUnknown: "Nicht aus Registry lesbar", transportUnknown: "Nicht erkannt", transportWired: "Kabel-Link zuletzt erkannt", transportAir: "Air Link zuletzt erkannt",
	serviceRunning: "Läuft", serviceStopped: "Gestoppt", serviceStarting: "Startet", serviceStopping: "Wird gestoppt", serviceUnknown: "Unbekannt",
	disclaimerTitle: "RASON Meta Link Fix · Sicherheitshinweis", disclaimerMain: "Bitte vor dem Fortfahren lesen",
	disclaimerBody:     "Dieser unabhängige Community-Fix adressiert den Meta-Horizon-Link-Freeze mit neueren NVIDIA-Treibern, indem er ausschließlich Sliced Encoding deaktiviert (NumSlices=1) und den Meta-Link-Dienst OVRService neu startet.\n\nDer Fix patcht KEINE NVIDIA- oder Meta-Binaries, installiert oder entfernt KEINEN Grafiktreiber, blendet die Meta-Warnung NICHT aus und verändert KEINE weiteren Windows-Grafikeinstellungen. Vor der ersten Änderung wird der vorhandene NumSlices-Zustand gesichert und kann wiederhergestellt werden.\n\nDie App ist kein offizielles Produkt von Meta oder NVIDIA. Der zugrunde liegende Workaround wurde auf einem betroffenen Quest-3/NVIDIA-System erfolgreich gegen Freezes/Stutter getestet. Für andere Hardware- und Softwarestände kann keine Garantie gegeben werden. Nutzung auf eigenes Risiko.\n\nMit „Ja“ akzeptierst du den Hinweis und startest die App.",
	adminRequiredTitle: "Administratorrechte erforderlich", adminRequiredText: "Diese App muss als Administrator ausgeführt werden, damit OVRService zuverlässig neu gestartet werden kann. Windows zeigt jetzt eine UAC-Abfrage.",
	applyOkTitle: "Fix aktiviert", applyOkText: "Sliced Encoding wurde deaktiviert (NumSlices=1), der Wert wurde zurückgelesen und OVRService wurde neu gestartet. Du kannst Meta Horizon Link jetzt testen.",
	applyErrTitle: "Fix konnte nicht vollständig angewendet werden", applyErrText: "Der Vorgang ist fehlgeschlagen:\n\n",
	restoreAsk: "Originale Sliced-Encoding-Einstellung wirklich wiederherstellen? Der beim ersten Anwenden gesicherte Zustand wird zurückgesetzt.", restoreOk: "Originaleinstellung wurde wiederhergestellt. Bei einer laufenden Link-Session bitte neu verbinden oder OVRService neu starten.", restoreErr: "Wiederherstellung fehlgeschlagen:\n\n",
	restartOk: "OVRService wurde erfolgreich neu gestartet.", restartErr: "OVRService konnte nicht neu gestartet werden:\n\n", launchErr: "Meta Horizon Link konnte nicht gestartet werden:\n\n", logErr: "Diagnose-Log konnte nicht geöffnet werden:\n\n",
}

var en = langPack{
	code: "en", windowTitle: "RASON Meta Link Fix", subtitle: "Meta Horizon Link · NVIDIA Freeze Workaround", badgeAdmin: "Running as administrator",
	sectionStatus: "SYSTEM AND FIX STATUS", statusActive: "FIX ACTIVE", statusInactive: "FIX NOT ACTIVE", statusOther: "NON-STANDARD SETTING",
	labelNvidia: "NVIDIA", labelMeta: "Meta Horizon Link", labelMetaVersion: "Meta version", labelTransport: "Connection", labelService: "OVRService", labelSetting: "Sliced Encoding", labelBackup: "Backup", backupYes: "Original state backed up", backupNo: "No backup created yet",
	exactChange: "WHAT THIS APP CHANGES", exactChangeText: "HKCU\\Software\\Oculus\\RemoteHeadset\\NumSlices = 1  ·  then restarts OVRService",
	safetyLine: "No driver/DLL patches · no downgrade · no Meta files modified · no network access",
	btnApply:   "Apply fix", btnRestore: "Restore original", btnRestart: "Restart OVRService", btnLaunch: "Open Meta Horizon Link", btnRefresh: "Refresh status", btnLog: "Open diagnostics log", btnLanguage: "Deutsch",
	footerReady: "Ready. It is best to disconnect an active Link session before making changes.", footerApplied: "Fix is active, verified, and OVRService was reinitialized.", footerRestored: "Original setting restored.", footerRefreshed: "Status refreshed.", footerWorking: "Operation in progress… The interface remains responsive.", footerChecking: "Checking system status…",
	notFound: "Not found", versionUnknown: "Not readable from registry", transportUnknown: "Not detected", transportWired: "Wired Link last detected", transportAir: "Air Link last detected",
	serviceRunning: "Running", serviceStopped: "Stopped", serviceStarting: "Starting", serviceStopping: "Stopping", serviceUnknown: "Unknown",
	disclaimerTitle: "RASON Meta Link Fix · Safety notice", disclaimerMain: "Please read before continuing",
	disclaimerBody:     "This independent community workaround targets the Meta Horizon Link freeze seen with newer NVIDIA drivers by changing only one Link encoder setting: Sliced Encoding is disabled (NumSlices=1), then the Meta Link service OVRService is restarted.\n\nThe tool does NOT patch NVIDIA or Meta binaries, does NOT install or remove a graphics driver, does NOT hide Meta's warning, and does NOT change other Windows graphics settings. Before the first change, the existing NumSlices state is backed up and can be restored.\n\nThis is not an official Meta or NVIDIA product. The underlying workaround has successfully eliminated freezes/stutter on an affected Quest 3/NVIDIA system. No guarantee can be made for every hardware/software combination. Use at your own risk.\n\nChoose “Yes” to accept this notice and start the app.",
	adminRequiredTitle: "Administrator rights required", adminRequiredText: "This app runs as administrator so it can reliably restart OVRService. Windows will now show a UAC prompt.",
	applyOkTitle: "Fix enabled", applyOkText: "Sliced Encoding was disabled (NumSlices=1), the value was read back successfully, and OVRService was restarted. You can test Meta Horizon Link now.",
	applyErrTitle: "Fix could not be fully applied", applyErrText: "The operation failed:\n\n",
	restoreAsk: "Restore the original Sliced Encoding setting? The state captured before the first change will be restored.", restoreOk: "Original setting restored. If a Link session is active, reconnect it or restart OVRService.", restoreErr: "Restore failed:\n\n",
	restartOk: "OVRService restarted successfully.", restartErr: "OVRService could not be restarted:\n\n", launchErr: "Meta Horizon Link could not be started:\n\n", logErr: "Diagnostics log could not be opened:\n\n",
}

var (
	tr                                                                                            = en
	mainWnd                                                                                       syscall.Handle
	titleWnd, subtitleWnd, adminWnd, sectionWnd                                                   syscall.Handle
	fixStateWnd, statusWnd, exactTitleWnd, exactWnd, safetyWnd, footerWnd                         syscall.Handle
	btnApplyWnd, btnRestoreWnd, btnRestartWnd, btnLaunchWnd, btnRefreshWnd, btnLogWnd, btnLangWnd syscall.Handle
	fontTitle, fontSubtitle, fontStatus, fontNormal, fontSmall                                    syscall.Handle
	iconLarge, iconSmall                                                                          syscall.Handle
	fixColor                                                                                      uint32
	busy, closing                                                                                 bool
	resultMu                                                                                      sync.Mutex
	lastJob                                                                                       jobResult
	lastStatus                                                                                    statusResult
	statusSeq                                                                                     uint32
)

type jobKind uint16

const (
	jobApply jobKind = iota + 1
	jobRestore
	jobRestart
)

type jobResult struct {
	kind    jobKind
	err     error
	warning error
}

type statusResult struct {
	seq     uint32
	state   string
	details string
	color   uint32
	notify  bool
}

func utf16ptr(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func loWord(v uintptr) uint16   { return uint16(v & 0xffff) }
func rgb(r, g, b byte) uint32   { return uint32(r) | uint32(g)<<8 | uint32(b)<<16 }

func setText(hwnd syscall.Handle, s string) {
	if hwnd != 0 {
		procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16ptr(s))))
	}
}
func msgBox(hwnd syscall.Handle, title, text string, flags uintptr) int {
	r, _, _ := procMessageBoxW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16ptr(text))), uintptr(unsafe.Pointer(utf16ptr(title))), flags)
	return int(r)
}

func systemLanguage() langPack {
	r, _, _ := procGetUserDefaultUILanguage.Call()
	primary := uint16(r) & 0x03ff
	if primary == 0x07 {
		return de
	}
	return en
}

func showDisclaimer() bool {
	var pressed int32
	if err := procTaskDialog.Find(); err == nil {
		hr, _, _ := procTaskDialog.Call(0, 0,
			uintptr(unsafe.Pointer(utf16ptr(tr.disclaimerTitle))),
			uintptr(unsafe.Pointer(utf16ptr(tr.disclaimerMain))),
			uintptr(unsafe.Pointer(utf16ptr(tr.disclaimerBody))),
			TDCBF_YES_BUTTON|TDCBF_NO_BUTTON, 0,
			uintptr(unsafe.Pointer(&pressed)))
		if int32(hr) >= 0 {
			return pressed == IDYES
		}
	}
	return msgBox(0, tr.disclaimerTitle, tr.disclaimerMain+"\n\n"+tr.disclaimerBody, MB_YESNO|MB_ICONWARNING|MB_DEFBUTTON1) == IDYES
}

func isAdmin() bool { r, _, _ := procIsUserAnAdmin.Call(); return r != 0 }
func elevateSelf() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	r, _, callErr := procShellExecuteW.Call(0,
		uintptr(unsafe.Pointer(utf16ptr("runas"))),
		uintptr(unsafe.Pointer(utf16ptr(exe))), 0, 0, SW_SHOW)
	if r <= 32 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return callErr
		}
		return fmt.Errorf("ShellExecute error code %d", r)
	}
	return nil
}

func logPath() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "RASON", "MetaLinkFix")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "MetaLinkFix.log")
}

func loadResourceIcon(hInst uintptr, size int32) syscall.Handle {
	// Resource ID 1 is the application icon embedded by embed_windows_resources.py.
	r, _, _ := procLoadImageW.Call(hInst, 1, IMAGE_ICON, uintptr(size), uintptr(size), 0)
	if r != 0 {
		return syscall.Handle(r)
	}
	r, _, _ = procLoadIconW.Call(0, IDI_APPLICATION)
	return syscall.Handle(r)
}

func logf(format string, args ...any) {
	line := fmt.Sprintf("%s  %s\r\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	if f, err := os.OpenFile(logPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		_, _ = f.WriteString(line)
		_ = f.Close()
	}
}

func regCreateOrOpen(root uintptr, subkey string, sam uint32) (syscall.Handle, error) {
	var h syscall.Handle
	var disp uint32
	r, _, _ := procRegCreateKeyExW.Call(root, uintptr(unsafe.Pointer(utf16ptr(subkey))), 0, 0, 0, uintptr(sam), 0, uintptr(unsafe.Pointer(&h)), uintptr(unsafe.Pointer(&disp)))
	if r != ERROR_SUCCESS {
		return 0, syscall.Errno(r)
	}
	return h, nil
}
func regOpen(root uintptr, subkey string, sam uint32) (syscall.Handle, error) {
	var h syscall.Handle
	r, _, _ := procRegOpenKeyExW.Call(root, uintptr(unsafe.Pointer(utf16ptr(subkey))), 0, uintptr(sam), uintptr(unsafe.Pointer(&h)))
	if r != ERROR_SUCCESS {
		return 0, syscall.Errno(r)
	}
	return h, nil
}
func regClose(h syscall.Handle) { procRegCloseKey.Call(uintptr(h)) }
func regQueryDWORD(root uintptr, subkey, name string, wow uint32) (uint32, bool, error) {
	h, err := regOpen(root, subkey, KEY_QUERY_VALUE|wow)
	if err != nil {
		if e, ok := err.(syscall.Errno); ok && uintptr(e) == ERROR_FILE_NOT_FOUND {
			return 0, false, nil
		}
		return 0, false, err
	}
	defer regClose(h)
	var typ, data, size uint32
	size = 4
	r, _, _ := procRegQueryValueExW.Call(uintptr(h), uintptr(unsafe.Pointer(utf16ptr(name))), 0, uintptr(unsafe.Pointer(&typ)), uintptr(unsafe.Pointer(&data)), uintptr(unsafe.Pointer(&size)))
	if r == ERROR_FILE_NOT_FOUND {
		return 0, false, nil
	}
	if r != ERROR_SUCCESS {
		return 0, false, syscall.Errno(r)
	}
	if typ != REG_DWORD || size != 4 {
		return 0, true, fmt.Errorf("unexpected registry type for %s", name)
	}
	return data, true, nil
}
func regSetDWORD(root uintptr, subkey, name string, value uint32) error {
	h, err := regCreateOrOpen(root, subkey, KEY_SET_VALUE|KEY_CREATE_SUB_KEY)
	if err != nil {
		return err
	}
	defer regClose(h)
	r, _, _ := procRegSetValueExW.Call(uintptr(h), uintptr(unsafe.Pointer(utf16ptr(name))), 0, REG_DWORD, uintptr(unsafe.Pointer(&value)), 4)
	if r != ERROR_SUCCESS {
		return syscall.Errno(r)
	}
	return nil
}
func regDeleteValue(root uintptr, subkey, name string) error {
	h, err := regOpen(root, subkey, KEY_SET_VALUE)
	if err != nil {
		if e, ok := err.(syscall.Errno); ok && uintptr(e) == ERROR_FILE_NOT_FOUND {
			return nil
		}
		return err
	}
	defer regClose(h)
	r, _, _ := procRegDeleteValueW.Call(uintptr(h), uintptr(unsafe.Pointer(utf16ptr(name))))
	if r == ERROR_FILE_NOT_FOUND {
		return nil
	}
	if r != ERROR_SUCCESS {
		return syscall.Errno(r)
	}
	return nil
}
func regQueryString(root uintptr, subkey, name string, wow uint32) (string, bool) {
	h, err := regOpen(root, subkey, KEY_QUERY_VALUE|wow)
	if err != nil {
		return "", false
	}
	defer regClose(h)
	var typ, size uint32
	r, _, _ := procRegQueryValueExW.Call(uintptr(h), uintptr(unsafe.Pointer(utf16ptr(name))), 0, uintptr(unsafe.Pointer(&typ)), 0, uintptr(unsafe.Pointer(&size)))
	if r != ERROR_SUCCESS || size < 2 || (typ != REG_SZ && typ != 2) {
		return "", false
	}
	buf := make([]uint16, size/2+1)
	r, _, _ = procRegQueryValueExW.Call(uintptr(h), uintptr(unsafe.Pointer(utf16ptr(name))), 0, uintptr(unsafe.Pointer(&typ)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if r != ERROR_SUCCESS {
		return "", false
	}
	s := syscall.UTF16ToString(buf)
	if typ == 2 {
		s = os.ExpandEnv(s)
	}
	return s, true
}

func ensureBackup() error {
	captured, ok, err := regQueryDWORD(HKEY_CURRENT_USER, appKey, "BackupCaptured", 0)
	if err != nil {
		return err
	}
	if ok && captured == 1 {
		return nil
	}
	v, exists, err := regQueryDWORD(HKEY_CURRENT_USER, linkKey, "NumSlices", 0)
	if err != nil {
		return fmt.Errorf("NumSlices backup failed: %w", err)
	}
	if exists {
		if err := regSetDWORD(HKEY_CURRENT_USER, appKey, "BackupNumSlicesExisted", 1); err != nil {
			return err
		}
		if err := regSetDWORD(HKEY_CURRENT_USER, appKey, "BackupNumSlicesValue", v); err != nil {
			return err
		}
	} else {
		if err := regSetDWORD(HKEY_CURRENT_USER, appKey, "BackupNumSlicesExisted", 0); err != nil {
			return err
		}
		_ = regDeleteValue(HKEY_CURRENT_USER, appKey, "BackupNumSlicesValue")
	}
	if err := regSetDWORD(HKEY_CURRENT_USER, appKey, "BackupCaptured", 1); err != nil {
		return err
	}
	logf("Backup captured: NumSlices existed=%v value=%d", exists, v)
	return nil
}
func applyFix() error {
	if err := ensureBackup(); err != nil {
		return err
	}
	if err := regSetDWORD(HKEY_CURRENT_USER, linkKey, "NumSlices", 1); err != nil {
		return fmt.Errorf("could not set NumSlices=1: %w", err)
	}
	v, exists, err := regQueryDWORD(HKEY_CURRENT_USER, linkKey, "NumSlices", 0)
	if err != nil || !exists || v != 1 {
		return fmt.Errorf("write verification failed (exists=%v value=%d error=%v)", exists, v, err)
	}
	_ = regSetDWORD(HKEY_CURRENT_USER, appKey, "FixActive", 1)
	logf("Fix applied and read-back verified: HKCU\\%s\\NumSlices=1", linkKey)
	return nil
}
func restoreFix() error {
	captured, ok, err := regQueryDWORD(HKEY_CURRENT_USER, appKey, "BackupCaptured", 0)
	if err != nil {
		return err
	}
	if !ok || captured != 1 {
		return fmt.Errorf("no backup created by this app is available; nothing was changed")
	}
	existed, _, err := regQueryDWORD(HKEY_CURRENT_USER, appKey, "BackupNumSlicesExisted", 0)
	if err != nil {
		return err
	}
	if existed == 1 {
		old, ok, err := regQueryDWORD(HKEY_CURRENT_USER, appKey, "BackupNumSlicesValue", 0)
		if err != nil || !ok {
			return fmt.Errorf("backup value is missing or invalid")
		}
		if err := regSetDWORD(HKEY_CURRENT_USER, linkKey, "NumSlices", old); err != nil {
			return err
		}
		logf("Original setting restored: NumSlices=%d", old)
	} else {
		if err := regDeleteValue(HKEY_CURRENT_USER, linkKey, "NumSlices"); err != nil {
			return err
		}
		logf("Original state restored: NumSlices was unset")
	}
	_ = regDeleteValue(HKEY_CURRENT_USER, appKey, "BackupCaptured")
	_ = regDeleteValue(HKEY_CURRENT_USER, appKey, "BackupNumSlicesExisted")
	_ = regDeleteValue(HKEY_CURRENT_USER, appKey, "BackupNumSlicesValue")
	_ = regDeleteValue(HKEY_CURRENT_USER, appKey, "FixActive")
	return nil
}
func backupExists() bool {
	v, ok, _ := regQueryDWORD(HKEY_CURRENT_USER, appKey, "BackupCaptured", 0)
	return ok && v == 1
}

func hiddenOutput(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := c.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("command timed out: %s", name)
	}
	return out, err
}
func nvidiaInfo(lp langPack) string {
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	for _, p := range []string{filepath.Join(windir, "System32", "nvidia-smi.exe"), "nvidia-smi.exe"} {
		out, err := hiddenOutput(3*time.Second, p, "--query-gpu=name,driver_version", "--format=csv,noheader")
		if err == nil {
			line := strings.TrimSpace(string(out))
			if line != "" {
				return strings.Split(line, "\n")[0]
			}
		}
	}
	return lp.notFound
}
func metaBase() string {
	key := `SOFTWARE\Oculus VR, LLC\Oculus`
	if s, ok := regQueryString(HKEY_LOCAL_MACHINE, key, "Base", KEY_WOW64_64KEY); ok && s != "" {
		return s
	}
	if s, ok := regQueryString(HKEY_LOCAL_MACHINE, key, "Base", KEY_WOW64_32KEY); ok && s != "" {
		return s
	}
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		c := filepath.Join(pf, "Oculus")
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
func metaVersion(lp langPack) string {
	key := `SOFTWARE\Oculus VR, LLC\Oculus`
	for _, n := range []string{"Version", "CoreVersion", "ClientVersion"} {
		if s, ok := regQueryString(HKEY_LOCAL_MACHINE, key, n, KEY_WOW64_64KEY); ok && strings.TrimSpace(s) != "" {
			return s
		}
		if s, ok := regQueryString(HKEY_LOCAL_MACHINE, key, n, KEY_WOW64_32KEY); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return lp.versionUnknown
}
func metaClientPath() string {
	base := metaBase()
	var c []string
	if base != "" {
		c = append(c, filepath.Join(base, "Support", "oculus-client", "OculusClient.exe"), filepath.Join(base, "OculusClient.exe"))
	}
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		c = append(c, filepath.Join(pf, "Oculus", "Support", "oculus-client", "OculusClient.exe"))
	}
	for _, p := range c {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

type svcState int

const (
	svcUnknown svcState = iota
	svcStopped
	svcStartPending
	svcStopPending
	svcRunning
)

func serviceStateRaw() svcState {
	out, err := hiddenOutput(3*time.Second, "sc.exe", "query", "OVRService")
	if err != nil && len(out) == 0 {
		return svcUnknown
	}
	s := string(out)
	re := regexp.MustCompile(`(?im)^\s*(?:STATE|ZUSTAND)\s*:\s*([1-7])\b`)
	if m := re.FindStringSubmatch(s); len(m) == 2 {
		switch m[1] {
		case "1":
			return svcStopped
		case "2":
			return svcStartPending
		case "3":
			return svcStopPending
		case "4":
			return svcRunning
		}
	}
	u := strings.ToUpper(s)
	switch {
	case strings.Contains(u, "RUNNING"):
		return svcRunning
	case strings.Contains(u, "STOP_PENDING"):
		return svcStopPending
	case strings.Contains(u, "START_PENDING"):
		return svcStartPending
	case strings.Contains(u, "STOPPED"):
		return svcStopped
	}
	return svcUnknown
}
func serviceStateText(lp langPack) string {
	switch serviceStateRaw() {
	case svcRunning:
		return lp.serviceRunning
	case svcStopped:
		return lp.serviceStopped
	case svcStartPending:
		return lp.serviceStarting
	case svcStopPending:
		return lp.serviceStopping
	default:
		return lp.serviceUnknown
	}
}
func waitService(target svcState, max time.Duration) bool {
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		if serviceStateRaw() == target {
			return true
		}
		time.Sleep(450 * time.Millisecond)
	}
	return serviceStateRaw() == target
}
func restartService() error {
	logf("OVRService restart requested")
	state := serviceStateRaw()
	if state == svcRunning || state == svcStopPending || state == svcStartPending {
		out, err := hiddenOutput(5*time.Second, "sc.exe", "stop", "OVRService")
		logf("sc stop OVRService: %s", strings.TrimSpace(string(out)))
		if err != nil && !strings.Contains(strings.ToUpper(string(out)), "SERVICE_NOT_ACTIVE") {
			return fmt.Errorf("stop failed: %v / %s", err, strings.TrimSpace(string(out)))
		}
		if !waitService(svcStopped, 15*time.Second) {
			return fmt.Errorf("OVRService did not stop in time")
		}
	}
	out, err := hiddenOutput(5*time.Second, "sc.exe", "start", "OVRService")
	logf("sc start OVRService: %s", strings.TrimSpace(string(out)))
	if err != nil && !strings.Contains(strings.ToUpper(string(out)), "SERVICE_ALREADY_RUNNING") {
		return fmt.Errorf("start failed: %v / %s", err, strings.TrimSpace(string(out)))
	}
	if !waitService(svcRunning, 15*time.Second) {
		return fmt.Errorf("OVRService did not start in time")
	}
	logf("OVRService is running")
	return nil
}
func transportInfo(lp langPack) string {
	p := filepath.Join(os.Getenv("LOCALAPPDATA"), "Oculus", "DeviceCache.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return lp.transportUnknown
	}
	var temp any
	if json.Unmarshal(b, &temp) != nil {
		return lp.transportUnknown
	}
	compact := strings.ToLower(strings.Join(strings.Fields(string(b)), ""))
	if strings.Contains(compact, `"isusingairlink":false`) {
		return lp.transportWired
	}
	if strings.Contains(compact, `"isusingairlink":true`) {
		return lp.transportAir
	}
	return lp.transportUnknown
}

func statusSnapshot(lp langPack) (string, string, uint32) {
	slices, exists, err := regQueryDWORD(HKEY_CURRENT_USER, linkKey, "NumSlices", 0)
	state := lp.statusInactive
	color := rgb(180, 45, 45)
	setting := "Default"
	if lp.code == "de" {
		setting = "Standard"
	}
	if err != nil {
		state = lp.statusOther
		if lp.code == "de" {
			setting = "Registry-Lesefehler"
		} else {
			setting = "Registry read error"
		}
		color = rgb(170, 110, 0)
	} else if exists && slices == 1 {
		state = lp.statusActive
		setting = "OFF · NumSlices=1"
		color = rgb(22, 128, 73)
	} else if exists {
		state = lp.statusOther
		setting = fmt.Sprintf("NumSlices=%d", slices)
		color = rgb(170, 110, 0)
	}
	base := metaBase()
	if base == "" {
		base = lp.notFound
	}
	backup := lp.backupNo
	if backupExists() {
		backup = lp.backupYes
	}
	details := fmt.Sprintf("%s:  %s\r\n%s:  %s\r\n%s:  %s\r\n%s:  %s\r\n%s:  %s\r\n%s:  %s\r\n%s:  %s",
		lp.labelNvidia, nvidiaInfo(lp), lp.labelMeta, base, lp.labelMetaVersion, metaVersion(lp), lp.labelTransport, transportInfo(lp), lp.labelService, serviceStateText(lp), lp.labelSetting, setting, lp.labelBackup, backup)
	return state, details, color
}
func setControlsEnabled(enabled bool) {
	v := uintptr(0)
	if enabled {
		v = 1
	}
	for _, h := range []syscall.Handle{btnApplyWnd, btnRestoreWnd, btnRestartWnd, btnLaunchWnd, btnRefreshWnd, btnLogWnd, btnLangWnd} {
		if h != 0 {
