package alert

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
)

func (s *Service) getMetricCategory(dataSource, metricName string) string {
	s.logger.Debug("getMetricCategory",
		zap.String("dataSource", dataSource),
		zap.String("metricName", metricName))

	// 檢查 MetricRules 是否為空
	if len(s.global.MetricRules) == 0 {
		s.logger.Warn("MetricRules 為空，無法匹配 category")
		return "unknown"
	}

	s.logger.Debug("開始匹配 category",
		zap.Int("rules_count", len(s.global.MetricRules)))

	for uid, rule := range s.global.MetricRules {
		s.logger.Debug("檢查規則",
			zap.String("uid", uid),
			zap.String("metric_raw_name", rule.MetricRawName),
			zap.Strings("match_datasource_names", rule.MatchDatasourceNames),
			zap.String("category", rule.Category))

		// 檢查 MetricRawName 是否匹配
		if rule.MetricRawName == metricName {
			s.logger.Debug("MetricRawName 匹配成功",
				zap.String("metric_raw_name", rule.MetricRawName))

			// 檢查 MatchDatasourceNames 是否為空或包含 dataSource
			if len(rule.MatchDatasourceNames) == 0 {
				s.logger.Debug("MatchDatasourceNames 為空，直接返回 category",
					zap.String("category", rule.Category))
				return rule.Category
			} else if slices.Contains(rule.MatchDatasourceNames, dataSource) {
				s.logger.Debug("MatchDatasourceNames 包含 dataSource，返回 category",
					zap.String("dataSource", dataSource),
					zap.String("category", rule.Category))
				return rule.Category
			} else {
				s.logger.Debug("MatchDatasourceNames 不包含 dataSource，繼續檢查下一個規則",
					zap.String("dataSource", dataSource),
					zap.Strings("match_datasource_names", rule.MatchDatasourceNames))
			}
		}
	}

	s.logger.Warn("未找到匹配的 category，返回 unknown",
		zap.String("dataSource", dataSource),
		zap.String("metricName", metricName))
	return "unknown"
}

func (s *Service) matchAutoApplyRule(realm, dataSource, metricName string) *[]models.Rule {
	matchMetricRuleUIDs := []string{}
	for _, rule := range s.global.MetricRules {
		if rule.MetricRawName == metricName &&
			(len(rule.MatchDatasourceNames) == 0 || slices.Contains(rule.MatchDatasourceNames, dataSource)) {
			matchMetricRuleUIDs = append(matchMetricRuleUIDs, rule.UID)
		}
	}
	if len(matchMetricRuleUIDs) == 0 {
		return nil
	}

	rules, err := s.mysql.GetAutoApplyRulesByMetricRuleUIDs(realm, matchMetricRuleUIDs)
	if err != nil {
		s.logger.Error("獲取告警規則失敗", zap.Error(err))
		return nil
	}

	if len(rules) == 0 {
		s.logger.Debug("沒有可以自動新增匹配的規則", zap.String("realm", realm), zap.String("datasource", dataSource), zap.String("metric", metricName))
	}

	return &rules
}

// lockFile 鎖定檔案
func (s *Service) LockFile(path string) error {
	lockPath := path + ".lock"
	for i := 0; i < 3; i++ { // 重試3次
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			// 創建鎖檔案
			if err := os.WriteFile(lockPath, []byte{}, 0644); err != nil {
				return err
			}
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("檔案已被鎖定: %s", path)
}

// 解除檔案鎖定
func (s *Service) UnlockFile(path string) error {
	return os.Remove(path + ".lock")
}

// 帶鎖寫入檔案
func (s *Service) WriteWithLock(path string, data []byte) error {
	if err := s.LockFile(path); err != nil {
		return err
	}
	defer s.UnlockFile(path)

	return os.WriteFile(path, data, 0644)
}

func (s *Service) GetLogger() *zap.Logger {
	return s.logger.GetLogger()
}

// 解析 Duration (e.g., "5m" → 300)
func (s *Service) parseDuration(input string) (int, error) {
	unitMap := map[string]time.Duration{
		"s": time.Second,
		"m": time.Minute,
		"h": time.Hour,
		"d": time.Hour * 24,
	}

	// 解析數字 + 單位
	for unit, duration := range unitMap {
		if strings.HasSuffix(input, unit) {
			valueStr := strings.TrimSuffix(input, unit)
			value, err := strconv.Atoi(valueStr)
			if err != nil {
				return 0, fmt.Errorf("invalid duration format")
			}
			return int(time.Duration(value) * duration / time.Second), nil
		}
	}

	return 0, fmt.Errorf("unsupported time unit")
}

// ================================

//* ======================== 5.triggered_log.go 觸發日誌 ========================

// // 根據通知管道分組 [contact.ID + contact.Type + rule.RuleState]
// func (s *Service) groupTriggeredLogs(triggereds []models.TriggeredLog, isResolved bool) map[string][]models.TriggeredLog {
// 	groups := make(map[string][]models.TriggeredLog)

// 	if len(triggereds) == 0 {
// 		return groups
// 	}

// 	for _, triggered := range triggereds {
// 		var ruleState string
// 		if triggered.ResolvedTime != nil {
// 			ruleState = "resolved"
// 		} else {
// 			ruleState = "alerting"
// 		}
// 		for _, contact := range triggered.Contacts {
// 			if isResolved && !contact.SendResolved {
// 				continue
// 			}
// 			key := fmt.Sprintf("%d_%s_%s", contact.ID, contact.Type, ruleState)
// 			groups[key] = append(groups[key], triggered)
// 		}
// 	}
// 	return groups
// }

// func (s *Service) writeResolvedLog(rule models.Rule, state models.RuleState) error {
// 	// 確保已經有 TriggeredLog
// 	exists, err := s.mysql.CheckTriggeredLogExists(rule.RuleID, rule.ResourceName, rule.MetricName, state.FirstTriggeredTime)
// 	if err != nil {
// 		return err
// 	}
// 	if !exists {
// 		s.logger.Warn("找不到對應的 TriggeredLog，無法寫入 ResolvedLog",
// 			zap.String("rule_id", rule.RuleID),
// 			zap.String("resource", rule.ResourceName),
// 			zap.String("metric", rule.MetricName))
// 		return nil
// 	}

// 	// **更新 TriggeredLog 狀態**
// 	err = s.mysql.UpdateTriggeredLogResolved(rule.RuleID, rule.ResourceName, rule.MetricName, state.LastTriggeredTime)
// 	if err != nil {
// 		s.logger.Error("更新 TriggeredLog 為 resolved 失敗",
// 			zap.String("rule_id", rule.RuleID),
// 			zap.String("resource", rule.ResourceName),
// 			zap.String("metric", rule.MetricName),
// 			zap.Error(err))
// 		return err
// 	}

// 	return nil
// }

//* ======================== 6.notify_log.go 通知日誌 ========================

// // * 創建通知日誌 [TriggeredLog 轉換為 NotifyLog]
// func (s *Service) generateNotifyLog(key string, triggereds []models.TriggeredLog) models.NotifyLog {
// 	now := time.Now().Unix()

// 	// 解析聯絡人資訊
// 	parts := strings.Split(key, "_")
// 	contactID, _ := strconv.ParseInt(parts[0], 10, 64)
// 	contactType := parts[1]
// 	ruleState := parts[2]
// 	contact, err := s.contactService.Get(contactID)
// 	if err != nil {
// 		s.logger.Error("獲取聯絡人資訊失敗", zap.Error(err))
// 	}
// 	notifyFormat := GetFormatByType(contactType) // 🔹 自動匹配 format

// 	// 取得對應的模板
// 	template := s.matchTemplate(ruleState, notifyFormat)
// 	if err != nil {
// 		s.logger.Error("獲取對應的模板失敗", zap.Error(err))
// 	}

// 	// 準備通知內容
// 	data := map[string]interface{}{
// 		"timestamp":     time.Unix(now, 0).Format("2006-01-02 15:04:05"),
// 		"resource_name": triggereds[0].ResourceName,
// 		"rule_name":     triggereds[0].RuleName,
// 		"severity":      triggereds[0].Severity,
// 		"current_value": triggereds[0].TriggeredValue,
// 		"threshold":     triggereds[0].Threshold,
// 		"unit":          triggereds[0].Unit,
// 		//"labels":        triggereds[0].Labels,
// 	}

// 	// 渲染通知內容
// 	message, err := s.templateService.RenderMessage(template, data)
// 	if err != nil {
// 		s.logger.Error("渲染通知內容失敗", zap.Error(err))
// 		message = "告警通知發生錯誤，請聯繫管理員"
// 	}

// 	// 解析聯絡人重試延遲
// 	contactRetryDelay, err := s.parseDuration(contact.RetryDelay)
// 	if err != nil {
// 		s.logger.Error("轉換 RetryDelay 失敗", zap.Error(err))
// 	}

// 	// 創建通知日誌
// 	notify := models.NotifyLog{
// 		ID:                id.Must(id.NewV7()).String(),
// 		Timestamp:         now,
// 		Title:             template.Title,
// 		Message:           message,
// 		ContactID:         contactID,
// 		ContactName:       contact.Name,
// 		ContactType:       contactType,
// 		ContactMaxRetry:   contact.MaxRetry,
// 		ContactRetryDelay: contactRetryDelay,
// 		RetryCounter:      0,
// 		TriggeredLogs:     make([]*models.TriggeredLog, len(triggereds)),
// 	}

// 	// 複製 TriggeredLog 指針
// 	for i := range triggereds {
// 		notify.TriggeredLogs[i] = &triggereds[i]
// 	}

// 	return notify
// }

//* ======================== 7.notify_sent.go 通知發送 ========================

// func (s *Service) sendNotifyLog(notify *models.NotifyLog) error {

// 	//* 依據 RetryCounter 決定是否繼續重試
// 	if notify.RetryCounter >= notify.ContactMaxRetry {
// 		notify.NotifyState = "failed"
// 		notify.Error = fmt.Sprintf("超過最大重試次數 %d", notify.ContactMaxRetry)
// 		if err := s.mysql.UpdateNotifyLog(*notify); err != nil {
// 			s.logger.Error("更新通知狀態失敗",
// 				zap.String("id", notify.ID),
// 				zap.Error(err))
// 		}
// 		return fmt.Errorf("超過最大重試次數 %d", notify.ContactMaxRetry)
// 	}

// 	//* 退避策略：如果上次重試失敗，則等待 ContactRetryDelay 時間後重試
// 	if notify.LastRetryTime > 0 {
// 		elapsed := time.Now().Unix() - notify.LastRetryTime
// 		if elapsed < int64(notify.ContactRetryDelay) {
// 			s.logger.Debug("等待重試時間",
// 				zap.String("id", notify.ID),
// 				zap.Int64("remaining_seconds", int64(notify.ContactRetryDelay)-elapsed))
// 			return nil
// 		} else {
// 			s.logger.Debug("重試時間已過，開始重試",
// 				zap.String("id", notify.ID),
// 				zap.Int64("elapsed_seconds", elapsed))
// 		}
// 	}

// 	//* 發送成功/失敗後，更新重試次數和狀態
// 	notify.RetryCounter++
// 	notify.NotifyState = "sending"
// 	notify.LastRetryTime = time.Now().Unix()

// 	if err := s.mysql.UpdateNotifyLog(*notify); err != nil {
// 		return fmt.Errorf("更新通知日誌失敗: %w", err)
// 	}

// 	// 發送通知
// 	notify.ContactConfig["title"] = notify.Title
// 	notify.ContactConfig["message"] = notify.Message
// 	err := s.notifyService.Send(common.NotifySetting{
// 		Type:   notify.ContactType,
// 		Config: notify.ContactConfig,
// 	})

// 	now := time.Now().Unix()
// 	if err != nil {
// 		notify.NotifyState = "failed"
// 		notify.Error = err.Error()
// 		notify.LastFailedTime = now
// 	} else {
// 		notify.NotifyState = "sent"
// 		notify.SentAt = &now
// 		notify.Error = ""
// 	}

// 	// 更新發送結果
// 	if err := s.mysql.UpdateNotifyLog(*notify); err != nil {
// 		s.logger.Error("更新通知狀態失敗",
// 			zap.String("id", notify.ID),
// 			zap.Error(err))
// 	}

// 	return err
// }

// var formatMap = map[string]string{
// 	"email":   "html",
// 	"slack":   "markdown",
// 	"discord": "markdown",
// 	"teams":   "markdown",
// 	"line":    "markdown",
// 	"webhook": "json",
// }

// // GetFormatByType 根據通知類型獲取對應的通知格式
// func GetFormatByType(contactType string) string {
// 	if format, exists := formatMap[contactType]; exists {
// 		return format
// 	}
// 	return "text" // 預設為 text
// }

// func (s *Service) matchTemplate(ruleState string, formatType string) models.Template {
// 	for _, template := range s.global.Templates {
// 		if template.RuleState == ruleState && template.FormatType == formatType {
// 			return template
// 		}
// 	}
// 	return models.Template{}
// }

// // 發送通知
// func (s *Service) Test(typ string) error {
// 	info := common.NotifySetting{
// 		Type: typ,
// 		Config: map[string]string{
// 			"title":   "Test " + typ,
// 			"message": "This is a test message from alert system.",
// 		},
// 	}

// 	return s.notifyService.Send(info)

// }

// // * 解析 metric_pattern，產生 partition_key 和 partition_value
// func (s *Service) ExtractPartitionsFromMetric(metricKey, metricPattern string) (map[string]string, error) {
// 	// 🔹 1️⃣ 確保 `metricPattern` 格式正確
// 	if !strings.Contains(metricPattern, "{") {
// 		return nil, fmt.Errorf("Invalid metricPattern format: %s", metricPattern)
// 	}

// 	// 🔹 2️⃣ 解析 `metricPattern`
// 	patternParts := strings.Split(metricPattern, ":")
// 	metricParts := strings.Split(metricKey, ":")

// 	if len(patternParts) != len(metricParts) {
// 		return nil, fmt.Errorf("Metric key does not match pattern: %s", metricKey)
// 	}

// 	// 🔹 3️⃣ 生成 `partition_key` 和 `partition_value`
// 	partitions := make(map[string]string)
// 	for i, part := range patternParts {
// 		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
// 			key := part[1 : len(part)-1] // 去掉 `{}` 提取變數名稱
// 			partitions[key] = metricParts[i]
// 		}
// 	}

// 	return partitions, nil
// }
