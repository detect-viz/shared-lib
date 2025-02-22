package parser

import (
	"time"
)

// 檔案來源和日期資訊
type FileInfo struct {
	Realm     string    `json:"realm"`
	Source    string    `json:"source"`
	FileName  string    `json:"file_name"`
	Host      string    `json:"host"`
	UserID    string    `json:"user_id"`
	Status    bool      `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
