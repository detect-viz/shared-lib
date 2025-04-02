package alert

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"go.uber.org/zap"
)

// 通知狀態常量：
// NotifyStatePending - 等待處理
// NotifyStateSent - 已發送告警
// NotifyStateSolved - 已發送恢復
// NotifyStateProcessed - 已處理
// NotifyStateDelayed - 等待重試
// NotifyStateFailed - 發送失敗

// NotificationService 子函數說明：
// GetTriggeredLogs - 查詢未發送通知的 TriggeredLog
// GroupByContact - 按 ContactID 分組
// RenderTemplate - 依 FormatType 渲染模板 (HTML / Markdown / JSON / Text)
// SendNotification - 發送通知 (Webhook, Email, Slack)
// RecordNotifyLog - 記錄 NotifyLog
// RetryFailedNotifications - retry 機制 (RetryDelay & MaxRetry)

// 通知狀態常量
const (
	NotifyStatePending   = "pending"   // 等待處理
	NotifyStateSent      = "sent"      // 已發送告警
	NotifyStateSolved    = "solved"    // 已發送恢復
	NotifyStateProcessed = "processed" // 已處理
	NotifyStateDelayed   = "delayed"   // 等待重試
	NotifyStateFailed    = "failed"    // 發送失敗
)

// ErrorMessage 錯誤訊息結構
type ErrorMessage struct {
	Time    int64                  `json:"time"`
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// ProcessNotifyLog 處理通知日誌
func (s *Service) ProcessNotifyLog() error {
	s.logger.Info("開始處理通知日誌")

	// 1. 查詢需要發送通知的 TriggeredLog
	currentTime := time.Now().Unix()

	// 獲取待發送告警通知的日誌
	alertingLogs, err := s.mysql.GetPendingTriggeredLogs(currentTime)
	if err != nil {
		return fmt.Errorf("獲取待通知的 TriggeredLog 失敗: %w", err)
	}

	// 獲取待發送恢復通知的日誌
	resolvedLogs, err := s.mysql.GetResolvedTriggeredLogs(currentTime)
	if err != nil {
		return fmt.Errorf("獲取待發送恢復通知的 TriggeredLog 失敗: %w", err)
	}

	s.logger.Info("找到需要發送通知的告警",
		zap.Int("alerting_count", len(alertingLogs)),
		zap.Int("resolved_count", len(resolvedLogs)))

	// 2. 分別處理異常通知和恢復通知
	if err := s.processNotifications(alertingLogs, "alerting"); err != nil {
		s.logger.Error("處理異常通知失敗", zap.Error(err))
	}

	if err := s.processNotifications(resolvedLogs, "resolved"); err != nil {
		s.logger.Error("處理恢復通知失敗", zap.Error(err))
	}

	// 3. 處理需要重試的通知
	if err := s.retryFailedNotifications(); err != nil {
		s.logger.Error("重試失敗的通知時出錯", zap.Error(err))
	}

	return nil
}

// processNotifications 處理指定類型的通知
func (s *Service) processNotifications(logs []models.TriggeredLog, notifyType string) error {
	if len(logs) == 0 {
		return nil
	}

	// 1. 按 ContactID 分組

	groupedLogs := s.GroupByContact(logs)

	// 2. 處理每組通知

	for contactID, logs := range groupedLogs {
		// 獲取聯絡人信息
		contact, err := s.mysql.GetContact([]byte(contactID))
		if err != nil {
			s.logger.Error("獲取聯絡人信息失敗", zap.Error(err), zap.String("contact_id", formatID([]byte(contactID))))
			continue
		}

		if contact.ChannelType == "line" {
			config, err := s.contactService.GetConfig(context.Background(), contact.ChannelType)
			if err != nil {
				s.logger.Error("獲取聯絡人配置失敗", zap.Error(err), zap.String("contact_id", formatID([]byte(contactID))))
				continue
			}

			// 印出完整的 LINE 配置
			fmt.Println("LINE 通知配置:")
			for k, v := range config {
				if k == "channel_token" {
					fmt.Printf("  %s: %s\n", k, v[:min(10, len(v))]+"...")
				} else if k == "message" {
					fmt.Printf("  %s: %s...(省略)\n", k, v[:min(50, len(v))])
				} else {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}

			// 確保 token 格式正確
			if token, ok := config["channel_token"]; ok && token != "" {
				config["channel_token"] = strings.TrimSpace(token)
			}

			contact.Config["channel_token"] = config["channel_token"]
			s.logger.Info("聯絡人類型為 LINE",
				zap.String("contact_id", formatID([]byte(contactID))),
				zap.String("token_prefix", config["channel_token"][:min(10, len(config["channel_token"]))]+"..."))
		}

		if contact == nil || !contact.Enabled {
			s.logger.Warn("聯絡人不存在或已禁用", zap.String("contact_id", formatID([]byte(contactID))))
			continue
		}

		// 創建通知日誌
		notifyLog := s.createNotifyLog(contact, logs)

		// 3. 渲染模板
		title, message, err := s.renderTemplate(contact, logs, notifyType)
		if err != nil {
			s.logger.Error("渲染模板失敗", zap.Error(err), zap.String("contact_id", formatID([]byte(contactID))))
			notifyLog.State = NotifyStateFailed

			// 創建錯誤訊息
			errorMsg := fmt.Sprintf("渲染模板失敗: %s", err.Error())
			errorMessages := make(common.JSONMap)
			errorMessages["error"] = errorMsg
			notifyLog.ErrorMessages = &errorMessages

			if err := s.mysql.CreateNotifyLog(notifyLog); err != nil {
				s.logger.Error("記錄通知失敗日誌失敗", zap.Error(err))
			}
			continue
		}

		// 4. 發送通知
		err = s.sendNotification(contact, title, message)
		sentTime := time.Now().Unix()

		if err != nil {
			// 通知發送失敗
			s.logger.Error("發送通知失敗", zap.Error(err), zap.String("contact_id", formatID([]byte(contactID))))
			notifyLog.State = NotifyStateFailed

			// 創建錯誤訊息
			errorMsg := fmt.Sprintf("發送通知失敗: %s", err.Error())
			errorMessages := make(common.JSONMap)
			errorMessages["error"] = errorMsg
			notifyLog.ErrorMessages = &errorMessages
		} else {
			// 通知發送成功
			s.logger.Info("發送通知成功", zap.String("contact_id", formatID([]byte(contactID))))
			if notifyType == "alerting" {
				notifyLog.State = NotifyStateSent
			} else {
				notifyLog.State = NotifyStateSolved
			}
			notifyLog.SentAt = &sentTime
		}

		// 5. 記錄 NotifyLog
		if err := s.mysql.CreateNotifyLog(notifyLog); err != nil {
			s.logger.Error("記錄通知日誌失敗", zap.Error(err))
			continue
		}

		// 6. 更新 TriggeredLog 的通知狀態
		for _, log := range logs {
			var err error
			if notifyType == "alerting" {
				err = s.mysql.UpdateTriggeredLogNotifyState(log.ID, notifyLog.State)
			} else {
				err = s.mysql.UpdateTriggeredLogResolvedNotifyState(log.ID, notifyLog.State)
			}
			if err != nil {
				s.logger.Error("更新 TriggeredLog 通知狀態失敗",
					zap.Error(err),
					zap.String("triggered_log_id", formatID(log.ID)),
					zap.String("notify_type", notifyType))
			}
		}
	}

	return nil
}

// 檢查聯絡人是否應該接收該嚴重度的告警
func (s *Service) shouldNotifyContact(contact models.Contact, severity string) bool {
	for _, s := range contact.Severities {
		if s == severity {
			return true
		}
	}
	return false
}

// 創建通知日誌
func (s *Service) createNotifyLog(contact *models.Contact, logs []models.TriggeredLog) models.NotifyLog {
	// 創建 TriggeredLogIDs 列表
	triggeredLogIDs := make(models.TriggeredLogIDsMap, 0, len(logs))
	for _, log := range logs {
		// 使用原始的二進制 ID 的 base64 編碼，避免二進制數據在 JSON 中的問題
		idStr := base64.StdEncoding.EncodeToString(log.ID)
		triggeredLogIDs = append(triggeredLogIDs, map[string]interface{}{
			"id": idStr,
		})
	}

	// 創建聯絡人快照
	contactData, _ := json.Marshal(contact)
	var contactSnapshot common.JSONMap
	json.Unmarshal(contactData, &contactSnapshot)

	// 創建錯誤訊息存儲
	errorMessages := make(common.JSONMap)

	// 生成通知日誌
	return models.NotifyLog{
		RealmName:       contact.RealmName,
		State:           NotifyStatePending,
		RetryCounter:    0,
		TriggeredLogIDs: triggeredLogIDs,
		ContactID:       contact.ID,
		ChannelType:     contact.ChannelType,
		ContactSnapshot: contactSnapshot,
		ErrorMessages:   &errorMessages,
	}
}

// formatSeverity 將 severity 值轉換為標準格式
func formatSeverity(severity string) string {
	// 先去除前後空格
	severity = strings.TrimSpace(severity)

	switch strings.ToLower(severity) {
	case "critical", "crit":
		return "Critical"
	case "warning", "warn":
		return "Warning"
	case "info":
		return "Info"
	default:
		return severity
	}
}

// 渲染通知模板
func (s *Service) renderTemplate(contact *models.Contact, logs []models.TriggeredLog, notifyType string) (string, string, error) {

	// 獲取 FormatType
	formatType := GetFormatByType(contact.ChannelType)

	// 獲取通知模板
	var tmpl models.Template
	var found bool = false

	// 遍歷所有模板，查找匹配的模板
	for _, defaultTmpl := range s.global.Templates {
		if defaultTmpl.RuleState == notifyType && defaultTmpl.FormatType == formatType {
			tmpl = defaultTmpl
			found = true
			break
		}
	}

	if !found {
		s.logger.Error("找不到匹配的通知模板",
			zap.String("notify_type", notifyType),
			zap.String("format_type", formatType),
			zap.Int("templates_count", len(s.global.Templates)))
		return "", "", fmt.Errorf("找不到匹配的通知模板 [notify_type=%s, format_type=%s]", notifyType, formatType)
	}

	// 準備模板數據
	data := map[string]interface{}{
		"contact": map[string]interface{}{
			"name":         contact.Name,
			"channel_type": contact.ChannelType,
			"realm_name":   contact.RealmName,
		},
		"count": len(logs),
	}

	// 計算影響的主機數量和告警類型
	hostMap := make(map[string]bool)
	categoryMap := make(map[string]bool)
	severityMap := make(map[string]bool)

	for _, log := range logs {
		hostMap[log.ResourceName] = true
		categoryMap[log.MetricRuleUID] = true
		if log.Severity != "" {
			// 去除前後空格後再添加到 severityMap
			severityMap[strings.TrimSpace(log.Severity)] = true
		}
	}

	// 添加統計信息
	data["affected_hosts_count"] = len(hostMap)
	data["affected_alerts_count"] = len(logs)

	// 設置默認的 severity 值
	data["severity"] = "crit"

	// 如果只有一種 severity，則使用該值
	if len(severityMap) == 1 {
		for severity := range severityMap {
			data["severity"] = severity
			break
		}
	} else if len(severityMap) > 1 {
		// 如果有多種 severity，優先使用最高級別的
		severityOrder := map[string]int{
			"critical": 0, "crit": 0,
			"warning": 1, "warn": 1,
			"info": 2,
		}

		highestSeverity := "crit"
		highestOrder := 999

		for severity := range severityMap {
			severityLower := strings.ToLower(severity)
			if order, exists := severityOrder[severityLower]; exists && order < highestOrder {
				highestOrder = order
				highestSeverity = severity
			}
		}

		data["severity"] = highestSeverity
	}

	// 格式化 severity 值並添加到數據中
	data["severity_formatted"] = formatSeverity(data["severity"].(string))

	// 記錄 severity 值
	s.logger.Debug("使用的 severity 值",
		zap.String("severity", data["severity"].(string)),
		zap.String("severity_formatted", data["severity_formatted"].(string)))

	// 確保 realm_name 有值
	realmName := contact.RealmName
	if realmName == "" {
		realmName = "系統"
	}
	data["realm_name"] = realmName

	s.logger.Debug("affected_hosts_count", zap.Int("affected_hosts_count", len(hostMap)))
	s.logger.Debug("affected_alerts_count", zap.Int("affected_alerts_count", len(logs)))

	// 收集告警類別
	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}

	data["alert_categories"] = strings.Join(categories, ", ")

	s.logger.Debug("收集告警類別",
		zap.Int("category_count", len(categoryMap)),
		zap.String("alert_categories", strings.Join(categories, ", ")),
	)

	// 處理告警數據
	if notifyType == "alerting" {
		// 按嚴重程度分組
		severityGroups := make(map[string]map[string]interface{})
		for _, log := range logs {
			// 如果該嚴重程度不存在，創建它
			if _, exists := severityGroups[log.Severity]; !exists {
				severityGroups[log.Severity] = map[string]interface{}{
					"severity": log.Severity,
					"count":    0,
					"hosts":    make(map[string]map[string]interface{}),
				}
			}

			// 增加該嚴重程度的計數
			severityGroups[log.Severity]["count"] = severityGroups[log.Severity]["count"].(int) + 1

			// 獲取主機映射
			hostsMap := severityGroups[log.Severity]["hosts"].(map[string]map[string]interface{})

			// 如果該主機不存在，創建它
			if _, exists := hostsMap[log.ResourceName]; !exists {
				hostsMap[log.ResourceName] = map[string]interface{}{
					"resource_name": log.ResourceName,
					"count":         0,
					"metrics":       make([]map[string]interface{}, 0),
				}
			}

			// 增加該主機的計數
			hostsMap[log.ResourceName]["count"] = hostsMap[log.ResourceName]["count"].(int) + 1

			// 添加指標信息
			metricInfo := map[string]interface{}{
				"metric_display_name": log.MetricRuleUID,
				"triggered_value":     log.TriggeredValue,
				"threshold":           log.Threshold,
				"last_triggered_at":   time.Unix(log.LastTriggeredAt, 0).Format(time.RFC3339),
			}

			// 持續時間計算
			if log.LastTriggeredAt > log.TriggeredAt {
				duration := log.LastTriggeredAt - log.TriggeredAt
				metricInfo["duration"] = duration
			}

			// 將指標添加到主機的指標列表中
			metrics := hostsMap[log.ResourceName]["metrics"].([]map[string]interface{})
			hostsMap[log.ResourceName]["metrics"] = append(metrics, metricInfo)
		}

		// 將嚴重程度分組轉換為數組
		alertsBySeverity := make([]map[string]interface{}, 0, len(severityGroups))
		for _, severityGroup := range severityGroups {
			// 將主機映射轉換為數組
			hostsMap := severityGroup["hosts"].(map[string]map[string]interface{})
			hosts := make([]map[string]interface{}, 0, len(hostsMap))
			for _, host := range hostsMap {
				hosts = append(hosts, host)
			}
			severityGroup["hosts"] = hosts
			alertsBySeverity = append(alertsBySeverity, severityGroup)
		}

		// 按嚴重程度排序（Critical > Warning > Info）
		sort.Slice(alertsBySeverity, func(i, j int) bool {
			severityOrder := map[string]int{"Critical": 0, "Warning": 1, "Info": 2}
			severityI := alertsBySeverity[i]["severity"].(string)
			severityJ := alertsBySeverity[j]["severity"].(string)
			return severityOrder[severityI] < severityOrder[severityJ]
		})

		data["alerts_by_severity"] = alertsBySeverity

		// 為了向後兼容，保留舊的數據結構
		alertsByHost := make([]map[string]interface{}, 0, len(logs))
		for _, log := range logs {
			alertData := map[string]interface{}{
				"id":                  formatID(log.ID),
				"triggered_at":        time.Unix(log.TriggeredAt, 0).Format(time.RFC3339),
				"severity":            log.Severity,
				"resource_name":       log.ResourceName,
				"partition_name":      log.PartitionName,
				"metric_display_name": log.MetricRuleUID,
				"triggered_value":     log.TriggeredValue,
				"threshold":           log.Threshold,
			}
			alertsByHost = append(alertsByHost, alertData)
		}
		data["alerts_by_host"] = alertsByHost

	} else if notifyType == "resolved" {
		// 處理恢復通知
		data["resolved_alerts_count"] = len(logs)

		// 按嚴重程度分組
		severityGroups := make(map[string]map[string]interface{})
		for _, log := range logs {
			// 如果該嚴重程度不存在，創建它
			if _, exists := severityGroups[log.Severity]; !exists {
				severityGroups[log.Severity] = map[string]interface{}{
					"severity": log.Severity,
					"count":    0,
					"hosts":    make(map[string]map[string]interface{}),
				}
			}

			// 增加該嚴重程度的計數
			severityGroups[log.Severity]["count"] = severityGroups[log.Severity]["count"].(int) + 1

			// 獲取主機映射
			hostsMap := severityGroups[log.Severity]["hosts"].(map[string]map[string]interface{})

			// 如果該主機不存在，創建它
			if _, exists := hostsMap[log.ResourceName]; !exists {
				hostsMap[log.ResourceName] = map[string]interface{}{
					"resource_name": log.ResourceName,
					"count":         0,
					"metrics":       make([]map[string]interface{}, 0),
				}
			}

			// 增加該主機的計數
			hostsMap[log.ResourceName]["count"] = hostsMap[log.ResourceName]["count"].(int) + 1

			// 添加指標信息
			metricInfo := map[string]interface{}{
				"metric_display_name": log.MetricRuleUID,
				"previous_value":      log.TriggeredValue,
				"resolved_value":      log.ResolvedValue,
				"threshold":           log.Threshold,
				"resolved_at":         time.Unix(*log.ResolvedAt, 0).Format(time.RFC3339),
			}

			// 計算持續時間（如果有）
			if log.ResolvedAt != nil && log.TriggeredAt > 0 {
				duration := *log.ResolvedAt - log.TriggeredAt
				if duration > 0 {
					metricInfo["duration"] = duration
				}
			}

			// 將指標添加到主機的指標列表中
			metrics := hostsMap[log.ResourceName]["metrics"].([]map[string]interface{})
			hostsMap[log.ResourceName]["metrics"] = append(metrics, metricInfo)
		}

		// 將嚴重程度分組轉換為數組
		resolvedBySeverity := make([]map[string]interface{}, 0, len(severityGroups))
		for _, severityGroup := range severityGroups {
			// 將主機映射轉換為數組
			hostsMap := severityGroup["hosts"].(map[string]map[string]interface{})
			hosts := make([]map[string]interface{}, 0, len(hostsMap))
			for _, host := range hostsMap {
				hosts = append(hosts, host)
			}
			severityGroup["hosts"] = hosts
			resolvedBySeverity = append(resolvedBySeverity, severityGroup)
		}

		// 按嚴重程度排序（Critical > Warning > Info）
		sort.Slice(resolvedBySeverity, func(i, j int) bool {
			severityOrder := map[string]int{"Critical": 0, "Warning": 1, "Info": 2}
			severityI := resolvedBySeverity[i]["severity"].(string)
			severityJ := resolvedBySeverity[j]["severity"].(string)
			return severityOrder[severityI] < severityOrder[severityJ]
		})

		data["resolved_by_severity"] = resolvedBySeverity

		// 為了向後兼容，保留舊的數據結構
		resolvedAlertsByHost := make([]map[string]interface{}, 0, len(logs))
		for _, log := range logs {
			if log.ResolvedAt != nil {
				alertData := map[string]interface{}{
					"id":                  formatID(log.ID),
					"resource_name":       log.ResourceName,
					"partition_name":      log.PartitionName,
					"metric_display_name": log.MetricRuleUID,
					"previous_value":      log.TriggeredValue,
					"resolved_value":      log.ResolvedValue,
					"resolved_at":         time.Unix(*log.ResolvedAt, 0).Format(time.RFC3339),
				}
				resolvedAlertsByHost = append(resolvedAlertsByHost, alertData)
			}
		}
		data["resolved_alerts_by_host"] = resolvedAlertsByHost
	}

	// 透過 templates 模組渲染消息內容
	message, err := s.templateService.RenderMessage(tmpl, data)
	if err != nil {
		return "", "", fmt.Errorf("渲染模板失敗: %w", err)
	}

	// 處理消息格式
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

	// 如果是 LINE 通知，進行額外的格式處理
	if formatType == "text" && contact.ChannelType == "line" {
		// 移除 Markdown 中的特殊格式，LINE 不支援完整的 Markdown
		message = strings.ReplaceAll(message, "**", "")
		message = strings.ReplaceAll(message, "*", "")
		message = strings.ReplaceAll(message, "###", "")
		message = strings.ReplaceAll(message, "##", "")
		message = strings.ReplaceAll(message, "#", "")

		// 獲取格式化後的 severity 值
		formattedSeverity := data["severity_formatted"].(string)

		// 處理 severity 的顯示
		message = strings.ReplaceAll(message, "[{{ .severity }}]", "["+data["severity"].(string)+"]")
		message = strings.ReplaceAll(message, "[{{.severity}}]", "["+data["severity"].(string)+"]")
		message = strings.ReplaceAll(message, "[{{ .severity_format .severity }}]", "["+formattedSeverity+"]")
		message = strings.ReplaceAll(message, "[{{.severity_format .severity}}]", "["+formattedSeverity+"]")
		message = strings.ReplaceAll(message, "[{{ .severity_formatted }}]", "["+formattedSeverity+"]")
		message = strings.ReplaceAll(message, "[{{.severity_formatted}}]", "["+formattedSeverity+"]")

		// 處理其他可能的 severity 格式
		severityPatterns := []string{
			"[{{ severity }}]", "[{{severity}}]",
			"[{{ .Severity }}]", "[{{.Severity}}]",
			"[{{ severity_format severity }}]", "[{{severity_format severity}}]",
			"[{{ severity_format .severity }}]", "[{{severity_format .severity}}]",
		}

		for _, pattern := range severityPatterns {
			message = strings.ReplaceAll(message, pattern, "["+formattedSeverity+"]")
		}

		// 確保換行符號正確
		message = strings.ReplaceAll(message, "\n\n\n", "\n\n")
	}

	// 渲染標題
	titleTmpl, err := template.New("title").Parse(tmpl.Title)
	if err != nil {
		return "", "", fmt.Errorf("解析標題模板失敗: %w", err)
	}

	var titleBuf bytes.Buffer
	if err := titleTmpl.Execute(&titleBuf, data); err != nil {
		return "", "", fmt.Errorf("渲染標題失敗: %w", err)
	}

	title := titleBuf.String()
	s.logger.Debug("渲染標題結果", zap.String("title", title))

	return title, message, nil
}

func GetFormatByType(contactType string) string {
	switch contactType {
	case "email":
		return "html"
	case "slack", "discord", "teams", "webex":
		return "markdown"
	case "webhook":
		return "json"
	case "line":
		return "text"
	default:
		return "text"
	}
}

// 發送通知
func (s *Service) sendNotification(contact *models.Contact, title, message string) error {
	newConfig := contact.Config
	newConfig["title"] = title
	newConfig["message"] = message

	// 確保 LINE 通知有必要的配置
	if contact.ChannelType == "line" {

		// 確保 to 字段存在
		if newConfig["to"] == "" {
			return fmt.Errorf("LINE 通知缺少接收者 ID (to)")
		}

		// 使用 contact.Config 中的 channel_token
		if token, ok := newConfig["channel_token"]; ok && token != "" {
			// 確保 token 格式正確
			token = strings.TrimSpace(token)
			newConfig["channel_token"] = token
			s.logger.Info("使用聯絡人配置的 LINE token",
				zap.String("token_prefix", token[:min(10, len(token))]+"..."))
		} else {
			s.logger.Error("未設置 LINE token，LINE 通知將會失敗")
			return fmt.Errorf("未設置 LINE token，無法發送 LINE 通知")
		}
	}

	// 發送通知
	return s.notifyService.Send(common.NotifySetting{
		Type:   contact.ChannelType,
		Config: newConfig,
	})
}

// min 返回兩個整數中的較小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 配置常量
const (
	DefaultMaxRetry    = 3   // 最大重試次數
	DefaultRetryDelay  = 300 // 重試延遲時間（秒）
	DefaultPendingTime = 600 // 通知等待時間（秒）
)

// HandleNotifyPendingTime 處理通知等待時間
func (s *Service) HandleNotifyPendingTime(triggeredLog *models.TriggeredLog) error {
	// 檢查是否需要等待
	currentTime := time.Now().Unix()
	pendingTime := DefaultPendingTime

	// 使用配置中的等待時間
	if s.config.NotifyPeriod > 0 {
		pendingTime = s.config.NotifyPeriod
	}

	// 如果觸發時間 + 等待時間 > 當前時間，則不需要發送通知
	if triggeredLog.TriggeredAt+int64(pendingTime) > currentTime {
		return nil
	}

	// 更新通知狀態為待發送
	return s.mysql.UpdateTriggeredLogNotifyState(triggeredLog.ID, NotifyStatePending)
}

// retryFailedNotifications 重試失敗的通知
func (s *Service) retryFailedNotifications() error {
	s.logger.Info("開始處理失敗的通知")

	// 獲取所有失敗的通知記錄
	failedLogs, err := s.mysql.GetFailedNotifyLogs()
	if err != nil {
		return fmt.Errorf("獲取失敗的通知記錄失敗: %w", err)
	}

	if len(failedLogs) == 0 {
		s.logger.Info("沒有需要重試的失敗通知")
		return nil
	}

	// 獲取配置
	maxRetry := DefaultMaxRetry
	retryDelay := DefaultRetryDelay

	// 使用聯絡人的重試次數
	for _, notifyLog := range failedLogs {
		// 檢查是否超過最大重試次數
		if notifyLog.RetryCounter >= maxRetry {
			s.logger.Warn("通知已超過最大重試次數，標記為最終失敗",
				zap.String("notify_log_id", formatID(notifyLog.ID)),
				zap.Int("retry_counter", notifyLog.RetryCounter))

			notifyLog.State = NotifyStateFailed

			// 創建錯誤訊息
			errorMessages := make(common.JSONMap)
			errorMessages["error"] = fmt.Sprintf("超過最大重試次數 %d", maxRetry)
			errorMessages["time"] = fmt.Sprintf("%d", time.Now().Unix())
			notifyLog.ErrorMessages = &errorMessages

			if err := s.mysql.UpdateNotifyLog(notifyLog); err != nil {
				s.logger.Error("更新通知日誌狀態失敗", zap.Error(err))
			}
			continue
		}

		// 檢查是否達到重試延遲時間
		currentTime := time.Now().Unix()
		if notifyLog.LastRetryAt != nil && currentTime-*notifyLog.LastRetryAt < int64(retryDelay) {
			continue
		}

		// 重新發送通知
		contact, err := s.mysql.GetContact(notifyLog.ContactID)
		if err != nil {
			s.logger.Error("獲取聯絡人信息失敗", zap.Error(err))
			continue
		}

		// 如果是 LINE 通知，獲取 channel_token
		if contact.ChannelType == "line" {
			config, err := s.contactService.GetConfig(context.Background(), contact.ChannelType)
			if err != nil {
				s.logger.Error("獲取聯絡人配置失敗", zap.Error(err), zap.String("contact_id", formatID(notifyLog.ContactID)))
				continue
			}

			// 印出完整的 LINE 配置
			fmt.Println("重試時的 LINE 通知配置:")
			for k, v := range config {
				if k == "channel_token" {
					fmt.Printf("  %s: %s\n", k, v[:min(10, len(v))]+"...")
				} else if k == "message" {
					fmt.Printf("  %s: %s...(省略)\n", k, v[:min(50, len(v))])
				} else {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}

			// 確保 token 格式正確
			if token, ok := config["channel_token"]; ok && token != "" {
				config["channel_token"] = strings.TrimSpace(token)
			}

			contact.Config["channel_token"] = config["channel_token"]
			s.logger.Info("聯絡人類型為 LINE",
				zap.String("contact_id", formatID(notifyLog.ContactID)),
				zap.String("token_prefix", config["channel_token"][:min(10, len(config["channel_token"]))]+"..."))
		}

		// 重新渲染模板
		var triggeredLogs []models.TriggeredLog
		for _, logIDMap := range notifyLog.TriggeredLogIDs {
			id, ok := logIDMap["id"].(string)
			if !ok {
				continue
			}

			// 嘗試從 base64 解碼 ID
			binaryID, err := base64.StdEncoding.DecodeString(id)
			if err != nil {
				// 如果 base64 解碼失敗，嘗試從十六進制解碼（向後兼容）
				binaryID, err = hex.DecodeString(id)
				if err != nil {
					s.logger.Error("解析 ID 失敗", zap.Error(err), zap.String("id", id))
					continue
				}
			}

			log, err := s.mysql.GetTriggeredLog(binaryID)
			if err != nil {
				s.logger.Error("獲取觸發日誌失敗", zap.Error(err))
				continue
			}
			if log != nil {
				triggeredLogs = append(triggeredLogs, *log)
			}
		}

		// 判斷通知類型
		notifyType := "alerting"
		if notifyLog.State == NotifyStateSolved {
			notifyType = "resolved"
		}

		title, message, err := s.renderTemplate(contact, triggeredLogs, notifyType)
		if err != nil {
			s.logger.Error("重試時渲染模板失敗", zap.Error(err))
			continue
		}

		// 重新發送通知
		err = s.sendNotification(contact, title, message)
		retryTime := time.Now().Unix()

		// 更新通知日誌
		notifyLog.RetryCounter++
		notifyLog.LastRetryAt = &retryTime

		if err != nil {
			s.logger.Error("重試發送通知失敗",
				zap.Error(err),
				zap.String("notify_log_id", formatID(notifyLog.ID)),
				zap.Int("retry_counter", notifyLog.RetryCounter))

			notifyLog.State = NotifyStateFailed

			// 創建錯誤訊息
			errorMessages := make(common.JSONMap)
			errorMessages["error"] = fmt.Sprintf("重試發送通知失敗: %s", err.Error())
			errorMessages["retry"] = fmt.Sprintf("%d", notifyLog.RetryCounter)
			errorMessages["time"] = fmt.Sprintf("%d", time.Now().Unix())
			notifyLog.ErrorMessages = &errorMessages
		} else {
			s.logger.Info("重試發送通知成功",
				zap.String("notify_log_id", formatID(notifyLog.ID)),
				zap.Int("retry_counter", notifyLog.RetryCounter))

			notifyLog.State = NotifyStateSent
			notifyLog.SentAt = &retryTime
		}

		if err := s.mysql.UpdateNotifyLog(notifyLog); err != nil {
			s.logger.Error("更新通知日誌失敗", zap.Error(err))
		}
	}

	return nil
}

// GroupByContact 將觸發日誌按聯絡人分組
func (s *Service) GroupByContact(triggeredLogs []models.TriggeredLog) map[string][]models.TriggeredLog {
	groupedLogs := make(map[string][]models.TriggeredLog)

	for _, log := range triggeredLogs {
		// 獲取規則關聯的聯絡人
		s.logger.Debug("獲取規則關聯的聯絡人", zap.String("rule_id", formatID(log.RuleID)))
		contacts, err := s.mysql.GetContactsByRuleID(log.RuleID)
		if err != nil {
			s.logger.Error("獲取規則關聯的聯絡人失敗",
				zap.Error(err),
				zap.String("rule_id", formatID(log.RuleID)))
			continue
		}

		// 按嚴重度篩選聯絡人
		for _, contact := range contacts {
			if !contact.Enabled {
				continue
			}

			if !s.shouldNotifyContact(contact, log.Severity) {
				continue
			}

			contactID := string(contact.ID)
			groupedLogs[contactID] = append(groupedLogs[contactID], log)
		}
	}

	return groupedLogs
}

// 添加一個輔助函數來格式化 ID
func formatID(id []byte) string {
	if len(id) == 0 {
		return "empty_id"
	}
	return hex.EncodeToString(id)
}
