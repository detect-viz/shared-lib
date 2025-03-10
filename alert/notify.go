package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationService 子函數說明：
// GetTriggeredLogs - 查詢未發送通知的 TriggeredLog
// GroupByContact - 按 ContactID 分組
// RenderTemplate - 依 FormatType 渲染模板 (HTML / Markdown / JSON / Text)
// SendNotification - 發送通知 (Webhook, Email, Slack)
// RecordNotifyLog - 記錄 NotifyLog
// RetryFailedNotifications - retry 機制 (RetryDelay & MaxRetry)

// ProcessNotifyLog 處理通知日誌
func (s *Service) ProcessNotifyLog() error {
	s.logger.Info("開始處理通知日誌")

	// 1. 查詢未發送通知的 TriggeredLog
	currentTime := time.Now().Unix()
	triggeredLogs, err := s.mysql.GetTriggeredLogsForAlertNotify(currentTime)
	if err != nil {
		return fmt.Errorf("獲取待通知的 TriggeredLog 失敗: %w", err)
	}

	if len(triggeredLogs) == 0 {
		s.logger.Info("沒有需要發送通知的告警")
		return nil
	}

	// 讓 TriggeredLog.notify_state = processed，避免重複發送
	for _, log := range triggeredLogs {
		log.NotifyState = "processed"
		if err := s.mysql.UpdateTriggeredLog(log); err != nil {
			s.logger.Error("更新 TriggeredLog 通知狀態失敗", zap.Error(err), zap.String("triggered_log_id", string(log.ID)))
		}
	}

	s.logger.Info("找到需要發送通知的告警", zap.Int("count", len(triggeredLogs)))

	// 2. 按 ContactID 分組
	groupedLogs := make(map[string][]models.TriggeredLog)

	for _, log := range triggeredLogs {
		// 獲取規則關聯的聯絡人
		contacts, err := s.mysql.GetContactsByRuleID(log.RuleID)
		if err != nil {
			s.logger.Error("獲取規則關聯的聯絡人失敗", zap.Error(err), zap.String("rule_id", string(log.RuleID)))
			continue
		}

		// 按嚴重度篩選聯絡人
		for _, contact := range contacts {
			// 檢查聯絡人是否啟用
			if !contact.Enabled {
				continue
			}

			// 檢查聯絡人是否接收該嚴重度的告警
			if !s.shouldNotifyContact(contact, log.Severity) {
				continue
			}

			// 將 TriggeredLog 添加到對應聯絡人的分組中
			contactID := string(contact.ID)
			groupedLogs[contactID] = append(groupedLogs[contactID], log)
		}
	}

	// 3. 處理每組通知
	for contactID, logs := range groupedLogs {
		// 獲取聯絡人信息
		contact, err := s.mysql.GetContact([]byte(contactID))
		if err != nil {
			s.logger.Error("獲取聯絡人信息失敗", zap.Error(err), zap.String("contact_id", contactID))
			continue
		}

		if contact == nil || !contact.Enabled {
			s.logger.Warn("聯絡人不存在或已禁用", zap.String("contact_id", contactID))
			continue
		}

		// 創建通知日誌
		notifyLog := s.createNotifyLog(contact, logs)

		// 4. 渲染模板
		title, message, err := s.renderTemplate(contact, logs)
		if err != nil {
			s.logger.Error("渲染模板失敗", zap.Error(err), zap.String("contact_id", contactID))
			notifyLog.State = "failed"
			errorMsg := fmt.Sprintf("渲染模板失敗: %s", err.Error())
			notifyLog.ErrorMessage = &errorMsg
			if err := s.mysql.CreateNotifyLog(notifyLog); err != nil {
				s.logger.Error("記錄通知失敗日誌失敗", zap.Error(err))
			}
			continue
		}

		// 5. 發送通知
		err = s.sendNotification(contact, title, message)
		sentTime := time.Now().Unix()

		if err != nil {
			// 通知發送失敗
			s.logger.Error("發送通知失敗", zap.Error(err), zap.String("contact_id", contactID))
			notifyLog.State = "failed"
			errorMsg := fmt.Sprintf("發送通知失敗: %s", err.Error())
			notifyLog.ErrorMessage = &errorMsg
		} else {
			// 通知發送成功
			s.logger.Info("發送通知成功", zap.String("contact_id", contactID))
			notifyLog.State = "sent"
			notifyLog.SentAt = &sentTime
		}

		// 6. 記錄 NotifyLog
		if err := s.mysql.CreateNotifyLog(notifyLog); err != nil {
			s.logger.Error("記錄通知日誌失敗", zap.Error(err))
			continue
		}

		// 更新 TriggeredLog 的通知狀態
		for _, log := range logs {
			if err := s.mysql.UpdateTriggeredLogNotifyState(log.ID, notifyLog.State); err != nil {
				s.logger.Error("更新 TriggeredLog 通知狀態失敗", zap.Error(err), zap.String("triggered_log_id", string(log.ID)))
			}
		}
	}

	// 7. 處理需要重試的通知
	if err := s.retryFailedNotifications(); err != nil {
		s.logger.Error("重試失敗的通知時出錯", zap.Error(err))
	}

	return nil
}

// shouldNotifyContact 檢查聯絡人是否應該接收該嚴重度的告警
func (s *Service) shouldNotifyContact(contact models.Contact, severity string) bool {
	for _, s := range contact.Severities {
		if s == severity {
			return true
		}
	}
	return false
}

// createNotifyLog 創建通知日誌
func (s *Service) createNotifyLog(contact *models.Contact, logs []models.TriggeredLog) models.NotifyLog {
	// 創建 TriggeredLogIDs 列表
	triggeredLogIDs := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		triggeredLogIDs = append(triggeredLogIDs, map[string]interface{}{
			"id": string(log.ID),
		})
	}

	// 創建聯絡人快照
	contactData, _ := json.Marshal(contact)
	var contactSnapshot common.JSONMap
	json.Unmarshal(contactData, &contactSnapshot)

	// 生成通知日誌
	id := uuid.New()
	return models.NotifyLog{
		RealmName:       contact.RealmName,
		ID:              id[:],
		State:           "pending",
		RetryCounter:    0,
		TriggeredLogIDs: triggeredLogIDs,
		ContactID:       contact.ID,
		ChannelType:     contact.ChannelType,
		ContactSnapshot: contactSnapshot,
	}
}

// renderTemplate 渲染通知模板
func (s *Service) renderTemplate(contact *models.Contact, logs []models.TriggeredLog) (string, string, error) {
	// 獲取適用的模板
	tmpl, err := s.mysql.GetTemplate(contact.RealmName, "alerting", contact.ChannelType)
	if err != nil {
		return "", "", fmt.Errorf("獲取模板失敗: %w", err)
	}

	// 準備模板數據
	data := map[string]interface{}{
		"contact": map[string]interface{}{
			"name":         contact.Name,
			"channel_type": contact.ChannelType,
			"realm_name":   contact.RealmName,
		},
		"alerts": make([]map[string]interface{}, 0, len(logs)),
		"count":  len(logs),
	}

	// 處理每個告警的數據
	for _, log := range logs {
		alertData := map[string]interface{}{
			"id":              string(log.ID),
			"triggered_at":    time.Unix(log.TriggeredAt, 0).Format(time.RFC3339),
			"severity":        log.Severity,
			"resource_name":   log.ResourceName,
			"partition_name":  log.PartitionName,
			"triggered_value": log.TriggeredValue,
			"threshold":       log.Threshold,
		}

		// 如果已解決，添加解決時間
		if log.ResolvedAt != nil {
			alertData["resolved_at"] = time.Unix(*log.ResolvedAt, 0).Format(time.RFC3339)
		}

		data["alerts"] = append(data["alerts"].([]map[string]interface{}), alertData)
	}

	// 根據通知渠道選擇合適的模板格式
	var titleTmpl, messageTmpl *template.Template
	var err1, err2 error

	switch contact.ChannelType {
	case "email":
		titleTmpl, err1 = template.New("title").Parse(tmpl.Title)
		messageTmpl, err2 = template.New("message").Parse(tmpl.Message)
	case "slack":
		// Slack 使用 Markdown 格式
		titleTmpl, err1 = template.New("title").Parse(tmpl.Title)
		messageTmpl, err2 = template.New("message").Funcs(template.FuncMap{
			"bold":  func(s string) string { return "*" + s + "*" },
			"code":  func(s string) string { return "`" + s + "`" },
			"quote": func(s string) string { return ">" + s },
			"timestamp": func(t int64) string {
				return fmt.Sprintf("<!date^%d^{date_num} {time_secs}|%s>", t, time.Unix(t, 0).Format(time.RFC3339))
			},
		}).Parse(tmpl.Message)
	case "webhook":
		// Webhook 使用純文本格式
		titleTmpl, err1 = template.New("title").Parse(tmpl.Title)
		messageTmpl, err2 = template.New("message").Parse(tmpl.Message)
	default:
		return "", "", fmt.Errorf("不支持的通知渠道類型: %s", contact.ChannelType)
	}

	if err1 != nil {
		return "", "", fmt.Errorf("解析標題模板失敗: %w", err1)
	}
	if err2 != nil {
		return "", "", fmt.Errorf("解析消息模板失敗: %w", err2)
	}

	// 渲染標題
	var titleBuf bytes.Buffer
	if err := titleTmpl.Execute(&titleBuf, data); err != nil {
		return "", "", fmt.Errorf("渲染標題失敗: %w", err)
	}

	// 渲染消息內容
	var messageBuf bytes.Buffer
	if err := messageTmpl.Execute(&messageBuf, data); err != nil {
		return "", "", fmt.Errorf("渲染消息內容失敗: %w", err)
	}

	return titleBuf.String(), messageBuf.String(), nil
}

// sendNotification 發送通知
func (s *Service) sendNotification(contact *models.Contact, title, message string) error {
	// 根據聯絡人類型選擇不同的通知方式
	switch contact.ChannelType {
	case "webhook":
		return s.sendWebhook(contact, title, message)
	case "email":
		return s.sendEmail(contact, title, message)
	case "slack":
		return s.sendSlack(contact, title, message)
	default:
		return fmt.Errorf("不支持的通知渠道類型: %s", contact.ChannelType)
	}
}

// sendWebhook 發送 Webhook 通知
func (s *Service) sendWebhook(contact *models.Contact, title, message string) error {
	// 這裡應該實現 Webhook 通知的邏輯
	// 例如使用 HTTP 客戶端發送 POST 請求
	s.logger.Info("模擬發送 Webhook 通知", zap.String("contact_id", string(contact.ID)), zap.String("title", title))
	return nil
}

// sendEmail 發送郵件通知
func (s *Service) sendEmail(contact *models.Contact, title, message string) error {
	// 這裡應該實現郵件通知的邏輯
	// 例如使用 SMTP 客戶端發送郵件
	s.logger.Info("模擬發送郵件通知", zap.String("contact_id", string(contact.ID)), zap.String("title", title))
	return nil
}

// sendSlack 發送 Slack 通知
func (s *Service) sendSlack(contact *models.Contact, title, message string) error {
	// 這裡應該實現 Slack 通知的邏輯
	// 例如使用 Slack API 發送消息
	s.logger.Info("模擬發送 Slack 通知", zap.String("contact_id", string(contact.ID)), zap.String("title", title))
	return nil
}

// 配置常量
const (
	maxRetry          = 3   // 最大重試次數
	retryDelay        = 300 // 重試延遲時間（秒）
	notifyPendingTime = 600 // 通知等待時間（秒）
)

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

	for _, notifyLog := range failedLogs {
		// 檢查是否超過最大重試次數
		if notifyLog.RetryCounter >= maxRetry {
			s.logger.Warn("通知已超過最大重試次數，標記為最終失敗",
				zap.String("notify_log_id", string(notifyLog.ID)),
				zap.Int("retry_counter", notifyLog.RetryCounter))

			notifyLog.State = "final_failed"
			if err := s.mysql.UpdateNotifyLog(notifyLog); err != nil {
				s.logger.Error("更新通知日誌狀態失敗", zap.Error(err))
			}
			continue
		}

		// 檢查是否達到重試延遲時間
		currentTime := time.Now().Unix()
		if notifyLog.LastRetryAt != nil && currentTime-*notifyLog.LastRetryAt < retryDelay {
			continue
		}

		// 重新發送通知
		contact, err := s.mysql.GetContact(notifyLog.ContactID)
		if err != nil {
			s.logger.Error("獲取聯絡人信息失敗", zap.Error(err))
			continue
		}

		// 重新渲染模板
		var triggeredLogs []models.TriggeredLog
		for _, logID := range notifyLog.TriggeredLogIDs {
			idValue, exists := logID["id"]
			if !exists {
				continue
			}

			// 嘗試將 interface{} 轉換為字符串
			var idStr string
			switch v := idValue.(type) {
			case string:
				idStr = v
			case []byte:
				idStr = string(v)
			case json.Number:
				idStr = string(v)
			default:
				s.logger.Error("無效的 ID 類型",
					zap.String("notify_log_id", string(notifyLog.ID)),
					zap.Any("id_type", fmt.Sprintf("%T", idValue)))
				continue
			}

			log, err := s.mysql.GetTriggeredLog([]byte(idStr))
			if err != nil {
				s.logger.Error("獲取觸發日誌失敗", zap.Error(err))
				continue
			}
			if log != nil {
				triggeredLogs = append(triggeredLogs, *log)
			}
		}

		title, message, err := s.renderTemplate(contact, triggeredLogs)
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
				zap.String("notify_log_id", string(notifyLog.ID)),
				zap.Int("retry_counter", notifyLog.RetryCounter))

			notifyLog.State = "failed"
			errorMsg := fmt.Sprintf("重試發送通知失敗: %s", err.Error())
			notifyLog.ErrorMessage = &errorMsg
		} else {
			s.logger.Info("重試發送通知成功",
				zap.String("notify_log_id", string(notifyLog.ID)),
				zap.Int("retry_counter", notifyLog.RetryCounter))

			notifyLog.State = "sent"
			notifyLog.SentAt = &retryTime
		}

		if err := s.mysql.UpdateNotifyLog(notifyLog); err != nil {
			s.logger.Error("更新通知日誌失敗", zap.Error(err))
		}
	}

	return nil
}

// HandleNotifyPendingTime 處理延遲通知機制
func (s *Service) HandleNotifyPendingTime(triggeredLog *models.TriggeredLog) error {
	// 如果沒有設置延遲時間，直接返回
	if notifyPendingTime <= 0 {
		return nil
	}

	currentTime := time.Now().Unix()
	pendingEndTime := triggeredLog.TriggeredAt + notifyPendingTime

	// 如果還在等待期內
	if currentTime < pendingEndTime {
		// 檢查告警是否已恢復
		ruleState, err := s.mysql.GetRuleState(triggeredLog.RuleID)
		if err != nil {
			return fmt.Errorf("獲取規則狀態失敗: %w", err)
		}

		// 如果告警已恢復，標記為忽略
		if ruleState != nil && ruleState.State == "resolved" {
			triggeredLog.NotifyState = "ignored"
			if err := s.mysql.UpdateTriggeredLog(*triggeredLog); err != nil {
				return fmt.Errorf("更新觸發日誌狀態失敗: %w", err)
			}
			return nil
		}

		// 還在等待期內，且告警未恢復，保持等待
		return nil
	}

	// 延遲期結束，將狀態改為待通知
	triggeredLog.NotifyState = "pending"
	if err := s.mysql.UpdateTriggeredLog(*triggeredLog); err != nil {
		return fmt.Errorf("更新觸發日誌狀態失敗: %w", err)
	}

	return nil
}

// GroupByContact 將觸發日誌按聯絡人分組
func (s *Service) GroupByContact(triggeredLogs []models.TriggeredLog) map[string][]models.TriggeredLog {
	groupedLogs := make(map[string][]models.TriggeredLog)

	for _, log := range triggeredLogs {
		// 獲取規則關聯的聯絡人
		contacts, err := s.mysql.GetContactsByRuleID(log.RuleID)
		if err != nil {
			s.logger.Error("獲取規則關聯的聯絡人失敗",
				zap.Error(err),
				zap.String("rule_id", string(log.RuleID)))
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
