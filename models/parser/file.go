package parser

import (
	"bytes"
	"time"

	"shared-lib/models/common"
)

// 檔案來源和日期資訊
type FileInfo struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	User       common.SSOUser
	FileName   string
	SourceName string
	Hostname   string        // 主機名
	Timestamp  time.Time     // 時間戳
	Content    *bytes.Buffer // 內容
}
