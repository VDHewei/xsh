//go:build windows

package config

import (
	"syscall"
	"unsafe"
)

// detectRawLocale Windows 通过 GetUserDefaultLocaleName API 获取
func init() {
	// 覆盖 locale.go 中的默认实现
	detectRawLocale = detectWindowsLocale
}

// detectWindowsLocale 调用 GetUserDefaultLocaleName 获取 Windows 用户默认 locale
func detectWindowsLocale() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getUserDefaultLocaleName := kernel32.NewProc("GetUserDefaultLocaleName")

	buf := make([]uint16, 85) // LOCALE_NAME_MAX_LENGTH
	ret, _, _ := getUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
