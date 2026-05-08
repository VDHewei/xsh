//go:build !windows

package config

import "os"

func init() {
	detectRawLocale = func() string {
		if v := os.Getenv("LANG"); v != "" {
			return v
		}
		if v := os.Getenv("LC_ALL"); v != "" {
			return v
		}
		if v := os.Getenv("LC_MESSAGES"); v != "" {
			return v
		}
		return ""
	}
}
