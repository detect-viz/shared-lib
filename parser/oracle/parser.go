package oracle

import "bytes"

// Parser Oracle 解析器介面
type Parser interface {
	Parse(content *bytes.Buffer) (map[string]interface{}, error)
}

// TableSpaceParser TableSpace 解析器
type TableSpaceParser struct{}

// ConnectionParser 連接數解析器
type ConnectionParser struct{}

// 根據文件類型選擇對應的解析器
func NewParser(fileType string) Parser {
	switch fileType {
	case "awr":
		return &AWRParser{}
	case "tablespace":
		return &TableSpaceParser{}
	case "connection":
		return &ConnectionParser{}
	default:
		return nil
	}
}
