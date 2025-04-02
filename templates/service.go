package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/google/wire"
)

var TemplateSet = wire.NewSet(NewService, wire.Bind(new(Service), new(*serviceImpl)))

// Service 處理格式轉換
type serviceImpl struct {
	logger logger.Logger
}

// NewService 創建模板服務
func NewService(logger logger.Logger) *serviceImpl {
	return &serviceImpl{
		logger: logger,
	}
}

// RenderMessage 渲染通知模板
func (s *serviceImpl) RenderMessage(t models.Template, data map[string]interface{}) (string, error) {
	var messageTemplate string

	// 根據 FormatType 選擇對應的模板
	switch t.FormatType {
	case "html":
		messageTemplate = t.Message
	case "markdown":
		messageTemplate = t.Message
	case "text":
		messageTemplate = t.Message
	case "json":
		// JSON 需要額外解析成標準格式
		return renderJSONTemplate(t.Message, data)
	default:
		return "", fmt.Errorf("未知的模板格式: %s", t.FormatType)
	}

	// 確保 severity 值存在且有效
	if severity, ok := data["severity"].(string); ok {
		// 如果 severity 為空，設置默認值
		if severity == "" {
			data["severity"] = "crit"
		}
	} else {
		// 如果 severity 不存在或不是字符串，設置默認值
		data["severity"] = "crit"
	}

	// 創建模板並添加自定義函數
	tmpl := template.New("message-" + t.FormatType).Funcs(template.FuncMap{
		// 添加 each 函數，用於遍歷數組
		"each": func(arr interface{}, fn func(interface{}) string) string {
			var result string
			switch v := arr.(type) {
			case []interface{}:
				for _, item := range v {
					result += fn(item)
				}
			case []map[string]interface{}:
				for _, item := range v {
					result += fn(item)
				}
			}
			return result
		},
		// 添加 if_eq 函數，用於比較兩個值是否相等
		"if_eq": func(a, b interface{}) bool {
			return a == b
		},
		// 添加 unless 函數，用於反向條件判斷
		"unless": func(cond bool, fn func() string) string {
			if !cond {
				return fn()
			}
			return ""
		},
		// 添加 last 函數，用於判斷是否為最後一個元素
		"last": func(index, length int) bool {
			return index == length-1
		},
		// 添加 severity_format 函數，用於格式化 severity 值
		"severity_format": func(severity interface{}) string {
			if s, ok := severity.(string); ok {
				// 去除前後空格
				s = strings.TrimSpace(s)
				switch strings.ToLower(s) {
				case "critical", "crit":
					return "Critical"
				case "warning", "warn":
					return "Warning"
				case "info":
					return "Info"
				default:
					return s
				}
			}
			return "Unknown"
		},
	})

	// 解析模板
	parsedTmpl, err := tmpl.Parse(messageTemplate)
	if err != nil {
		return "", fmt.Errorf("解析模板失敗: %w", err)
	}

	// 渲染模板
	var messageBuf bytes.Buffer
	if err := parsedTmpl.Execute(&messageBuf, data); err != nil {
		return "", fmt.Errorf("渲染模板失敗: %w", err)
	}

	// 處理消息格式
	message := messageBuf.String()
	if t.FormatType != "json" { // JSON 格式不需要處理每一行
		// 1. 分割成行
		lines := strings.Split(message, "\n")

		// 2. 處理每一行，保留有意義的縮排
		var processedLines []string
		for _, line := range lines {
			// 計算前導空格數量
			leadingSpaces := 0
			for i, char := range line {
				if char != ' ' {
					leadingSpaces = i
					break
				}
			}

			// 去除尾部空格
			line = strings.TrimRight(line, " ")

			// 如果是空行，不添加任何空格
			if len(strings.TrimSpace(line)) == 0 {
				processedLines = append(processedLines, "")
				continue
			}

			// 保留最多 2 個前導空格的縮排
			if leadingSpaces > 0 {
				// 對於告警詳情，保留縮排但標準化為 2 個空格
				processedLines = append(processedLines, "  "+strings.TrimSpace(line))
			} else {
				processedLines = append(processedLines, strings.TrimSpace(line))
			}
		}

		// 3. 移除連續的空白行，只保留一個
		var finalLines []string
		var prevLineEmpty bool = false
		for _, line := range processedLines {
			isEmptyLine := len(line) == 0

			// 如果當前行是空行且前一行也是空行，則跳過
			if isEmptyLine && prevLineEmpty {
				continue
			}

			finalLines = append(finalLines, line)
			prevLineEmpty = isEmptyLine
		}

		// 4. 移除開頭和結尾的空行
		for len(finalLines) > 0 && finalLines[0] == "" {
			finalLines = finalLines[1:]
		}
		for len(finalLines) > 0 && finalLines[len(finalLines)-1] == "" {
			finalLines = finalLines[:len(finalLines)-1]
		}

		// 5. 重新組合消息
		message = strings.Join(finalLines, "\n")
	}

	return message, nil
}

func renderJSONTemplate(jsonTemplate string, data map[string]interface{}) (string, error) {
	// 確保 severity 值存在且有效
	if severity, ok := data["severity"].(string); ok {
		// 如果 severity 為空，設置默認值
		if severity == "" {
			data["severity"] = "crit"
		}
	} else {
		// 如果 severity 不存在或不是字符串，設置默認值
		data["severity"] = "crit"
	}

	// 創建模板並添加自定義函數
	tmpl := template.New("json").Funcs(template.FuncMap{
		// 添加 each 函數，用於遍歷數組
		"each": func(arr interface{}, fn func(interface{}) string) string {
			var result string
			switch v := arr.(type) {
			case []interface{}:
				for _, item := range v {
					result += fn(item)
				}
			case []map[string]interface{}:
				for _, item := range v {
					result += fn(item)
				}
			}
			return result
		},
		// 添加 if_eq 函數，用於比較兩個值是否相等
		"if_eq": func(a, b interface{}) bool {
			return a == b
		},
		// 添加 unless 函數，用於反向條件判斷
		"unless": func(cond bool, fn func() string) string {
			if !cond {
				return fn()
			}
			return ""
		},
		// 添加 last 函數，用於判斷是否為最後一個元素
		"last": func(index, length int) bool {
			return index == length-1
		},
		// 添加 severity_format 函數，用於格式化 severity 值
		"severity_format": func(severity interface{}) string {
			if s, ok := severity.(string); ok {
				// 去除前後空格
				s = strings.TrimSpace(s)
				switch strings.ToLower(s) {
				case "critical", "crit":
					return "Critical"
				case "warning", "warn":
					return "Warning"
				case "info":
					return "Info"
				default:
					return s
				}
			}
			return "Unknown"
		},
	})

	// 解析模板
	parsedTmpl, err := tmpl.Parse(jsonTemplate)
	if err != nil {
		return "", fmt.Errorf("解析 JSON 模板失敗: %w", err)
	}

	// 渲染模板
	var jsonBuf bytes.Buffer
	if err := parsedTmpl.Execute(&jsonBuf, data); err != nil {
		return "", fmt.Errorf("渲染 JSON 內容失敗: %w", err)
	}

	// 解析渲染後的 JSON 來確保格式正確
	var formattedJSON map[string]interface{}
	if err := json.Unmarshal(jsonBuf.Bytes(), &formattedJSON); err != nil {
		return "", fmt.Errorf("渲染後的 JSON 格式錯誤: %w", err)
	}

	// 美化輸出的 JSON
	formattedOutput, err := json.MarshalIndent(formattedJSON, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化 JSON 失敗: %w", err)
	}

	return string(formattedOutput), nil
}
