package config

import (
	"strings"
)

// DetectLocale 检测系统 locale, 返回支持的语言 tag
// 支持: "zh-CN", "zh-TW", "en"
func DetectLocale() string {
	raw := detectRawLocale()
	return mapLocale(raw)
}

// mapLocale 将原始 locale 字符串映射到支持的语言
func mapLocale(raw string) string {
	lower := strings.ToLower(raw)

	// 繁体中文
	if containsAny(lower, "zh-hant", "zh_tw", "zh-tw", "zh_hk", "zh-hk", "zh_mo", "zh-mo") {
		return "zh-TW"
	}

	// 简体中文
	if containsAny(lower, "zh-hans", "zh_cn", "zh-cn", "zh_sg", "zh-sg", "zh") {
		return "zh-CN"
	}

	// 英文
	if strings.HasPrefix(lower, "en") {
		return "en"
	}

	// 默认简体中文
	return "zh-CN"
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// detectRawLocale is platform-specific, defined in locale_unix.go / locale_windows.go
var detectRawLocale func() string
