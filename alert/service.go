package alert

import (
	"encoding/json"
	"fmt"

	"shared-lib/models"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shared-lib/interfaces"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service 告警服務
type Service struct {
	config       models.AlertConfig // 只使用自己的配置
	rules        map[string][]models.CheckRule
	logger       interfaces.Logger
	db           interfaces.Database
	parser       interfaces.Parser
	stateManager *AlertStateManager
}

// NewService 創建告警服務
func NewService(config models.AlertConfig, logSvc interfaces.Logger, db interfaces.Database) *Service {
	logger := logSvc.With(zap.String("module", "alert"))
	rules := make(map[string][]models.CheckRule)
	//TODO: 從資料庫獲取規則

	return &Service{
		config:       config,
		rules:        rules,
		logger:       logger,
		db:           db,
		stateManager: &AlertStateManager{db: db},
	}
}

// InitAlertDirs 初始化告警服務所需目錄
func (s *Service) InitAlertDirs() error {

	dirs := []string{
		s.config.WorkPath.Sent,
		s.config.WorkPath.Notify,
		s.config.WorkPath.Silence,
		s.config.WorkPath.Unresolved,
		s.config.WorkPath.Resolved,
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			s.logger.Error("創建告警目錄失敗",
				zap.String("path", dir),
				zap.Error(err))
			return err
		}
	}

	return nil
}

// Init 初始化服務
func (s *Service) Init() error {
	if err := s.InitAlertDirs(); err != nil {
		return fmt.Errorf("初始化目錄失敗: %w", err)
	}

	s.logger.Info("告警服務初始化完成")
	return nil
}

// ProcessFile 處理檔案
func (s *Service) ProcessFile(file models.FileInfo, metrics map[string][]map[string]interface{}) error {
	// 確保目錄存在
	if err := s.InitAlertDirs(); err != nil {
		return err
	}

	// 檢查規則是否存在
	rules, ok := s.rules[file.Host]
	if !ok || len(rules) == 0 {
		return nil
	}

	var exceeded bool
	// 檢查每個規則
	for _, rule := range rules {
		key := fmt.Sprintf("%s:%s", rule.Metric, rule.ResourceName)
		metricData, ok := metrics[key]
		if !ok {
			continue // 找不到對應的指標數據，跳過此規則
		}
		exceeded = s.Check(rule, file, metricData)

		if exceeded {
			// 5. 檢查靜音期
			now := time.Now().Unix()
			rule.Status = "alerting"       // 預設狀態
			rule.NotifyStatus = "alerting" // 預設通知狀態

			// 1. 檢查是否處於靜音時段
			if rule.SilenceStart != nil && rule.SilenceEnd != nil {
				if now >= *rule.SilenceStart && now <= *rule.SilenceEnd {
					rule.Status = "silenced"
					rule.NotifyStatus = "pending"
				}
			}

			// 2. 檢查是否處於抑制時段
			if rule.MuteStart != nil && rule.MuteEnd != nil {
				if now >= *rule.MuteStart && now <= *rule.MuteEnd {
					rule.Status = "muted"
					rule.NotifyStatus = "muting"
				}
			}

			// 轉換為 TriggerLog
			trigger := s.convertCheckRuleToTriggerLog(&rule)

			// 根據通知狀態決定路徑
			var logPath string
			switch rule.NotifyStatus {
			case "muting":
				logPath = filepath.Join(s.config.WorkPath.Unresolved, s.generateLogName(&trigger))
			case "pending":
				logPath = filepath.Join(s.config.WorkPath.Silence, s.generateLogName(&trigger))
			case "alerting":
				logPath = filepath.Join(s.config.WorkPath.Sent, s.generateLogName(&trigger))
			}

			if err := s.writeTriggeredLog(logPath, &trigger); err != nil {
				return fmt.Errorf("寫入觸發日誌失敗: %v", err)
			}

			// 寫入資料庫
			if err := s.db.WriteTriggeredLog(trigger); err != nil {
				return fmt.Errorf("寫入資料庫失敗: %v", err)
			}
		}
	}
	return nil
}

// ConvertToCheckRule 將 AlertRule 轉換為扁平的 CheckRule 格式
func (s *Service) ConvertToCheckRule(rule models.AlertRule) (*models.CheckRule, error) {
	// 1. 獲取 MetricRule 資訊
	var metricRule models.MetricRule
	if rule.MetricRule.ID != 0 {
		metricRule = rule.MetricRule
	} else {
		mr, err := s.db.GetMetricRule(rule.MetricRuleID)
		if err != nil {
			return nil, err
		}
		metricRule = mr
	}

	// 2. 獲取 RuleDetail 資訊
	var resourceName string
	var partitionName *string
	if len(rule.AlertRuleDetails) > 0 {
		detail := rule.AlertRuleDetails[0]
		resourceName = detail.ResourceName
		partitionName = detail.PartitionName
	} else {
		details, err := s.db.GetAlertRuleDetails(rule.ID)
		if err != nil {
			return nil, err
		}
		if len(details) > 0 {
			resourceName = details[0].ResourceName
			partitionName = details[0].PartitionName
		}
	}

	// 3. 構建 CheckRule
	checkRule := &models.CheckRule{
		UUID:          fmt.Sprintf("%d", rule.ID),
		ResourceGroup: fmt.Sprintf("%d", rule.ResourceGroupID),
		ResourceName:  resourceName,
		PartitionName: partitionName,
		Metric:        metricRule.MetricName,
		Unit:          metricRule.Unit,
		Duration:      *rule.Duration,
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		Status:        "alerting",
		NotifyStatus:  "pending",
		Labels:        make(models.JSONMap),
	}

	// 4. 設置閾值和嚴重程度
	if rule.CritThreshold != nil {
		checkRule.Threshold = *rule.CritThreshold
		checkRule.Severity = "crit"
	} else if rule.WarnThreshold != nil {
		checkRule.Threshold = *rule.WarnThreshold
		checkRule.Severity = "warn"
	} else if rule.InfoThreshold != nil {
		checkRule.Threshold = *rule.InfoThreshold
		checkRule.Severity = "info"
	}

	// 5. 添加標籤資訊
	labels, err := s.db.GetCustomLabels(rule.ID)
	if err != nil {
		return nil, err
	}
	for k, v := range labels {
		checkRule.Labels[k] = v
	}

	// 6. 添加靜音時間
	if len(rule.MuteRules) > 0 {
		mute := rule.MuteRules[0]
		startTime := int64(mute.StartTime)
		endTime := int64(mute.EndTime)
		checkRule.SilenceStart = &startTime
		checkRule.SilenceEnd = &endTime
		if checkRule.Status == "alerting" {
			checkRule.Status = "silenced"
		}
	}

	return checkRule, nil
}

// generateLogName 產生日誌檔名
func (s *Service) generateLogName(rule interface{}) string {
	var timestamp int64
	var resourceName, ruleID, uuid, status string

	switch r := rule.(type) {
	case *models.CheckRule:
		timestamp = r.Timestamp
		resourceName = r.ResourceName
		ruleID = fmt.Sprintf("%d", r.RuleID)
		uuid = r.UUID
		status = r.Status
	case *models.TriggerLog:
		timestamp = r.Timestamp
		resourceName = r.ResourceName
		ruleID = fmt.Sprintf("%d", r.RuleID)
		uuid = r.UUID
		status = r.Status
	}

	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	return fmt.Sprintf("%d_%s_%s_%s_%s.log",
		timestamp,
		resourceName,
		ruleID,
		uuid,
		status)
}

// ProcessTriggers 批次處理觸發日誌
func (s *Service) ProcessTriggers() error {
	// 確保目錄存在
	if err := s.InitAlertDirs(); err != nil {
		return fmt.Errorf("初始化目錄失敗: %v", err)
	}

	// 1. 讀取 sent_path 的觸發日誌
	triggers, err := s.readTriggerLogs()
	if err != nil {
		return fmt.Errorf("讀取觸發日誌失敗: %v", err)
	}

	if len(triggers) == 0 {
		s.logger.Debug("沒有需要處理的觸發日誌")
		return nil
	}

	// 2. 根據通知管道分組
	groups := s.groupTriggers(triggers)

	// 3. 產生通知日誌
	var notifications []models.NotificationLog
	for key, group := range groups {
		notification := s.createNotification(key, group)
		notifications = append(notifications, notification)
	}

	// 4. 寫入 notify_path
	if err := s.writeNotifications(notifications); err != nil {
		return fmt.Errorf("寫入通知日誌失敗: %v", err)
	}

	// 5. 觸發日誌歸檔到 unresolved_path
	if err := s.archiveTriggers(triggers); err != nil {
		return fmt.Errorf("歸檔觸發日誌失敗: %v", err)
	}

	s.logger.Info("處理觸發日誌完成",
		zap.Int("trigger_count", len(triggers)),
		zap.Int("notification_count", len(notifications)))

	return nil
}

// readTriggerLogs 讀取觸發日誌
func (s *Service) readTriggerLogs() ([]models.TriggerLog, error) {
	pattern := filepath.Join(s.config.WorkPath.Sent, "*_trigger.log")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var triggers []models.TriggerLog
	for _, file := range files {
		trigger, err := s.readTriggerLog(file)
		if err != nil {
			s.logger.Error("讀取觸發日誌失敗",
				zap.String("file", file),
				zap.Error(err))
			continue
		}
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

// groupTriggers 根據通知管道分組
func (s *Service) groupTriggers(triggers []models.TriggerLog) map[string][]models.TriggerLog {
	groups := make(map[string][]models.TriggerLog)

	for _, trigger := range triggers {
		for _, contact := range trigger.Contacts {
			key := fmt.Sprintf("%d_%s", contact.ID, contact.Type) // 使用 ID 和類型組合作為 key
			groups[key] = append(groups[key], trigger)
		}
	}
	return groups
}

// createNotification 創建通知日誌
func (s *Service) createNotification(key string, triggers []models.TriggerLog) models.NotificationLog {
	now := time.Now().Unix()

	// 解析聯絡人資訊
	parts := strings.Split(key, "_")
	contactID, _ := strconv.ParseInt(parts[0], 10, 64)
	contactType := parts[1]

	// 產生通知標題
	subject := fmt.Sprintf("[%s] %s - %s",
		triggers[0].Severity,
		triggers[0].ResourceName,
		triggers[0].RuleName)

	// 產生通知內容
	var body strings.Builder
	body.WriteString(fmt.Sprintf("觸發時間: %s\n", time.Unix(now, 0).Format("2006-01-02 15:04:05")))
	body.WriteString(fmt.Sprintf("資源名稱: %s\n", triggers[0].ResourceName))
	body.WriteString(fmt.Sprintf("告警規則: %s\n", triggers[0].RuleName))
	body.WriteString(fmt.Sprintf("異常等級: %s\n", triggers[0].Severity))
	body.WriteString(fmt.Sprintf("當前數值: %.2f %s\n", triggers[0].Value, triggers[0].Unit))
	body.WriteString(fmt.Sprintf("閾值設定: %.2f %s\n", triggers[0].Threshold, triggers[0].Unit))

	if len(triggers[0].Labels) > 0 {
		body.WriteString("\n標籤:\n")
		for k, v := range triggers[0].Labels {
			body.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
		}
	}

	notification := models.NotificationLog{
		UUID:          uuid.New().String(),
		Timestamp:     now,
		ContactID:     contactID,
		ContactName:   triggers[0].Contacts[0].Name,
		ChannelType:   contactType,
		Severity:      triggers[0].Severity,
		Subject:       subject,
		Body:          body.String(),
		Status:        "pending",
		NotifyRetry:   0,
		RetryDeadline: now + int64(s.config.NotifyPeriod),
		TriggerLogs:   make([]*models.TriggerLog, len(triggers)),
	}

	// 複製 TriggerLog 指針
	for i := range triggers {
		notification.TriggerLogs[i] = &triggers[i]
	}

	return notification
}

// writeNotifications 寫入通知日誌
func (s *Service) writeNotifications(notifications []models.NotificationLog) error {
	for _, notification := range notifications {
		filename := fmt.Sprintf("%d_%s_%s_%s_%s.log",
			notification.Timestamp,
			notification.ContactName,
			notification.ChannelType,
			notification.UUID,
			notification.Status)
		path := filepath.Join(s.config.WorkPath.Notify, filename)

		if err := s.writeNotification(path, notification); err != nil {
			s.logger.Error("寫入通知日誌失敗",
				zap.String("file", path),
				zap.Error(err))
			continue
		}
	}
	return nil
}

// archiveTriggers 歸檔觸發日誌
func (s *Service) archiveTriggers(triggers []models.TriggerLog) error {
	for _, trigger := range triggers {
		oldPath := filepath.Join(s.config.WorkPath.Sent, s.generateLogName(&trigger))
		newPath := filepath.Join(s.config.WorkPath.Unresolved, s.generateLogName(&trigger))

		if err := os.Rename(oldPath, newPath); err != nil {
			s.logger.Error("移動觸發日誌失敗",
				zap.String("from", oldPath),
				zap.String("to", newPath),
				zap.Error(err))
			continue
		}
	}
	return nil
}

// readTriggerLog 讀取單個觸發日誌
func (s *Service) readTriggerLog(path string) (models.TriggerLog, error) {
	var trigger models.TriggerLog

	// 1. 讀取檔案內容
	data, err := os.ReadFile(path)
	if err != nil {
		return trigger, fmt.Errorf("讀取檔案失敗: %v", err)
	}

	// 2. 解析 JSON
	if err := json.Unmarshal(data, &trigger); err != nil {
		return trigger, fmt.Errorf("解析觸發日誌失敗: %v", err)
	}

	s.logger.Debug("讀取觸發日誌成功",
		zap.String("path", path),
		zap.String("rule_id", fmt.Sprintf("%d", trigger.RuleID)))

	return trigger, nil
}

// writeNotification 寫入單個通知日誌
func (s *Service) writeNotification(path string, notification models.NotificationLog) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("序列化通知日誌失敗: %v", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("創建目錄失敗: %v", err)
	}

	if err := s.writeWithLock(path, data); err != nil {
		return fmt.Errorf("寫入檔案失敗: %v", err)
	}

	s.logger.Debug("寫入通知日誌成功",
		zap.String("path", path),
		zap.String("uuid", notification.UUID))

	return nil
}

// writeTriggeredLog 寫入觸發日誌到指定路徑
func (s *Service) writeTriggeredLog(path string, trigger *models.TriggerLog) error {
	data, err := json.Marshal(trigger)
	if err != nil {
		return fmt.Errorf("序列化觸發日誌失敗: %v", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("創建目錄失敗: %v", err)
	}

	if err := s.writeWithLock(path, data); err != nil {
		return fmt.Errorf("寫入檔案失敗: %v", err)
	}

	s.logger.Debug("寫入觸發日誌成功",
		zap.String("path", path),
		zap.String("rule_id", fmt.Sprintf("%d", trigger.RuleID)))

	return nil
}

// lockFile 鎖定檔案
func (s *Service) lockFile(path string) error {
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

// unlockFile 解除檔案鎖定
func (s *Service) unlockFile(path string) error {
	return os.Remove(path + ".lock")
}

// writeWithLock 帶鎖寫入檔案
func (s *Service) writeWithLock(path string, data []byte) error {
	if err := s.lockFile(path); err != nil {
		return err
	}
	defer s.unlockFile(path)

	return os.WriteFile(path, data, 0644)
}

// convertCheckRuleToTriggerLog 將 CheckRule 轉換為 TriggerLog
func (s *Service) convertCheckRuleToTriggerLog(checkRule *models.CheckRule) models.TriggerLog {
	now := time.Now().Unix()

	// 獲取聯絡人資訊
	contacts, err := s.db.GetAlertContacts(checkRule.RuleID)
	if err != nil {
		s.logger.Error("獲取聯絡人失敗",
			zap.String("rule_id", fmt.Sprintf("%d", checkRule.RuleID)),
			zap.Error(err))
	}

	trigger := models.TriggerLog{
		UUID:             checkRule.UUID,
		Timestamp:        now,
		FirstTriggerTime: now,
		RuleID:           checkRule.RuleID,
		RuleName:         checkRule.RuleName,
		ResourceGroup:    checkRule.ResourceGroup,
		ResourceName:     checkRule.ResourceName,
		PartitionName:    checkRule.PartitionName,
		Metric:           checkRule.Metric,
		Value:            checkRule.Value,
		Threshold:        checkRule.Threshold,
		Unit:             checkRule.Unit,
		Severity:         checkRule.Severity,
		Duration:         checkRule.Duration,
		Status:           checkRule.Status,
		SilenceStart:     checkRule.SilenceStart,
		SilenceEnd:       checkRule.SilenceEnd,
		Labels:           checkRule.Labels,
		Contacts:         contacts,
	}
	return trigger
}

// applySilenceAndMute 應用靜音和抑制規則
func (s *Service) applySilenceAndMute(rule *models.CheckRule) {
	now := time.Now().Unix()
	rule.Status = "alerting"       // 預設狀態
	rule.NotifyStatus = "alerting" // 預設通知狀態

	// 1. 檢查是否處於靜音時段
	if rule.SilenceStart != nil && rule.SilenceEnd != nil {
		if now >= *rule.SilenceStart && now <= *rule.SilenceEnd {
			rule.Status = "silenced"
			rule.NotifyStatus = "pending"
		}
	}

	// 2. 檢查是否處於抑制時段
	if rule.MuteStart != nil && rule.MuteEnd != nil {
		if now >= *rule.MuteStart && now <= *rule.MuteEnd {
			rule.Status = "muted"
			rule.NotifyStatus = "muting"
		}
	}
}

// Check 檢查告警規則
func (s *Service) Check(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	exceeded := s.CheckSingle(rule, file, metrics)

	if exceeded {
		s.applySilenceAndMute(&rule)

		// 轉換為 TriggerLog
		trigger := s.convertCheckRuleToTriggerLog(&rule)

		// 根據通知狀態決定路徑
		var logPath string
		switch rule.NotifyStatus {
		case "muting":
			logPath = filepath.Join(s.config.WorkPath.Unresolved, s.generateLogName(&trigger))
		case "pending":
			logPath = filepath.Join(s.config.WorkPath.Silence, s.generateLogName(&trigger))
		case "alerting":
			logPath = filepath.Join(s.config.WorkPath.Sent, s.generateLogName(&trigger))
		}

		if err := s.writeTriggeredLog(logPath, &trigger); err != nil {
			s.logger.Error("寫入觸發日誌失敗", zap.Error(err))
			return false
		}

		// 寫入資料庫
		if err := s.db.WriteTriggeredLog(trigger); err != nil {
			s.logger.Error("寫入資料庫失敗", zap.Error(err))
			return false
		}
	}

	return exceeded
}

// checkJoint 聯合檢查
func (s *Service) CheckJoint(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	// 同時滿足絕對值和振幅條件
	return s.CheckAbsolute(rule, file, metrics) && s.CheckAmplitude(rule, file, metrics)
}

// checkSingle 單一檢查
func (s *Service) CheckSingle(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	switch rule.CheckType {
	case "absolute":
		return s.CheckAbsolute(rule, file, metrics)
	case "amplitude":
		return s.CheckAmplitude(rule, file, metrics)
	default:
		s.logger.Error("未知的規則類型", zap.String("type", rule.CheckType))
		return false
	}
}
