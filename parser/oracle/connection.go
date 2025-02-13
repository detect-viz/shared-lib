package oracle

import (
	"bufio"
	"bytes"
	"shared-lib/parser/base"
	"strconv"
	"strings"
)

// ConnectionParser 連接數解析器
type ConnectionParser struct {
	*base.BaseParser
}

// NewConnectionParser 創建連接數解析器
func NewConnectionParser() *ConnectionParser {
	return &ConnectionParser{
		BaseParser: base.NewBaseParser(),
	}
}

// Parse 實現 Parser 介面
func (p *ConnectionParser) Parse(content *bytes.Buffer) (map[string]interface{}, error) {
	// 根據文件內容判斷是日誌還是錯誤文件
	if strings.Contains(content.String(), "success") {
		return parseOracleConnLog(content)
	}
	return parseOracleConnErr(content)
}

// parseOracleConnLog 解析連接日誌 (.log)
// 格式: timestamp,dbname,status
func parseOracleConnLog(content *bytes.Buffer) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	scanner := bufio.NewScanner(content)

	// 初始化指標切片
	connectionStatus := make([]map[string]interface{}, 0)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			continue
		}

		// 解析時間戳
		timestamp, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			continue
		}

		// 解析狀態
		status := 0.0 // success = 0, failed = 1
		if strings.TrimSpace(parts[2]) == "failed" {
			status = 1.0
		}

		metric := map[string]interface{}{
			"timestamp": timestamp,
			"dbname":    strings.TrimSpace(parts[1]),
			"value":     status,
		}
		connectionStatus = append(connectionStatus, metric)
	}

	metrics["connection_status"] = connectionStatus
	return metrics, nil
}

// parseOracleConnErr 解析錯誤日誌 (.err)
// 格式: timestamp,dbname,status
func parseOracleConnErr(content *bytes.Buffer) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	scanner := bufio.NewScanner(content)

	// 初始化指標切片
	connectionRefused := make([]map[string]interface{}, 0)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 3 || strings.TrimSpace(parts[2]) != "failed" {
			continue
		}

		// 解析時間戳
		timestamp, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			continue
		}

		metric := map[string]interface{}{
			"timestamp": timestamp,
			"dbname":    strings.TrimSpace(parts[1]),
			"value":     1.0, // failed = 1
		}
		connectionRefused = append(connectionRefused, metric)
	}

	metrics["connection_refused"] = connectionRefused
	return metrics, nil
}
