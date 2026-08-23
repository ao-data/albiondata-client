// internal/console/console_windows.go
//go:build windows

package console

import (
	"sync/atomic"
	"syscall"
	"unsafe"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	user32                    = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow      = kernel32.NewProc("GetConsoleWindow")
	procShowWindow            = user32.NewProc("ShowWindow")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

// SW_HIDE and SW_SHOW are the ShowWindow nCmdShow values used here. See
// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-showwindow
const (
	swHide = 0
	swShow = 5
)

var hidden atomic.Bool

// Supported reports whether console hiding is implemented on this
// platform. Always true on Windows.
func Supported() bool {
	return true
}

// Owned reports whether this process is the only one attached to its
// console - true for a fresh Explorer/shortcut launch (Windows allocates
// a private console), false when sharing a parent shell's console (e.g.
// launched from an already-open PowerShell/cmd window). Hiding a shared
// console would hide the user's own terminal with no way to restore it,
// so callers should only Hide() when this returns true.
func Owned() bool {
	var pids [2]uint32
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), 2)
	return n == 1
}

// Hide hides the process's console window, if one exists (e.g. none
// exists if the process wasn't allocated a console at all).
func Hide() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swHide)
	hidden.Store(true)
}

// Show reveals the process's console window, if one exists.
func Show() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swShow)
	hidden.Store(false)
}

// Hidden reports whether Hide has been called more recently than Show.
func Hidden() bool {
	return hidden.Load()
}
