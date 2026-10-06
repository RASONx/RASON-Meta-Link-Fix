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
