package alert

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/detect-viz/shared-lib/labels"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/mutes"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/detect-viz/shared-lib/templates"

	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//* ======================== 1.init.go ========================

// Service 告警服務
type Service struct {
	ruleService      rules.Service
	muteService      mutes.Service
	contactService   contacts.Service
	notifyService    notifier.Service
	schedulerService *scheduler.Service
	templateService  templates.Service
	labelService     labels.Service
	config           models.AlertConfig
	global           models.GlobalConfig
	globalRules      map[string]map[string]map[string][]models.CheckRule
	logger           logger.Logger
	mysql            *mysql.Client
	keycloak         *keycloak.Client
}

// ✅ 直接返回已注入的 `RuleService`
func (s *Service) GetRuleService() rules.Service {
	return s.ruleService
}

// ✅ 直接返回已注入的 `MuteService`
func (s *Service) GetMuteService() mutes.Service {
	return s.muteService
}

// ✅ 直接返回已注入的 `NotifyService`
func (s *Service) GetNotifyService() notifier.Service {
	return s.notifyService
}

// ✅ 直接返回已注入的 `ContactService`
func (s *Service) GetContactService() contacts.Service {
	return s.contactService
}

// ✅ 直接返回已注入的 `LabelService`
func (s *Service) GetLabelService() labels.Service {
	return s.labelService
}

// NewService 創建告警服務
func NewService(
	config models.AlertConfig,
	global models.GlobalConfig,
	mysqlClient *mysql.Client,
	logSvc logger.Logger,
	rule rules.Service,
	mute mutes.Service,
	keycloak *keycloak.Client,
	notify notifier.Service,
	contact contacts.Service,
	scheduler *scheduler.Service,
	template templates.Service,
	label labels.Service,
) *Service {
	logger := logSvc.With(zap.String("module", "alert"))
	alertService := &Service{
		config:           config,
		global:           global,
		logger:           logger,
		mysql:            mysqlClient,
		ruleService:      rule,
		muteService:      mute,
		keycloak:         keycloak,
		notifyService:    notify,
		contactService:   contact,
		schedulerService: scheduler,
		templateService:  template,
		labelService:     label,
	}

	alertService.mysql.LoadAlertMigrate(alertService.config.MigratePath)

	// 註冊批次通知任務
	alertService.registerNotifyTask()

	allRules := alertService.getGlobalRules()
	if len(allRules) == 0 {
		alertService.logger.Warn("獲取告警規則為空")
		return nil
	}
	alertService.logger.Info("獲取告警規則成功")
	if alertService.logger.IsDebugMode() {
		globalRules, _ := json.MarshalIndent(allRules, "", "\t")
		os.Stdout.Write(globalRules)
	}
	alertService.globalRules = allRules

	return alertService
}

// 獲取所有告警規則
func (s *Service) getGlobalRules() map[string]map[string]map[string][]models.CheckRule {
	//* 0. 初始化 [realm][resource][metric:partition]{rule}
	allCheckRules := make(map[string]map[string]map[string][]models.CheckRule)

	//* 1. 獲取規則
	alertRules, err := s.mysql.GetRules()
	if err != nil {
		s.logger.Error("獲取告警規則失敗", zap.Error(err))
		return nil
	}

	//* 2. 轉換規則
	for realm, alert_rules := range alertRules {
		//* 第一層 - realm
		allCheckRules[realm] = make(map[string]map[string][]models.CheckRule)

		for _, alert_rule := range alert_rules {
			resourceGroupName, err := s.mysql.GetResourceGroupName(alert_rule.ResourceGroupID)
			if err != nil {
				s.logger.Error("獲取資源群組名稱失敗", zap.Error(err))
				continue
			}

			for _, detail := range alert_rule.AlertRuleDetails {
				//* 第二層 - resource
				if _, ok := allCheckRules[realm][detail.ResourceName]; !ok {
					allCheckRules[realm][detail.ResourceName] = make(map[string][]models.CheckRule)
				}

				var muteStart, muteEnd int64
				if s.muteService.IsRuleMuted(alert_rule.ResourceGroupID, time.Now()) {
					muteStart, muteEnd = s.muteService.GetMutePeriod(alert_rule.ResourceGroupID, time.Now())
					s.logger.Debug("抑制規則啟用",
						zap.Int64("mute_start", muteStart),
						zap.Int64("mute_end", muteEnd))
				}

				labels, err := s.mysql.GetRuleLabels(alert_rule.ID)
				if err != nil {
					s.logger.Error("獲取自定義標籤失敗", zap.Error(err))
					continue
				}

				// Convert labels to JSONMap
				labelsMap := make(models.JSONMap)
				for key, value := range labels {
					labelsMap[key] = value
				}
				contacts, err := s.contactService.GetContactsByRuleID(alert_rule.ID)
				if err != nil {
					s.logger.Error("獲取通知對象失敗", zap.Error(err))
					continue
				}

				alertState, err := s.mysql.GetAlertState(detail.ID)
				if err != nil {
					s.logger.Error("獲取告警狀態失敗", zap.Error(err))
					continue
				}
				//* 靜態資訊轉換
				check_rule := models.CheckRule{
					RealmName:         realm,                      // 告警規則所在的 realm
					ResourceGroupID:   alert_rule.ResourceGroupID, // 資源群組 ID
					ResourceGroupName: resourceGroupName,          // 資源群組
					ResourceName:      detail.ResourceName,        // 監控的主機/設備
					PartitionName:     detail.PartitionName,       // 分區名稱 (可選)
					MetricName:        alert_rule.MetricRule.MetricName,
					CheckType:         alert_rule.MetricRule.CheckType,
					Operator:          alert_rule.MetricRule.Operator,
					InfoThreshold:     alert_rule.InfoThreshold,
					WarnThreshold:     alert_rule.WarnThreshold,
					CritThreshold:     alert_rule.CritThreshold,
					DisplayUnit:       alert_rule.MetricRule.DisplayUnit,
					RawUnit:           alert_rule.MetricRule.RawUnit,
					Scale:             alert_rule.MetricRule.Scale,

					Duration:     *alert_rule.Duration,    // 異常持續時間
					RuleID:       alert_rule.ID,           // 關聯的告警規則 ID
					RuleName:     alert_rule.Name,         // 規則名稱
					SilenceStart: alertState.SilenceStart, // 靜音開始時間
					SilenceEnd:   alertState.SilenceEnd,   // 靜音結束時間
					MuteStart:    &muteStart,              // 抑制開始時間(最早)
					MuteEnd:      &muteEnd,                // 抑制結束時間(最晚)
					Labels:       labelsMap,               // 其他標籤
					Contacts:     contacts,                // 通知對象
				}

				//* 第三層 - metric
				var key string
				if detail.PartitionName != nil && *detail.PartitionName != "" && *detail.PartitionName != "total" {
					check_rule.Tags = make(map[string]string)
					check_rule.Tags[s.global.Tag.Base.Host] = detail.ResourceName
					key = alert_rule.MetricRule.MetricName + ":" + *detail.PartitionName
					tagkeys := strings.Split(alert_rule.MetricRule.Tags, ",")
					tagValues := strings.Split(*detail.PartitionName, ",")
					if len(tagkeys) != len(tagValues) {
						s.logger.Error("標籤數量不匹配", zap.String("metric", alert_rule.MetricRule.MetricName), zap.String("partition", *detail.PartitionName))
						continue
					}
					for i, tagKey := range tagkeys {
						check_rule.Tags[tagKey] = tagValues[i]
					}
				} else {
					key = alert_rule.MetricRule.MetricName
				}

				allCheckRules[realm][detail.ResourceName][key] = append(
					allCheckRules[realm][detail.ResourceName][key],
					check_rule,
				)
			}
		}
	}
	return allCheckRules
}

// 註冊批次通知任務
func (s *Service) registerNotifyTask() error {
	if s.config.NotifyPeriod > 0 {
		job := common.Task{
			Name:        "batch_notify",
			Spec:        fmt.Sprintf("@every %ds", s.config.NotifyPeriod),
			Type:        "cron",
			Enabled:     true,
			Timezone:    "Asia/Taipei",
			Description: "批次通知任務",
			RetryCount:  3,
			RetryDelay:  10 * time.Second,
			Duration:    10 * time.Second,
			ExecFunc: func() error {
				return s.ProcessNotifyLog()
			},
		}

		if err := s.schedulerService.RegisterTask(job); err != nil {
			s.logger.Error("註冊批次通知任務失敗",
				zap.Error(err),
				zap.Int("period", s.config.NotifyPeriod))
			return err
		}

		s.logger.Info("已註冊批次通知任務",
			zap.Int("period", s.config.NotifyPeriod))
	}

	s.logger.Info("通知服務初始化完成")

	return nil
}

// * ======================== 2.service.go 檢查主程式 ========================
func (s *Service) Process(file models.FileInfo, metrics map[string][]map[string]interface{}) error {
	matchRuleCounter := 0
	triggeredRuleCounter := 0
	//* 檢查 resource 是否有設定告警規則
	resourceRules, ok := s.globalRules[file.Realm][file.Host]
	if !ok {
		s.logger.Debug("找不到主機對應的告警規則",
			zap.String("Realm", file.Realm),
			zap.String("Host", file.Host))
		return nil
	}

	//* metric 跟 rule 匹配啟動 check 函式
	for metricKey, metricData := range metrics {
		metricRules, ok := resourceRules[metricKey]
		if !ok {
			s.logger.Debug("Metric 找不到對應的 rule ", zap.String("key", metricKey))
			continue
		}
		matchRuleCounter += len(metricRules)
		for _, metricRule := range metricRules {

			//* 使用 Check 方法處理告警邏輯
			rule := metricRule
			rule.Timestamp = time.Now().Unix()

			//* 確認 Contact State
			s.applySilence(&rule)
			s.applyMute(&rule)

			state, err := s.mysql.GetAlertState(rule.RuleDetailID)
			if err != nil {
				s.logger.Error("獲取告警狀態失敗", zap.Error(err))
				continue
			}
			exceeded, value, timestamp := s.CheckSingle(&rule, file, metricData, state)
			if !exceeded {
				continue
			}

			// 記錄最新數據
			state.LastTriggerValue = value
			state.LastTriggerTime = timestamp

			// 計算 stack_duration
			stackDuration := timestamp - state.LastTriggerTime

			// **異常觸發**
			if stackDuration >= int64(rule.Duration) {
				if state.RuleState != "alerting" {
					state.RuleState = "alerting"
					s.writeTriggerLog(rule, state) // 只寫入一次
				}
			} else {
				// **異常恢復**
				if state.RuleState == "alerting" {
					state.RuleState = "resolved"
					s.writeResolvedLog(rule, state) // 記錄恢復狀態
				} else {
					state.RuleState = "normal"
				}
				state.FirstTriggerTime = 0 // 清除 FirstTriggerTime
				stackDuration = 0          // 重置 stack_duration
			}

			// 更新 AlertState
			err = s.mysql.SaveAlertState(state)
			if err != nil {
				s.logger.Error("更新 AlertState 失敗", zap.Error(err))
				continue
			}
			triggeredRuleCounter++

		}

	}

	s.logger.Debug(fmt.Sprintf("檔案 %v 告警規則檢查總共 %v 條，觸發告警規則 %v 條", file.FileName, matchRuleCounter, triggeredRuleCounter))
	return nil
}

// 記錄 TriggerLog 並確保異常只寫入一次
func (s *Service) writeTriggerLog(rule models.CheckRule, state models.AlertState) error {

	// 檢查 CurrentThreshold 是否為 nil
	var threshold float64
	if rule.Threshold != nil {
		threshold = *rule.Threshold
	}

	// **查詢是否已有異常記錄**
	existingTrigger, err := s.mysql.GetActiveTriggerLog(rule.RuleID, rule.ResourceName, rule.MetricName)
	if err != nil {
		s.logger.Error("查詢 TriggerLog 失敗",
			zap.Int64("rule_id", rule.RuleID),
			zap.String("resource", rule.ResourceName),
			zap.String("metric", rule.MetricName),
			zap.Error(err))
		return err
	}

	// **如果異常持續發生，則更新 TriggerLog**
	if existingTrigger != nil {
		existingTrigger.Timestamp = time.Now().Unix()
		existingTrigger.TriggerValue = state.LastTriggerValue
		existingTrigger.Severity = rule.Severity
		return s.mysql.UpdateTriggerLog(*existingTrigger)
	}

	// **如果沒有異常記錄，則寫入新的 TriggerLog**
	trigger := models.TriggerLog{
		UUID:             uuid.New().String(),
		RuleID:           rule.RuleID,
		ResourceName:     rule.ResourceName,
		PartitionName:    rule.PartitionName,
		MetricName:       rule.MetricName,
		TriggerValue:     state.LastTriggerValue,
		Threshold:        threshold,
		Timestamp:        time.Now().Unix(),
		FirstTriggerTime: state.FirstTriggerTime,
		Severity:         rule.Severity,
		ContactState:     rule.ContactState,
		SilenceStart:     rule.SilenceStart,
		SilenceEnd:       rule.SilenceEnd,
		MuteStart:        rule.MuteStart,
		MuteEnd:          rule.MuteEnd,
		Labels:           rule.Labels,
		Contacts:         rule.Contacts,
	}

	return s.mysql.CreateTriggerLog(trigger)

}

// * ======================== 4.absolute.go 絕對值 ========================
func (s *Service) CheckAbsolute(rule *models.CheckRule, file models.FileInfo, metrics []map[string]interface{}, state models.AlertState) (bool, float64, int64) {
	if len(metrics) == 0 {
		return false, 0, 0
	}

	now := time.Now().Unix()
	var lastValue float64
	var allExceeded = true

	for _, data := range metrics {
		floatValue, err := s.parseMetricValue(data["value"])
		if err != nil {
			continue
		}

		exceeded, severity, threshold := s.checkThreshold(*rule, rule.Operator, floatValue)
		if !exceeded {
			allExceeded = false
		}

		rule.TriggerValue = floatValue
		rule.Threshold = threshold
		rule.Severity = severity
		lastValue = floatValue
	}

	return allExceeded, lastValue, now
}

// * ======================== 4.amplitude.go 振幅 ========================
func (s *Service) CheckAmplitude(rule *models.CheckRule, file models.FileInfo, metrics []map[string]interface{}, state models.AlertState) (bool, float64, int64) {
	if len(metrics) < 2 {
		return false, 0, 0 // 需要至少兩筆數據來比較
	}

	now := time.Now().Unix()
	var maxValue, minValue float64

	dataRange := metrics[max(0, len(metrics)-rule.Duration):]
	for _, data := range dataRange {
		floatValue, err := s.parseMetricValue(data["value"])
		if err != nil {
			continue
		}

		if floatValue > maxValue {
			maxValue = floatValue
		}
		if floatValue < minValue || minValue == 0 {
			minValue = floatValue
		}
	}

	amplitude := ((maxValue - minValue) / minValue) * 100
	exceeded, severity, threshold := s.checkThreshold(*rule, rule.Operator, amplitude)

	rule.TriggerValue = amplitude
	rule.Threshold = threshold
	rule.Severity = severity

	return exceeded, amplitude, now
}

// * ======================== 3.check.go 檢查告警邏輯 ========================
// 更新異常狀態
// parseMetricValue 解析數值

// 檢查閾值並返回嚴重程度
func (s *Service) checkThreshold(rule models.CheckRule, operator string, value float64) (bool, string, *float64) {

	var severity string

	switch operator {
	case ">", ">=":
		if rule.CritThreshold != nil && value > *rule.CritThreshold {
			return true, s.global.Code.Severity.Crit.Name, rule.CritThreshold
		}
		if rule.WarnThreshold != nil && value > *rule.WarnThreshold {
			return true, s.global.Code.Severity.Warn.Name, rule.WarnThreshold
		}
		if rule.InfoThreshold != nil && value > *rule.InfoThreshold {
			return true, s.global.Code.Severity.Info.Name, rule.InfoThreshold
		}
	case "<", "<=":
		if rule.CritThreshold != nil && value < *rule.CritThreshold {
			return true, s.global.Code.Severity.Crit.Name, rule.CritThreshold
		}
		if rule.WarnThreshold != nil && value < *rule.WarnThreshold {
			return true, s.global.Code.Severity.Warn.Name, rule.WarnThreshold
		}
		if rule.InfoThreshold != nil && value < *rule.InfoThreshold {
			return true, s.global.Code.Severity.Info.Name, rule.InfoThreshold
		}
	}
	return false, severity, nil
}

// checkJoint 聯合檢查
func (s *Service) CheckJoint(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}, state models.AlertState) bool {
	// TODO: 實作聯合檢查: 同時滿足絕對值和振幅條件
	return false
}

// checkSingle 單一檢查
func (s *Service) CheckSingle(rule *models.CheckRule, file models.FileInfo, metrics []map[string]interface{}, state models.AlertState) (bool, float64, int64) {
	switch rule.CheckType {
	case "absolute":
		return s.CheckAbsolute(rule, file, metrics, state)
	case "amplitude":
		return s.CheckAmplitude(rule, file, metrics, state)
	default:
		s.logger.Error("未知的規則類型", zap.String("type", rule.CheckType))
		return false, 0, 0
	}
}

//* ======================== 5.trigger_log.go 觸發日誌 ========================

// 根據通知管道分組 [contact.ID + contact.Type + rule.RuleState]
func (s *Service) groupTriggerLogs(triggers []models.TriggerLog, isResolved bool) map[string][]models.TriggerLog {
	groups := make(map[string][]models.TriggerLog)

	if len(triggers) == 0 {
		return groups
	}

	for _, trigger := range triggers {
		var ruleState string
		if trigger.ResolvedTime != nil {
			ruleState = "resolved"
		} else {
			ruleState = "alerting"
		}
		for _, contact := range trigger.Contacts {
			if isResolved && !contact.SentResolved {
				continue
			}
			key := fmt.Sprintf("%d_%s_%s", contact.ID, contact.Type, ruleState)
			groups[key] = append(groups[key], trigger)
		}
	}
	return groups
}

func (s *Service) writeResolvedLog(rule models.CheckRule, state models.AlertState) error {
	// 確保已經有 TriggerLog
	exists, err := s.mysql.CheckTriggerLogExists(rule.RuleDetailID, rule.ResourceName, rule.MetricName, state.FirstTriggerTime)
	if err != nil {
		return err
	}
	if !exists {
		s.logger.Warn("找不到對應的 TriggerLog，無法寫入 ResolvedLog",
			zap.Int64("rule_id", rule.RuleID),
			zap.String("resource", rule.ResourceName),
			zap.String("metric", rule.MetricName))
		return nil
	}

	// **更新 TriggerLog 狀態**
	err = s.mysql.UpdateTriggerLogResolved(rule.RuleID, rule.ResourceName, rule.MetricName, state.LastTriggerTime)
	if err != nil {
		s.logger.Error("更新 TriggerLog 為 resolved 失敗",
			zap.Int64("rule_id", rule.RuleID),
			zap.String("resource", rule.ResourceName),
			zap.String("metric", rule.MetricName),
			zap.Error(err))
		return err
	}

	return nil
}

//* ======================== 6.notify_log.go 通知日誌 ========================

// 批次處理觸發日誌
func (s *Service) ProcessNotifyLog() error {
	timestamp := time.Now().Unix()
	successAlertCounter, failAlertCounter := 0, 0
	successResolvedCounter, failResolvedCounter := 0, 0

	// **1️⃣ 查詢異常通知**
	triggerLogs, err := s.mysql.GetTriggerLogsForAlertNotify(timestamp)
	if err != nil {
		s.logger.Error("獲取待發送異常通知的 TriggerLog 失敗", zap.Error(err))
		return err
	}

	// **2️⃣ 查詢恢復通知**
	resolvedLogs, err := s.mysql.GetTriggerLogsForResolvedNotify(timestamp)
	if err != nil {
		s.logger.Error("獲取待發送恢復通知的 TriggerLog 失敗", zap.Error(err))
		return err
	}

	// **3️⃣ 根據通知管道分組**
	groupTriggerLogs := s.groupTriggerLogs(triggerLogs, false)  // 分組異常通知
	groupResolvedLogs := s.groupTriggerLogs(resolvedLogs, true) // 分組恢復通知

	// **4️⃣ 發送異常通知**
	for key, groupTriggerLog := range groupTriggerLogs {
		notifyLog := s.generateNotifyLog(key, groupTriggerLog)

		sendErr := s.sendNotifyLog(&notifyLog)
		if sendErr != nil {
			failAlertCounter++
			s.logger.Error("發送異常通知失敗", zap.Error(sendErr))
			notifyLog.NotifyState = s.global.Code.State.Notify.Failed.Name
			notifyLog.Error = sendErr.Error()
		} else {
			successAlertCounter++
			now := time.Now().Unix()
			notifyLog.NotifyState = s.global.Code.State.Notify.Solved.Name
			notifyLog.SentAt = &now
		}

		// **更新 TriggerLog 狀態**
		for _, trigger := range groupTriggerLog {
			err = s.mysql.UpdateTriggerLogNotifyState(trigger.UUID, notifyLog.NotifyState)
			if err != nil {
				s.logger.Error("更新 TriggerLog 通知狀態失敗", zap.Error(err))
			}
		}

		// **記錄 NotifyLog**
		err = s.mysql.CreateNotifyLog(notifyLog)
		if err != nil {
			s.logger.Error("寫入 NotifyLog 失敗", zap.Error(err))
		}
	}

	// **5️⃣ 發送恢復通知**
	for key, groupResolvedLog := range groupResolvedLogs {
		notifyLog := s.generateNotifyLog(key, groupResolvedLog)

		sendErr := s.sendNotifyLog(&notifyLog)
		if sendErr != nil {
			failResolvedCounter++
			s.logger.Error("發送恢復通知失敗", zap.Error(sendErr))
			notifyLog.NotifyState = s.global.Code.State.Notify.Failed.Name
			notifyLog.Error = sendErr.Error()
		} else {
			successResolvedCounter++
			now := time.Now().Unix()
			notifyLog.NotifyState = s.global.Code.State.Notify.Solved.Name
			notifyLog.SentAt = &now
		}

		// **更新 TriggerLog 狀態**
		for _, trigger := range groupResolvedLog {
			err = s.mysql.UpdateTriggerLogResolvedNotifyState(trigger.UUID, notifyLog.NotifyState)
			if err != nil {
				s.logger.Error("更新 TriggerLog 恢復通知狀態失敗", zap.Error(err))
			}
		}

		// **記錄 NotifyLog**
		err = s.mysql.CreateNotifyLog(notifyLog)
		if err != nil {
			s.logger.Error("寫入 NotifyLog 失敗", zap.Error(err))
		}
	}

	// **6️⃣ 結束 Log**
	s.logger.Info("批次處理發送通知日誌完成",
		zap.Int("success_alert_count", successAlertCounter),
		zap.Int("fail_alert_count", failAlertCounter),
		zap.Int("success_resolved_count", successResolvedCounter),
		zap.Int("fail_resolved_count", failResolvedCounter))

	return nil
}

// 創建通知日誌
func (s *Service) generateNotifyLog(key string, triggers []models.TriggerLog) models.NotifyLog {
	now := time.Now().Unix()

	// 解析聯絡人資訊
	parts := strings.Split(key, "_")
	contactID, _ := strconv.ParseInt(parts[0], 10, 64)
	contactType := parts[1]
	ruleState := parts[2]
	contact, err := s.contactService.Get(contactID)
	if err != nil {
		s.logger.Error("獲取聯絡人資訊失敗", zap.Error(err))
	}
	notifyFormat := GetFormatByType(contactType) // 🔹 自動匹配 format

	// 取得對應的模板
	template, err := s.mysql.GetTemplate(contact.RealmName, ruleState, notifyFormat)
	if err != nil {
		s.logger.Error("獲取對應的模板失敗", zap.Error(err))
	}

	// 準備通知內容
	data := map[string]interface{}{
		"timestamp":     time.Unix(now, 0).Format("2006-01-02 15:04:05"),
		"resource_name": triggers[0].ResourceName,
		"rule_name":     triggers[0].RuleName,
		"severity":      triggers[0].Severity,
		"current_value": triggers[0].TriggerValue,
		"threshold":     triggers[0].Threshold,
		"unit":          triggers[0].Unit,
		"labels":        triggers[0].Labels,
	}

	// 渲染通知內容
	message, err := s.templateService.RenderMessage(template, data)
	if err != nil {
		s.logger.Error("渲染通知內容失敗", zap.Error(err))
		message = "告警通知發生錯誤，請聯繫管理員"
	}

	notify := models.NotifyLog{
		UUID:        uuid.New().String(),
		Timestamp:   now,
		ContactID:   contactID,
		ContactName: contact.Name,
		ContactType: contactType,

		Title:        template.Title,
		Message:      message,
		RetryCounter: 0,
		TriggerLogs:  make([]*models.TriggerLog, len(triggers)),
	}

	// 複製 TriggerLog 指針
	for i := range triggers {
		notify.TriggerLogs[i] = &triggers[i]
	}

	return notify
}

//* ======================== 7.notify_sent.go 通知發送 ========================

func (s *Service) sendNotifyLog(notify *models.NotifyLog) error {
	// 檢查重試次數
	if notify.RetryCounter >= notify.ContactMaxRetry {
		notify.NotifyState = "failed"
		notify.Error = fmt.Sprintf("超過最大重試次數 %d", notify.ContactMaxRetry)
		if err := s.mysql.UpdateNotifyLog(*notify); err != nil {
			s.logger.Error("更新通知狀態失敗",
				zap.String("uuid", notify.UUID),
				zap.Error(err))
		}
		return fmt.Errorf("超過最大重試次數 %d", notify.ContactMaxRetry)
	}

	// 更新重試次數和狀態
	notify.RetryCounter++
	notify.NotifyState = "sending"
	notify.LastRetryTime = time.Now().Unix()

	if err := s.mysql.UpdateNotifyLog(*notify); err != nil {
		return fmt.Errorf("更新通知日誌失敗: %w", err)
	}

	// 發送通知
	notify.ContactConfig["title"] = notify.Title
	notify.ContactConfig["message"] = notify.Message
	err := s.notifyService.Send(common.NotifySetting{
		Type:   notify.ContactType,
		Config: notify.ContactConfig,
	})

	now := time.Now().Unix()
	if err != nil {
		notify.NotifyState = "failed"
		notify.Error = err.Error()
		notify.LastFailedTime = now
	} else {
		notify.NotifyState = "sent"
		notify.SentAt = &now
		notify.Error = ""
	}

	// 更新發送結果
	if err := s.mysql.UpdateNotifyLog(*notify); err != nil {
		s.logger.Error("更新通知狀態失敗",
			zap.String("uuid", notify.UUID),
			zap.Error(err))
	}

	return err
}

// GetFormatByType 根據通知類型獲取對應的通知格式
func GetFormatByType(contactType string) string {
	switch contactType {
	case "email":
		return "html"
	case "slack", "discord", "teams", "webex", "line":
		return "markdown"
	case "webhook":
		return "json"
	default:
		return "text"
	}
}
