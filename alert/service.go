package alert

import (
	"encoding/json"
	"fmt"

	"shared-lib/models"
	"shared-lib/models/common"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shared-lib/interfaces"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	SentPath       = "sent"       // 待轉換的觸發日誌目錄
	NotifyPath     = "notify"     // 通知目錄
	SilencePath    = "silence"    // 靜音目錄
	UnresolvedPath = "unresolved" // 未解決目錄
	ResolvedPath   = "resolved"   // 已解決目錄
)

// Service 告警服務
type Service struct {
	config       models.AlertConfig // 只使用自己的配置
	rules        map[string]map[string][]models.CheckRule
	logger       interfaces.Logger
	logMgr       interfaces.LogManager
	db           interfaces.Database
	stateManager *AlertStateManager
}

// NewService 創建告警服務
func NewService(config models.AlertConfig, db interfaces.Database, logSvc interfaces.Logger, logMgr interfaces.LogManager) *Service {
	logger := logSvc.With(zap.String("module", "alert"))
	alertService := &Service{
		config: config,

		logger:       logger,
		logMgr:       logMgr,
		db:           db,
		stateManager: &AlertStateManager{db: db},
	}

	// TODO: 從資料庫獲取規則
	allCheckRules := make(map[string]map[string][]models.CheckRule)
	checkRules := make(map[string][]models.CheckRule)
	alertRules, err := db.GetAlertRules()
	if err != nil {
		logger.Error("獲取告警規則失敗", zap.Error(err))
		return nil
	}

	for realm, rules := range alertRules {
		allCheckRules[realm] = make(map[string][]models.CheckRule)
		for _, rule := range rules {
			// 2. alert_rules 扁平化而且轉換為 map[resource_name][]check_rules
			var muteStart, muteEnd int64
			if len(rule.MuteRules) > 0 {
				muteStart = int64(rule.MuteRules[0].StartTime)
				muteEnd = int64(rule.MuteRules[0].EndTime)
				for _, mute := range rule.MuteRules[1:] {
					if int64(mute.StartTime) < muteStart {
						muteStart = int64(mute.StartTime)
					}
					if int64(mute.EndTime) > muteEnd {
						muteEnd = int64(mute.EndTime)
					}
				}
			}
			for _, detail := range rule.AlertRuleDetails {
				resourceGroupName, err := alertService.db.GetResourceGroupName(rule.ResourceGroupID)
				if err != nil {
					logger.Error("獲取資源群組名稱失敗", zap.Error(err))
					continue
				}
				labels, err := alertService.db.GetCustomLabels(rule.ID)
				if err != nil {
					logger.Error("獲取自定義標籤失敗", zap.Error(err))
					continue
				}
				contacts, err := alertService.db.GetAlertContacts(rule.ID)
				if err != nil {
					logger.Error("獲取通知對象失敗", zap.Error(err))
					continue
				}

				check_rule := models.CheckRule{
					// 動態檢查
					UUID:         "", // 唯一識別碼
					Timestamp:    0,  // 異常檢測時間
					CurrentValue: 0,  // 異常數值
					Severity:     "", // info / warn / crit
					Status:       "", // alerting / normal / silenced

					// 靜態資訊
					RealmName:         realm,                // 告警規則所在的 realm
					ResourceGroupID:   rule.ResourceGroupID, // 資源群組 ID
					ResourceGroupName: resourceGroupName,    // 資源群組
					ResourceName:      detail.ResourceName,  // 監控的主機/設備
					PartitionName:     detail.PartitionName, // 分區名稱 (可選)
					MetricName:        rule.MetricRule.MetricName,
					CheckType:         rule.MetricRule.CheckType,
					Operator:          rule.MetricRule.Operator,
					InfoThreshold:     rule.InfoThreshold,
					WarnThreshold:     rule.WarnThreshold,
					CritThreshold:     rule.CritThreshold,
					Unit:              rule.MetricRule.Unit,
					Duration:          *rule.Duration,      // 異常持續時間
					RuleID:            rule.ID,             // 關聯的告警規則 ID
					RuleName:          rule.Name,           // 規則名稱
					SilenceStart:      detail.SilenceStart, // 靜音開始時間
					SilenceEnd:        detail.SilenceEnd,   // 靜音結束時間
					MuteStart:         &muteStart,          // 抑制開始時間(最早)
					MuteEnd:           &muteEnd,            // 抑制結束時間(最晚)
					Labels:            labels,              // 其他標籤
					Contacts:          contacts,            // 通知對象
				}

				checkRules[detail.ResourceName] = append(checkRules[detail.ResourceName], check_rule)
				allCheckRules[realm][detail.ResourceName] = append(allCheckRules[realm][detail.ResourceName], check_rule)
			}
		}
	}
	// json, _ := json.Marshal(allCheckRules)
	// fmt.Printf("allCheckRules: \n%v\n", string(json))
	alertService.rules = allCheckRules
	return alertService
}

// Init 初始化服務
func (s *Service) Init() error {
	s.db.LoadAlertMigrate(s.config.MigratePath)
	// 初始化目錄
	workDir := filepath.Join(s.config.WorkPath, ResolvedPath)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		s.logger.Error("創建通知目錄失敗",
			zap.String("path", workDir),
			zap.Error(err))
		return err
	}

	// 註冊輪轉任務
	if s.config.Rotate.Enabled {
		task := common.RotateTask{
			JobID:      "notify_rotate_" + workDir,
			SourcePath: workDir,
			DestPath:   workDir,
			RotateSetting: common.RotateSetting{
				Schedule:            "0 0 1 * * *",
				MaxAge:              time.Duration(s.config.Rotate.MaxAge),
				MaxSizeMB:           s.config.Rotate.MaxSizeMB,
				CompressEnabled:     true,
				CompressMatchRegex:  "*${YYYYMMDD}*.log",
				CompressOffsetHours: 2,
				CompressSaveRegex:   "${YYYYMMDD}.tar.gz",
				MinDiskFreeMB:       300,
			},
		}

		if err := s.logMgr.RegisterRotateTask(task); err != nil {
			return fmt.Errorf("註冊輪轉任務失敗: %w", err)
		}
		s.logger.Info("已註冊通知日誌輪轉任務",
			zap.String("source", task.SourcePath),
			zap.String("dest", task.DestPath))
	}

	s.logger.Info("通知服務初始化完成")
	return nil
}

// ProcessFile 處理檔案
func (s *Service) ProcessFile(file models.FileInfo, metrics map[string][]map[string]interface{}) error {

	// json, _ := json.Marshal(s.rules)
	// fmt.Printf("rules: %v\n", string(json))
	fmt.Printf("file.Realm: %v\n", file.Realm)
	fmt.Printf("file.Host: %v\n", file.Host)
	// 檢查規則是否存在
	rules, ok := s.rules[file.Realm][file.Host]
	if !ok || len(rules) == 0 {
		s.logger.Debug("找不到對應的規則", zap.String("host", file.Host))
		return nil
	}

	var exceeded bool
	// 檢查每個規則
	for _, rule := range rules {
		rule.UUID = uuid.New().String()
		rule.Timestamp = time.Now().Unix()
		var key string
		if rule.MetricName == "user_usage" || rule.MetricName == "system_usage" {
			key = fmt.Sprintf("%s:total", rule.MetricName)
		} else if rule.MetricName == "mem_usage" {
			key = rule.MetricName
		} else {
			key = fmt.Sprintf("%v:%v", rule.MetricName, rule.PartitionName)
		}

		metricData, ok := metrics[key]
		if !ok {
			s.logger.Debug("找不到對應的指標數據", zap.String("key", key))
			continue
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
			json, _ := json.Marshal(rule)
			fmt.Printf("TriggerLog rule: \n%v\n", string(json))
			trigger := s.convertCheckRuleToTriggerLog(&rule)

			// 根據通知狀態決定路徑
			var logPath string
			switch rule.NotifyStatus {
			case "muting":
				logPath = filepath.Join(s.config.WorkPath, UnresolvedPath, s.generateLogName(&trigger))
			case "pending":
				logPath = filepath.Join(s.config.WorkPath, SilencePath, s.generateLogName(&trigger))
			case "alerting":
				logPath = filepath.Join(s.config.WorkPath, SentPath, s.generateLogName(&trigger))
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
	dateStr := time.Unix(timestamp, 0).Format("20060102150405")
	return fmt.Sprintf("%s_%s_%s_%s_%s.log",
		dateStr,
		resourceName,
		ruleID,
		uuid,
		status)
}

// ProcessTriggers 批次處理觸發日誌
func (s *Service) ProcessTriggers() error {

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
	pattern := filepath.Join(s.config.WorkPath, SentPath, "*.log")
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
		dateStr := time.Unix(notification.Timestamp, 0).Format("20060102150405")
		filename := fmt.Sprintf("%s_%s_%s_%s_%s.log",
			dateStr,
			notification.ContactName,
			notification.ChannelType,
			notification.UUID,
			notification.Status)
		path := filepath.Join(s.config.WorkPath, NotifyPath, filename)

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
		oldPath := filepath.Join(s.config.WorkPath, SentPath, s.generateLogName(&trigger))
		newPath := filepath.Join(s.config.WorkPath, UnresolvedPath, s.generateLogName(&trigger))

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

	// 檢查 CurrentThreshold 是否為 nil
	var threshold float64
	if checkRule.CurrentThreshold != nil {
		threshold = *checkRule.CurrentThreshold
	}

	trigger := models.TriggerLog{
		UUID:              checkRule.UUID,
		Timestamp:         now,
		FirstTriggerTime:  now,
		RuleID:            checkRule.RuleID,
		RuleName:          checkRule.RuleName,
		ResourceGroupName: checkRule.ResourceGroupName,
		ResourceName:      checkRule.ResourceName,
		PartitionName:     checkRule.PartitionName,
		MetricName:        checkRule.MetricName,
		Value:             checkRule.CurrentValue,
		Threshold:         threshold, // 使用安全的值
		Unit:              checkRule.Unit,
		Severity:          checkRule.Severity,
		Duration:          checkRule.Duration,
		Status:            checkRule.Status,
		SilenceStart:      checkRule.SilenceStart,
		SilenceEnd:        checkRule.SilenceEnd,
		Labels:            checkRule.Labels,
		Contacts:          checkRule.Contacts,
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
			logPath = filepath.Join(s.config.WorkPath, UnresolvedPath, s.generateLogName(&trigger))
		case "pending":
			logPath = filepath.Join(s.config.WorkPath, SilencePath, s.generateLogName(&trigger))
		case "alerting":
			logPath = filepath.Join(s.config.WorkPath, SentPath, s.generateLogName(&trigger))
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
