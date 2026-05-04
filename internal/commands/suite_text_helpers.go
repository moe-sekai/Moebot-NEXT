package commands

import (
	"time"

	"moebot-next/internal/suite"
)

func suiteSourceText(profile suite.BaseProfile) string {
	source := profile.Source
	if profile.LocalSource != "" {
		source += "(" + profile.LocalSource + ")"
	}
	if source == "" {
		return "未知"
	}
	return source
}

func suiteUpdateText(uploadTime int64) string {
	if uploadTime <= 0 {
		return "未知"
	}
	return time.UnixMilli(uploadTime).Format("2006-01-02 15:04:05")
}
