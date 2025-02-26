package utils

import (
	"time"

	"github.com/detect-viz/shared-lib/notifier/errors"
)

// FormatTime 格式化時間
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime 解析時間
func ParseTime(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return t, errors.NewNotifyError("Time", "invalid time format", err)
	}
	return t, nil
}
