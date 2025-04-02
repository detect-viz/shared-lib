package alert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/detect-viz/shared-lib/templates"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertService 子函數說明：
// CheckPayload - 初步驗證 AlertPayload
// AutoApplyTarget - 自動匹配監控對象
// AutoApplyRule - 自動匹配告警規則
// GetActiveRules - 查詢符合條件的規則
// CheckSingle - 執行異常檢測 (CheckAbsolute / CheckAmplitude)
// updateAlertState - 更新 rule_states
// processTriggerLog - 建立 TriggeredLog 記錄

type Service struct {
	ruleService      rules.Service
	contactService   contacts.Service
	notifyService    notifier.Service
	schedulerService scheduler.Service
	templateService  templates.Service
	config           models.AlertConfig
	global           models.GlobalConfig
	globalRules      map[string]map[string]map[string][]models.Rule
	logger           logger.Logger
	mysql            *mysql.Client
}

func (s *Service) GetRuleService() rules.Service {
	return s.ruleService
}

func (s *Service) GetNotifyService() notifier.Service {
	return s.notifyService
}

func (s *Service) GetContactService() contacts.Service {
	return s.contactService
}

func (s *Service) GetSchedulerService() scheduler.Service {
	return s.schedulerService
}

func (s *Service) GetTemplateService() templates.Service {
	return s.templateService
}

func (s *Service) GetMysqlClient() *mysql.Client {
	return s.mysql
}

// 創建告警服務
func NewService(
	config models.AlertConfig,
	global models.GlobalConfig,
	mysqlClient *mysql.Client,
	logSvc logger.Logger,
	rule rules.Service,
	notify notifier.Service,
	contact contacts.Service,
	scheduler scheduler.Service,
	template templates.Service,
) *Service {
	alertService := &Service{
		ruleService:      rule,
		contactService:   contact,
		notifyService:    notify,
		schedulerService: scheduler,
		templateService:  template,
		config:           config,
		global:           global,
		logger:           logSvc,
		mysql:            mysqlClient,
	}

	// 註冊通知任務
	if err := alertService.registerNotifyTask(); err != nil {
		logSvc.Error("註冊通知任務失敗", zap.Error(err))
	}

	// 載入告警遷移
	alertService.mysql.LoadAlertMigrate(config.MigratePath)

	// 載入所有規則
	alertService.globalRules = alertService.getGlobalRules()

	// 啟動定期重新載入任務
	alertService.scheduleRulesReload()

	// 設置規則變更回調
	rule.SetRuleChangeCallback(alertService.updateGlobalRules)

	// 設置聯絡人變更回調
	contact.SetContactChangeCallback(alertService.updateGlobalContacts)

	return alertService
}

// AutoApply 自動匹配監控對象和告警規則，並返回匹配的規則列表
func (s *Service) AutoApply(payload models.AlertPayload) ([]models.Rule, error) {
	s.logger.Debug("自動匹配監控對象&告警規則",
		zap.String("realm", payload.Metadata.RealmName),
		zap.String("resource", payload.Metadata.ResourceName),
		zap.String("datasource", payload.Metadata.DatasourceName))

	// 檢查是否啟用自動應用規則
	alertConfig := s.config
	if !alertConfig.AutoApplyRule {
		s.logger.Info("自動應用規則功能已禁用，跳過自動匹配")
		// 仍然返回現有的規則
		rules, err := s.mysql.GetActiveRules(payload.Metadata.RealmName, payload.Metadata.ResourceName)
		if err != nil {
			s.logger.Error("獲取活動規則失敗",
				zap.Error(err),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))
			return nil, err
		}
		return rules, nil
	}

	// 用於存儲匹配的規則
	var matchedRules []models.Rule
	var targetsCreated bool = false

	// 檢查每個 metric 的 partition
	for metricKey := range payload.Data {
		// 解析 metricKey 獲取 partition
		parts := strings.Split(metricKey, ":")
		var metricName, partitionName string

		if len(parts) < 2 {
			// 如果格式不符合預期，使用整個 metricKey 作為 metricName，partition 設為空
			metricName = metricKey
			partitionName = ""
			s.logger.Warn("metric key 格式不符合 metric:partition 格式，使用整個 key 作為 metric 名稱",
				zap.String("metric_key", metricKey),
				zap.String("metric_name", metricName))
		} else {
			metricName = parts[0]
			partitionName = strings.Join(parts[1:], ":")
		}

		// 檢查 target 是否存在
		exists, err := s.mysql.CheckTargetExists(
			payload.Metadata.RealmName,
			payload.Metadata.DatasourceName,
			payload.Metadata.ResourceName,
			partitionName,
		)
		if err != nil {
			s.logger.Error("檢查 target 失敗",
				zap.Error(err),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName),
				zap.String("partition", partitionName))
			continue
		}

		if exists {
			s.logger.Debug("target 已存在，獲取相關規則",
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName),
				zap.String("partition", partitionName))

			// 獲取與此 target 相關的規則
			targetRules, err := s.mysql.GetRulesByTarget(
				payload.Metadata.RealmName,
				payload.Metadata.ResourceName,
				partitionName)
			if err != nil {
				s.logger.Error("獲取 target 相關規則失敗",
					zap.Error(err),
					zap.String("realm", payload.Metadata.RealmName),
					zap.String("resource", payload.Metadata.ResourceName),
					zap.String("partition", partitionName))
			} else if len(targetRules) > 0 {
				// 將找到的規則添加到匹配規則列表中
				matchedRules = append(matchedRules, targetRules...)
			}
			continue
		}

		// 如果 target 不存在，則創建
		if errors.Is(err, gorm.ErrRecordNotFound) || !exists {
			targetsCreated = true
			category := s.getMetricCategory(payload.Metadata.DatasourceName, metricName)

			var collection_interval int
			if len(payload.Data[metricKey]) > 2 {
				collection_interval = int(payload.Data[metricKey][0].Timestamp - payload.Data[metricKey][1].Timestamp)
			}

			var reporting_interval int
			if len(payload.Data[metricKey]) > 2 {
				reporting_interval = int(payload.Data[metricKey][0].Timestamp - payload.Data[metricKey][len(payload.Data[metricKey])-1].Timestamp)
			}

			s.logger.Info("自動創建 target",
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName),
				zap.String("partition", partitionName),
				zap.String("metric", metricName),
				zap.String("datasource", payload.Metadata.DatasourceName))

			// 創建新的 target
			newTarget := models.Target{
				RealmName:          payload.Metadata.RealmName,
				ResourceName:       payload.Metadata.ResourceName,
				PartitionName:      partitionName,
				DatasourceName:     payload.Metadata.DatasourceName,
				CollectionInterval: collection_interval,
				ReportingInterval:  reporting_interval,
			}
			systemStr := "system"
			newTarget.Category = category
			newTarget.CreatedBy = &systemStr
			// 移除不存在的 Status 欄位
			newTarget.IsHidden = false
			// 保存到數據庫
			newTargetResult, err := s.mysql.CreateTarget(&newTarget)
			if err != nil {
				s.logger.Error("創建 target 失敗",
					zap.Error(err),
					zap.Any("target", newTarget))
				continue
			}

			s.logger.Info("成功創建 target",
				zap.String("id", string(newTargetResult.ID)),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName),
				zap.String("partition", partitionName))

			// 自動匹配告警規則
			var createdRules []models.Rule
			autoApplyRules := s.matchAutoApplyRule(payload.Metadata.RealmName, payload.Metadata.DatasourceName, metricName)
			if autoApplyRules != nil && len(*autoApplyRules) > 0 {
				s.logger.Info("找到匹配的自動應用規則",
					zap.Int("count", len(*autoApplyRules)),
					zap.String("realm", payload.Metadata.RealmName),
					zap.String("metric", metricName))

				for _, rule := range *autoApplyRules {
					var newRule models.Rule
					newRule.RealmName = payload.Metadata.RealmName
					newRule.MetricRuleUID = rule.MetricRuleUID
					newRule.TargetID = newTargetResult.ID
					newRule.CreateType = "system"
					newRule.CreatedBy = &systemStr
					newRule.Enabled = true
					newRule.AutoApply = false
					newRule.InfoThreshold = rule.InfoThreshold
					newRule.WarnThreshold = rule.WarnThreshold
					newRule.CritThreshold = rule.CritThreshold
					newRule.Duration = rule.Duration
					newRule.Times = rule.Times
					newRule.SilencePeriod = rule.SilencePeriod
					createdRules = append(createdRules, newRule)
				}
			} else {
				// 如果沒有找到匹配的自動應用規則，但 AutoApplyRule 為 true，則嘗試從 MetricRules 中找到匹配的規則
				s.logger.Info("未找到匹配的自動應用規則，嘗試從 MetricRules 中找到匹配的規則",
					zap.String("realm", payload.Metadata.RealmName),
					zap.String("metric", metricName))

				// 從 MetricRules 中找到匹配的規則
				for _, metricRule := range s.global.MetricRules {
					if metricRule.MetricRawName == metricName &&
						(len(metricRule.MatchDatasourceNames) == 0 || slices.Contains(metricRule.MatchDatasourceNames, payload.Metadata.DatasourceName)) {
						s.logger.Info("找到匹配的 MetricRule",
							zap.String("uid", metricRule.UID),
							zap.String("name", metricRule.Name),
							zap.String("metric", metricName))

						// 創建新的規則
						var newRule models.Rule
						newRule.RealmName = payload.Metadata.RealmName
						newRule.MetricRuleUID = metricRule.UID
						newRule.TargetID = newTargetResult.ID
						newRule.CreateType = "system"
						newRule.CreatedBy = &systemStr
						newRule.Enabled = true
						newRule.AutoApply = false

						// 設置閾值
						if metricRule.Thresholds.Info != nil {
							newRule.InfoThreshold = metricRule.Thresholds.Info
						}
						if metricRule.Thresholds.Warn != nil {
							newRule.WarnThreshold = metricRule.Thresholds.Warn
						}
						newRule.CritThreshold = metricRule.Thresholds.Crit

						// 設置其他參數
						// 解析持續時間
						durationStr := "5m" // 默認 5 分鐘
						if metricRule.Duration != "" {
							durationStr = metricRule.Duration
						}
						newRule.Duration = durationStr

						// 默認值
						times := 1
						newRule.Times = times

						silencePeriod := "1h" // 默認 1 小時
						newRule.SilencePeriod = silencePeriod

						createdRules = append(createdRules, newRule)
					}
				}
			}

			if len(createdRules) > 0 {
				s.logger.Info("創建規則",
					zap.Int("count", len(createdRules)),
					zap.String("realm", payload.Metadata.RealmName),
					zap.String("resource", payload.Metadata.ResourceName))

				err := s.mysql.CreateRules(createdRules)
				if err != nil {
					s.logger.Error("創建規則失敗",
						zap.Error(err),
						zap.Any("rules", createdRules),
						zap.String("realm", payload.Metadata.RealmName),
						zap.String("resource", payload.Metadata.ResourceName),
						zap.String("datasource", payload.Metadata.DatasourceName),
						zap.String("partition", partitionName))

					// 嘗試獲取更詳細的錯誤信息
					if strings.Contains(err.Error(), "Data truncated") {
						s.logger.Error("資料截斷錯誤，請檢查欄位值是否符合資料庫定義",
							zap.Error(err),
							zap.String("realm", payload.Metadata.RealmName))
					}
				} else {
					// 創建成功後，獲取新創建的規則
					newRules, err := s.mysql.GetRulesByTarget(
						payload.Metadata.RealmName,
						payload.Metadata.ResourceName,
						partitionName)
					if err != nil {
						s.logger.Error("獲取新創建的規則失敗",
							zap.Error(err),
							zap.String("realm", payload.Metadata.RealmName),
							zap.String("resource", payload.Metadata.ResourceName),
							zap.String("partition", partitionName))
					} else {
						// 自動套用通知管道
						if err := s.autoApplyContacts(newRules); err != nil {
							s.logger.Error("自動套用通知管道失敗",
								zap.Error(err),
								zap.String("realm", payload.Metadata.RealmName),
								zap.String("resource", payload.Metadata.ResourceName),
								zap.String("partition", partitionName))
						}

						matchedRules = append(matchedRules, newRules...)
					}
				}
			} else {
				s.logger.Warn("未找到匹配的規則模板，無法自動創建規則",
					zap.String("realm", payload.Metadata.RealmName),
					zap.String("resource", payload.Metadata.ResourceName),
					zap.String("metric", metricName))
			}
		}
	}

	// 如果沒有找到任何匹配的規則，則從資料庫獲取所有活動的規則
	if len(matchedRules) == 0 {
		if targetsCreated {
			s.logger.Info("已創建新的 targets，但尚未找到匹配的規則，請檢查 metric_rules 表是否有適用的規則模板",
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))
		} else {
			s.logger.Info("未找到匹配的 targets 或規則，嘗試獲取所有活動規則",
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))
		}

		rules, err := s.mysql.GetActiveRules(payload.Metadata.RealmName, payload.Metadata.ResourceName)
		if err != nil {
			s.logger.Error("獲取活動規則失敗",
				zap.Error(err),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))
			return nil, err
		}

		if len(rules) == 0 {
			s.logger.Warn("未找到任何活動規則，請確保已配置相關規則",
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))
		}

		matchedRules = rules
	}

	return matchedRules, nil
}

// 獲取所有告警規則，並組織成高效的查詢結構
func (s *Service) getGlobalRules() map[string]map[string]map[string][]models.Rule {
	//* 0. 初始化 [realm][resource][metric:partition]{rule}
	allRules := make(map[string]map[string]map[string][]models.Rule)

	//* 1. 獲取所有活動的規則
	// 從所有 realm 和 resource 獲取規則
	activeRules, err := s.mysql.GetAllActiveRules()
	if err != nil {
		s.logger.Error("獲取所有告警規則失敗", zap.Error(err))
		return nil
	}

	//* 2. 轉換規則為高效的查詢結構
	for _, rule := range activeRules {
		realmName := rule.RealmName
		resourceName := rule.Target.ResourceName
		partitionName := rule.Target.PartitionName

		// 初始化 map 結構
		if _, ok := allRules[realmName]; !ok {
			allRules[realmName] = make(map[string]map[string][]models.Rule)
		}
		if _, ok := allRules[realmName][resourceName]; !ok {
			allRules[realmName][resourceName] = make(map[string][]models.Rule)
		}

		// 使用 metricRuleUID:partitionName 作為 key
		key := rule.MetricRuleUID
		if partitionName != "" {
			key = key + ":" + partitionName
		}

		allRules[realmName][resourceName][key] = append(allRules[realmName][resourceName][key], rule)
	}

	s.logger.Info("已載入所有告警規則",
		zap.Int("realm_count", len(allRules)),
		zap.Int("rule_count", len(activeRules)))

	return allRules
}

// 添加一個互斥鎖來保護 globalRules
var globalRulesMutex sync.RWMutex

// 從 globalRules 中快速匹配規則
func (s *Service) matchRulesFromGlobalRules(payload models.AlertPayload) []models.Rule {
	globalRulesMutex.RLock()
	defer globalRulesMutex.RUnlock()

	var matchedRules []models.Rule
	realmName := payload.Metadata.RealmName
	resourceName := payload.Metadata.ResourceName

	// 1. 檢查是否有匹配的 realm 和 resource
	resourceRules, ok := s.globalRules[realmName]
	if !ok {
		return matchedRules
	}

	metricRules, ok := resourceRules[resourceName]
	if !ok {
		return matchedRules
	}

	// 2. 為每個 payload 中的 metric 找到匹配的規則
	metricKeyMap := make(map[string]bool)
	for metricKey := range payload.Data {
		metricKeyMap[metricKey] = true
	}

	// 3. 遍歷所有規則，檢查是否有匹配的 metric
	for ruleKey, rules := range metricRules {
		// 解析 ruleKey 獲取 metricRuleUID 和 partitionName
		parts := strings.Split(ruleKey, ":")
		metricRuleUID := parts[0]

		// 獲取對應的 MetricRule
		var metricRule models.MetricRule
		metricRule, exists := s.global.MetricRules[metricRuleUID]
		if !exists {
			s.logger.Error("找不到對應的 MetricRule",
				zap.String("metric_rule_uid", metricRuleUID))
			continue
		}

		if metricRule.UID == "" {
			continue
		}

		// 構建完整的 metricKey
		metricKey := metricRule.MetricRawName
		if len(parts) > 1 {
			partitionName := strings.Join(parts[1:], ":")
			metricKey = metricKey + ":" + partitionName
		}

		// 檢查 payload 中是否有對應的 metric 數據
		if metricKeyMap[metricKey] {
			matchedRules = append(matchedRules, rules...)
		}
	}

	return matchedRules
}

// 定期重新載入所有規則
func (s *Service) scheduleRulesReload() {
	// 每天凌晨 3 點重新載入所有規則
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
			time.Sleep(next.Sub(now))

			globalRulesMutex.Lock()
			s.globalRules = s.getGlobalRules()
			globalRulesMutex.Unlock()

			s.logger.Info("已重新載入所有規則")
		}
	}()
}

// 在 CRUD 操作後更新 globalRules
func (s *Service) updateGlobalRules(rule models.Rule, operation string) {
	globalRulesMutex.Lock()
	defer globalRulesMutex.Unlock()

	realmName := rule.RealmName
	resourceName := rule.Target.ResourceName
	partitionName := rule.Target.PartitionName

	// 初始化 map 結構（如果需要）
	if _, ok := s.globalRules[realmName]; !ok {
		s.globalRules[realmName] = make(map[string]map[string][]models.Rule)
	}
	if _, ok := s.globalRules[realmName][resourceName]; !ok {
		s.globalRules[realmName][resourceName] = make(map[string][]models.Rule)
	}

	// 使用 metricRuleUID:partitionName 作為 key
	key := rule.MetricRuleUID
	if partitionName != "" {
		key = key + ":" + partitionName
	}

	switch operation {
	case "create":
		s.globalRules[realmName][resourceName][key] = append(s.globalRules[realmName][resourceName][key], rule)
	case "update":
		// 找到並更新規則
		for i, r := range s.globalRules[realmName][resourceName][key] {
			if bytes.Equal(r.ID, rule.ID) {
				s.globalRules[realmName][resourceName][key][i] = rule
				break
			}
		}
	case "delete":
		// 找到並刪除規則
		for i, r := range s.globalRules[realmName][resourceName][key] {
			if bytes.Equal(r.ID, rule.ID) {
				s.globalRules[realmName][resourceName][key] = append(s.globalRules[realmName][resourceName][key][:i], s.globalRules[realmName][resourceName][key][i+1:]...)
				break
			}
		}
	}
}

// 在 Contact CRUD 操作後處理聯絡人變更
func (s *Service) updateGlobalContacts(contact models.Contact, operation string) {
	s.logger.Info("聯絡人變更",
		zap.String("contact_id", string(contact.ID)),
		zap.String("operation", operation))

	// 聯絡人變更後，重新載入所有規則
	// 這是因為聯絡人變更可能影響多個規則的通知設定
	globalRulesMutex.Lock()
	s.globalRules = s.getGlobalRules()
	globalRulesMutex.Unlock()

	s.logger.Info("已重新載入所有規則（由聯絡人變更觸發）")
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

	}

	s.logger.Info("通知服務初始化完成")

	return nil
}

// * ======================== 2.service.go 檢查主程式 (主入口) ========================
func (s *Service) ProcessAlert(payload models.AlertPayload) error {
	// 1. 檢查 payload 格式
	if err := s.CheckPayload(payload); err != nil {
		s.logger.Error("檢查 payload 失敗",
			zap.Error(err),
			zap.String("realm", payload.Metadata.RealmName),
			zap.String("resource", payload.Metadata.ResourceName),
			zap.String("datasource", payload.Metadata.DatasourceName))
		return fmt.Errorf("檢查 payload 失敗: %w", err)
	}

	// 2. 從 globalRules 中快速匹配規則
	rules := s.matchRulesFromGlobalRules(payload)

	// 如果沒有匹配到規則，則嘗試 AutoApply
	if len(rules) == 0 {
		s.logger.Debug("從 globalRules 中未匹配到規則，嘗試 AutoApply",
			zap.String("realm", payload.Metadata.RealmName),
			zap.String("resource", payload.Metadata.ResourceName))

		var err error
		rules, err = s.AutoApply(payload)
		if err != nil {
			s.logger.Error("自動匹配監控對象和規則失敗",
				zap.Error(err),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName),
				zap.String("datasource", payload.Metadata.DatasourceName))

			// 如果出錯，仍然嘗試從資料庫獲取規則
			rules, err = s.mysql.GetActiveRules(payload.Metadata.RealmName, payload.Metadata.ResourceName)
			if err != nil {
				s.logger.Error("獲取告警規則失敗",
					zap.Error(err),
					zap.String("realm", payload.Metadata.RealmName),
					zap.String("resource", payload.Metadata.ResourceName))
				return fmt.Errorf("獲取告警規則失敗 [realm:%s, resource:%s]: %w",
					payload.Metadata.RealmName, payload.Metadata.ResourceName, err)
			}
		}

		// 如果通過 AutoApply 創建了新規則，更新 globalRules
		if len(rules) > 0 {
			s.logger.Info("通過 AutoApply 創建了新規則，更新 globalRules",
				zap.Int("rule_count", len(rules)),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))

			// 重新載入 globalRules
			globalRulesMutex.Lock()
			s.globalRules = s.getGlobalRules()
			globalRulesMutex.Unlock()
		}
	}

	s.logger.Info("開始檢查告警規則",
		zap.String("realm", payload.Metadata.RealmName),
		zap.String("resource", payload.Metadata.ResourceName),
		zap.Int("rule_count", len(rules)))

	// 計數器
	matchRuleCounter := 0
	triggeredRuleCounter := 0

	// 3. 對每個規則進行檢查
	for _, rule := range rules {
		// 獲取對應的 MetricRule
		var metricRule models.MetricRule
		metricRule, exists := s.global.MetricRules[rule.MetricRuleUID]
		if !exists {
			s.logger.Error("找不到對應的 MetricRule",
				zap.String("metric_rule_uid", rule.MetricRuleUID),
				zap.String("rule_id", string(rule.ID)),
				zap.String("realm", payload.Metadata.RealmName),
				zap.String("resource", payload.Metadata.ResourceName))
			continue
		}

		matchRuleCounter++
		s.logger.Debug("找到對應的告警規則",
			zap.String("rule_id", string(rule.ID)),
			zap.String("metric_rule_uid", rule.MetricRuleUID))

		// 獲取 metric 數據
		metricKey := metricRule.MetricRawName
		if rule.Target.PartitionName != "" {
			metricKey = metricKey + ":" + rule.Target.PartitionName
		}

		metricValues, ok := payload.Data[metricKey]
		if !ok {
			s.logger.Debug("找不到對應的 metric 數據",
				zap.String("metric_key", metricKey),
				zap.String("rule_id", string(rule.ID)),
				zap.String("resource", payload.Metadata.ResourceName))
			continue
		}

		// 將 MetricValue 轉換為 MetricPoint
		var metricData []models.MetricValue
		for _, mv := range metricValues {
			metricData = append(metricData, models.MetricValue(mv))
		}

		// 4. 檢查告警邏輯
		currentTime := time.Now().Unix()
		exceeded, value, severity := s.CheckSingle(&rule, metricData, currentTime)

		// 5. 更新告警狀態
		// 無論是否觸發告警，都更新 LastCheckValue
		var newState *models.RuleState
		var err error
		if exceeded {
			// 如果觸發告警，使用完整的更新邏輯
			newState, err = s.updateAlertState(rule, value, severity, currentTime)
		} else {
			// 如果未觸發告警，僅更新 LastCheckValue
			s.logger.Debug("未觸發告警，僅更新最後檢查值",
				zap.String("rule_id", string(rule.ID)),
				zap.Float64("value", value),
				zap.String("metric", metricKey))
			newState, err = s.updateLastCheckValue(rule, value, currentTime)
		}

		if err != nil {
			s.logger.Error("更新告警狀態失敗",
				zap.String("rule_id", string(rule.ID)),
				zap.Error(err),
				zap.String("resource", payload.Metadata.ResourceName),
				zap.String("metric", metricKey),
				zap.Float64("value", value),
				zap.String("severity", severity))
			continue
		}

		// 只有在觸發告警時才處理觸發日誌
		if exceeded {
			// 6. 處理觸發日誌
			if err := s.processTriggerLog(rule, metricRule, *newState, value, severity, currentTime); err != nil {
				s.logger.Error("處理觸發日誌失敗",
					zap.String("rule_id", string(rule.ID)),
					zap.Error(err),
					zap.String("resource", payload.Metadata.ResourceName),
					zap.String("metric", metricKey),
					zap.Float64("value", value),
					zap.String("severity", severity),
					zap.String("state", newState.State))
				continue
			}

			triggeredRuleCounter++
		}
	}

	s.logger.Info("告警檢查完成",
		zap.String("realm", payload.Metadata.RealmName),
		zap.String("resource", payload.Metadata.ResourceName),
		zap.Int("match_rule_count", matchRuleCounter),
		zap.Int("triggered_rule_count", triggeredRuleCounter))

	return nil
}

// CheckSingle 檢查單個規則是否觸發告警
func (s *Service) CheckSingle(rule *models.Rule, metricData []models.MetricValue, currentTime int64) (bool, float64, string) {
	// 獲取對應的 MetricRule
	var metricRule models.MetricRule
	metricRule, exists := s.global.MetricRules[rule.MetricRuleUID]
	if !exists {
		s.logger.Error("找不到對應的 MetricRule",
			zap.String("metric_rule_uid", rule.MetricRuleUID))
		return false, 0, ""
	}

	// 檢查數據是否足夠
	if len(metricData) == 0 {
		s.logger.Warn("沒有數據可供檢查",
			zap.String("rule_id", string(rule.ID)))
		return false, 0, ""
	}

	// 根據 detection_type 選擇檢查方法
	switch metricRule.DetectionType {
	case "absolute":
		return s.checkAbsolute(rule, metricData, metricRule)
	case "amplitude":
		return s.checkAmplitude(rule, metricData, metricRule)
	default:
		// 默認使用 absolute 檢查
		return s.checkAbsolute(rule, metricData, metricRule)
	}
}

// checkAbsolute 檢查絕對值是否超過閾值
func (s *Service) checkAbsolute(rule *models.Rule, metricData []models.MetricValue, metricRule models.MetricRule) (bool, float64, string) {
	// 檢查數據是否足夠
	if len(metricData) == 0 {
		return false, 0, ""
	}

	// 獲取時間窗口（Duration）
	durationSeconds := 300 // 默認 5 分鐘
	if rule.Duration != "" {
		// 解析 duration 字符串，例如 "5m"、"1h" 等
		durationValue, unit := parseDuration(rule.Duration)
		switch unit {
		case "s":
			durationSeconds = durationValue
		case "m":
			durationSeconds = durationValue * 60
		case "h":
			durationSeconds = durationValue * 3600
		case "d":
			durationSeconds = durationValue * 86400
		}
	}

	// 獲取時間窗口內的數據點
	latestTimestamp := metricData[len(metricData)-1].Timestamp
	windowStartTimestamp := latestTimestamp - int64(durationSeconds)

	var windowData []models.MetricValue
	for _, point := range metricData {
		if point.Timestamp >= windowStartTimestamp {
			windowData = append(windowData, point)
		}
	}

	// 如果窗口內沒有足夠的數據點，則使用所有可用的數據點
	if len(windowData) < 2 {
		windowData = metricData
	}

	// 檢查窗口內的所有數據點是否都超過閾值
	exceededCount := 0
	totalPoints := len(windowData)
	var lastValue float64
	var lastSeverity string

	// 檢查每個數據點
	for _, point := range windowData {
		value := point.Value * metricRule.Scale
		lastValue = value // 記錄最後檢查的值

		// 檢查是否超過閾值
		operator := metricRule.Operator
		if operator == "" {
			operator = "gt" // 默認為大於
		}

		// 檢查嚴重程度（從高到低）
		exceeded := false
		if rule.CritThreshold > 0 {
			if (operator == "gt" && value > rule.CritThreshold) ||
				(operator == "lt" && value < rule.CritThreshold) {
				exceeded = true
				lastSeverity = "crit"
			}
		}

		if !exceeded && rule.WarnThreshold != nil && *rule.WarnThreshold > 0 {
			if (operator == "gt" && value > *rule.WarnThreshold) ||
				(operator == "lt" && value < *rule.WarnThreshold) {
				exceeded = true
				lastSeverity = "warn"
			}
		}

		if !exceeded && rule.InfoThreshold != nil && *rule.InfoThreshold > 0 {
			if (operator == "gt" && value > *rule.InfoThreshold) ||
				(operator == "lt" && value < *rule.InfoThreshold) {
				exceeded = true
				lastSeverity = "info"
			}
		}

		if exceeded {
			exceededCount++
		}
	}

	// 檢查是否所有數據點都超過閾值
	// 如果窗口內的所有數據點都超過閾值，則觸發告警
	if exceededCount == totalPoints && totalPoints > 0 {
		return true, lastValue, lastSeverity
	}

	// 如果沒有達到所需的條件，則返回未觸發
	return false, lastValue, ""
}

// checkAmplitude 檢查振幅是否超過閾值
func (s *Service) checkAmplitude(rule *models.Rule, metricData []models.MetricValue, metricRule models.MetricRule) (bool, float64, string) {
	// 檢查數據是否足夠
	if len(metricData) < 2 {
		return false, 0, ""
	}

	// 獲取時間窗口
	durationSeconds := 300 // 默認 5 分鐘
	if rule.Duration != "" {
		// 解析 duration 字符串，例如 "5m"、"1h" 等
		durationValue, unit := parseDuration(rule.Duration)
		switch unit {
		case "s":
			durationSeconds = durationValue
		case "m":
			durationSeconds = durationValue * 60
		case "h":
			durationSeconds = durationValue * 3600
		case "d":
			durationSeconds = durationValue * 86400
		}
	}

	// 獲取時間窗口內的數據點
	latestTimestamp := metricData[len(metricData)-1].Timestamp
	windowStartTimestamp := latestTimestamp - int64(durationSeconds)

	var windowData []models.MetricValue
	for _, point := range metricData {
		if point.Timestamp >= windowStartTimestamp {
			windowData = append(windowData, point)
		}
	}

	// 如果窗口內沒有足夠的數據點，則使用所有可用的數據點
	if len(windowData) < 2 {
		windowData = metricData
	}

	// 計算最大值和最小值
	var maxValue, minValue float64
	maxValue = windowData[0].Value * metricRule.Scale
	minValue = windowData[0].Value * metricRule.Scale

	for _, point := range windowData {
		value := point.Value * metricRule.Scale
		if value > maxValue {
			maxValue = value
		}
		if value < minValue {
			minValue = value
		}
	}

	// 計算振幅
	if minValue == 0 {
		// 避免除以零
		minValue = 0.000001
	}
	amplitude := (maxValue - minValue) / minValue * 100

	// 檢查是否超過閾值
	if rule.CritThreshold > 0 && amplitude > rule.CritThreshold {
		return true, amplitude, "crit"
	}

	if rule.WarnThreshold != nil && *rule.WarnThreshold > 0 && amplitude > *rule.WarnThreshold {
		return true, amplitude, "warn"
	}

	if rule.InfoThreshold != nil && *rule.InfoThreshold > 0 && amplitude > *rule.InfoThreshold {
		return true, amplitude, "info"
	}

	return false, amplitude, ""
}

// 解析 duration 字符串，例如 "5m"、"1h" 等
func parseDuration(duration string) (int, string) {
	if duration == "" {
		return 5, "m" // 默認 5 分鐘
	}

	// 找到數字和單位
	var valueStr string
	var unit string

	for i, c := range duration {
		if c >= '0' && c <= '9' {
			valueStr += string(c)
		} else {
			unit = duration[i:]
			break
		}
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return 5, "m" // 默認 5 分鐘
	}

	return value, unit
}

// CheckPayload 初步驗證 AlertPayload
func (s *Service) CheckPayload(payload models.AlertPayload) error {
	// 檢查 metadata 是否完整
	if payload.Metadata.RealmName == "" {
		return fmt.Errorf("realm_name 不能為空 [datasource:%s, resource:%s]",
			payload.Metadata.DatasourceName, payload.Metadata.ResourceName)
	}
	if payload.Metadata.DatasourceName == "" {
		return fmt.Errorf("datasource_name 不能為空 [realm:%s, resource:%s]",
			payload.Metadata.RealmName, payload.Metadata.ResourceName)
	}
	if payload.Metadata.ResourceName == "" {
		return fmt.Errorf("resource_name 不能為空 [realm:%s, datasource:%s]",
			payload.Metadata.RealmName, payload.Metadata.DatasourceName)
	}
	if payload.Metadata.Timestamp == 0 {
		return fmt.Errorf("timestamp 不能為空 [realm:%s, resource:%s, datasource:%s]",
			payload.Metadata.RealmName, payload.Metadata.ResourceName, payload.Metadata.DatasourceName)
	}

	// 檢查 data 是否有數據
	if len(payload.Data) == 0 {
		return fmt.Errorf("metrics 不能為空 [realm:%s, resource:%s, datasource:%s]",
			payload.Metadata.RealmName, payload.Metadata.ResourceName, payload.Metadata.DatasourceName)
	}

	// 檢查每個 metric 的數據格式
	for metricKey, metricData := range payload.Data {
		if len(metricData) == 0 {
			return fmt.Errorf("metric %s 的數據不能為空 [realm:%s, resource:%s]",
				metricKey, payload.Metadata.RealmName, payload.Metadata.ResourceName)
		}

		// 不再嚴格檢查 metric key 格式，允許更靈活的格式
		// 只要確保 metric key 不為空即可
		if metricKey == "" {
			return fmt.Errorf("metric key 不能為空 [realm:%s, resource:%s]",
				payload.Metadata.RealmName, payload.Metadata.ResourceName)
		}

		// 檢查每個數據點是否有 timestamp 和 value，並且確保它們是數字而不是字符串
		for i, point := range metricData {
			if point.Timestamp == 0 {
				return fmt.Errorf("metric %s 的第 %d 個數據點缺少 timestamp 或 timestamp 不是數字 [realm:%s, resource:%s]",
					metricKey, i, payload.Metadata.RealmName, payload.Metadata.ResourceName)
			}

			// 檢查 value 是否為有效的數字
			// 由於 Go 的類型系統，如果 Value 是 float64 類型，這裡不需要額外檢查
			// 但我們可以檢查是否為 NaN 或 Infinity
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				return fmt.Errorf("metric %s 的第 %d 個數據點的 value 不是有效的數字 [realm:%s, resource:%s, value:%v]",
					metricKey, i, payload.Metadata.RealmName, payload.Metadata.ResourceName, point.Value)
			}
		}
	}

	return nil
}

// updateAlertState 更新告警狀態
func (s *Service) updateAlertState(rule models.Rule, triggeredValue float64, severity string, currentTime int64) (*models.RuleState, error) {
	// 獲取當前規則狀態並鎖定
	oldState, err := s.mysql.GetRuleStateAndLock(rule.ID)
	if err != nil {
		return nil, fmt.Errorf("獲取規則狀態失敗 [規則ID:%s, 資源:%s, 分區:%s]: %w",
			string(rule.ID), rule.Target.ResourceName, rule.Target.PartitionName, err)
	}

	if oldState == nil {
		// 如果沒有狀態記錄，創建一個新的
		oldState = &models.RuleState{
			RuleID:         rule.ID,
			State:          "normal",
			ContactState:   "normal",
			ContactCounter: 0,
		}
	}

	// 創建新狀態，初始化為舊狀態的副本
	newState := *oldState

	// 更新最後檢查值
	newState.LastCheckValue = triggeredValue

	// 檢查是否在靜默期內
	inSilencePeriod := false
	if newState.SilenceStartAt != nil && newState.SilenceEndAt != nil {
		if currentTime >= *newState.SilenceStartAt && currentTime < *newState.SilenceEndAt {
			inSilencePeriod = true
		} else if currentTime >= *newState.SilenceEndAt {
			// 靜默期結束，重置靜默期和通知計數器
			newState.SilenceStartAt = nil
			newState.SilenceEndAt = nil
			newState.ContactState = "normal"
			newState.ContactCounter = 0
		}
	}

	// 根據是否觸發異常更新狀態
	if severity != "" {
		// 異常觸發
		newState.LastTriggeredValue = &triggeredValue
		newState.LastTriggeredSeverity = &severity
		newState.LastTriggeredAt = &currentTime

		// 如果是首次觸發，記錄首次觸發時間
		if newState.FirstTriggeredAt == nil {
			newState.FirstTriggeredAt = &currentTime
		}

		// 計算堆疊持續時間
		var duration int64
		if rule.Duration != "" {
			// 解析持續時間字符串，例如 "5m"
			durationStr := rule.Duration
			unit := durationStr[len(durationStr)-1:]
			value, _ := strconv.Atoi(durationStr[:len(durationStr)-1])

			switch unit {
			case "s":
				duration = int64(value)
			case "m":
				duration = int64(value * 60)
			case "h":
				duration = int64(value * 3600)
			default:
				duration = 300 // 默認 5 分鐘
			}
		} else {
			duration = 300 // 默認 5 分鐘
		}

		// 檢查是否超過持續時間
		stackDuration := currentTime - *newState.FirstTriggeredAt

		// 如果超過持續時間，則更新為 alerting
		if stackDuration >= duration {
			// 如果狀態不是 alerting，則更新為 alerting
			if newState.State != "alerting" {
				newState.State = "alerting"
			}

			// 如果狀態是 alerting，則增加通知計數器
			if newState.State == "alerting" && !inSilencePeriod {
				newState.ContactCounter++

				// 檢查是否達到連續觸發通知的次數
				requiredTimes := rule.Times
				if requiredTimes <= 0 {
					requiredTimes = 1 // 默認至少需要一次
				}

				// 如果達到連續觸發次數，則設置靜默期
				if newState.ContactCounter >= requiredTimes {
					// 設置靜默期開始時間
					newState.SilenceStartAt = &currentTime

					// 計算靜默期結束時間
					var silencePeriod int64
					if rule.SilencePeriod != "" {
						// 解析靜默期字符串，例如 "1h"
						silenceStr := rule.SilencePeriod
						unit := silenceStr[len(silenceStr)-1:]
						value, _ := strconv.Atoi(silenceStr[:len(silenceStr)-1])

						switch unit {
						case "s":
							silencePeriod = int64(value)
						case "m":
							silencePeriod = int64(value * 60)
						case "h":
							silencePeriod = int64(value * 3600)
						case "d":
							silencePeriod = int64(value * 86400)
						default:
							silencePeriod = 3600 // 默認 1 小時
						}
					} else {
						silencePeriod = 3600 // 默認 1 小時
					}

					// 設置靜默期結束時間
					silenceEndAt := currentTime + silencePeriod
					newState.SilenceEndAt = &silenceEndAt

					// 更新通知狀態
					newState.ContactState = "silence"
				}
			}
		}
	} else {
		// 恢復正常
		if newState.State == "alerting" {
			// 如果之前是 alerting 狀態，則更新為 resolved
			newState.State = "resolved"
			// 這裡可以添加恢復告警的邏輯
		} else {
			newState.State = "normal"
		}

		// 清除觸發相關信息
		newState.FirstTriggeredAt = nil
		newState.LastTriggeredSeverity = nil

		// 如果不在靜默期內，重置通知計數器
		if !inSilencePeriod {
			newState.ContactCounter = 0
			newState.ContactState = "normal"
		}
	}

	// 更新數據庫
	if err := s.mysql.UpdateRuleStateWithUpdates(*oldState, newState); err != nil {
		return nil, fmt.Errorf("更新規則狀態失敗 [規則ID:%s, 資源:%s, 分區:%s, 狀態:%s]: %w",
			string(rule.ID), rule.Target.ResourceName, rule.Target.PartitionName, newState.State, err)
	}

	return &newState, nil
}

// updateLastCheckValue 僅更新 LastCheckValue 而不改變其他狀態
func (s *Service) updateLastCheckValue(rule models.Rule, value float64, currentTime int64) (*models.RuleState, error) {
	// 獲取當前規則狀態並鎖定
	oldState, err := s.mysql.GetRuleStateAndLock(rule.ID)
	if err != nil {
		return nil, fmt.Errorf("獲取規則狀態失敗 [規則ID:%s, 資源:%s, 分區:%s]: %w",
			string(rule.ID), rule.Target.ResourceName, rule.Target.PartitionName, err)
	}

	if oldState == nil {
		// 如果沒有狀態記錄，創建一個新的
		oldState = &models.RuleState{
			RuleID:         rule.ID,
			State:          "normal",
			ContactState:   "normal",
			ContactCounter: 0,
		}
	}

	// 創建新狀態，初始化為舊狀態的副本
	newState := *oldState

	// 只更新最後檢查值
	newState.LastCheckValue = value

	s.logger.Debug("更新最後檢查值",
		zap.String("rule_id", string(rule.ID)),
		zap.Float64("value", value),
		zap.String("resource", rule.Target.ResourceName),
		zap.String("partition", rule.Target.PartitionName))

	// 更新數據庫
	if err := s.mysql.UpdateRuleStateWithUpdates(*oldState, newState); err != nil {
		return nil, fmt.Errorf("更新規則狀態的最後檢查值失敗 [規則ID:%s, 資源:%s, 分區:%s, 值:%.2f]: %w",
			string(rule.ID), rule.Target.ResourceName, rule.Target.PartitionName, value, err)
	}

	return &newState, nil
}

// processTriggerLog 建立或更新 TriggeredLog 記錄
func (s *Service) processTriggerLog(rule models.Rule, metricRule models.MetricRule, state models.RuleState, triggeredValue float64, severity string, currentTime int64) error {
	// 檢查是否在靜默期內
	inSilencePeriod := false
	if state.SilenceStartAt != nil && state.SilenceEndAt != nil {
		if currentTime >= *state.SilenceStartAt && currentTime < *state.SilenceEndAt {
			inSilencePeriod = true
			s.logger.Debug("規則處於靜默期，跳過通知",
				zap.String("rule_id", string(rule.ID)),
				zap.Int64("silence_start", *state.SilenceStartAt),
				zap.Int64("silence_end", *state.SilenceEndAt),
				zap.Int64("current_time", currentTime))
		}
	}

	// 如果在靜默期內且通知狀態為 silenced，則跳過通知
	if inSilencePeriod && state.ContactState == "silenced" {
		s.logger.Debug("規則處於靜默期且通知狀態為 silenced，跳過通知",
			zap.String("rule_id", string(rule.ID)))
		return nil
	}

	// 檢查是否需要創建新的 TriggeredLog
	if state.State == "alerting" && (state.LastTriggeredLogID == nil || len(*state.LastTriggeredLogID) == 0) {
		// 需要創建新的 TriggeredLog
		return s.createTriggeredLog(rule, metricRule, state, triggeredValue, severity, currentTime)
	} else if state.State == "alerting" && state.LastTriggeredLogID != nil && len(*state.LastTriggeredLogID) > 0 {
		// 需要更新現有的 TriggeredLog
		return s.updateTriggeredLog(rule, metricRule, state, triggeredValue, severity, currentTime)
	} else if state.State == "resolved" && state.LastTriggeredLogID != nil && len(*state.LastTriggeredLogID) > 0 {
		// 需要標記 TriggeredLog 為已解決
		return s.resolveTriggeredLog(rule, state, currentTime)
	}

	return nil
}

// createTriggeredLog 創建新的 TriggeredLog
func (s *Service) createTriggeredLog(rule models.Rule, metricRule models.MetricRule, state models.RuleState, triggeredValue float64, severity string, currentTime int64) error {
	// 序列化 rule 和 state 為 JSON
	ruleSnapshot, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("序列化 rule 失敗 [規則ID:%s, 資源:%s]: %w",
			string(rule.ID), rule.Target.ResourceName, err)
	}

	stateSnapshot, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("序列化 state 失敗 [規則ID:%s, 狀態:%s]: %w",
			string(rule.ID), state.State, err)
	}

	// 將 JSON 轉換為 map[string]interface{}
	var ruleMap map[string]interface{}
	if err := json.Unmarshal(ruleSnapshot, &ruleMap); err != nil {
		return fmt.Errorf("將 rule 轉換為 map 失敗 [規則ID:%s]: %w",
			string(rule.ID), err)
	}

	var stateMap map[string]interface{}
	if err := json.Unmarshal(stateSnapshot, &stateMap); err != nil {
		return fmt.Errorf("將 state 轉換為 map 失敗 [規則ID:%s]: %w",
			string(rule.ID), err)
	}

	// 將 map[string]interface{} 轉換為 common.JSONMap
	ruleJSONMap := make(common.JSONMap)
	for k, v := range ruleMap {
		// 將所有值轉換為字符串
		ruleJSONMap[k] = fmt.Sprintf("%v", v)
	}

	stateJSONMap := make(common.JSONMap)
	for k, v := range stateMap {
		// 將所有值轉換為字符串
		stateJSONMap[k] = fmt.Sprintf("%v", v)
	}

	// 創建 TriggeredLog
	triggeredLog := models.TriggeredLog{
		NotifyState:       "pending",
		RealmName:         rule.RealmName,
		TriggeredAt:       currentTime,
		LastTriggeredAt:   currentTime,
		ResourceName:      rule.Target.ResourceName,
		PartitionName:     rule.Target.PartitionName,
		MetricRuleUID:     metricRule.UID,
		RuleID:            rule.ID,
		RuleSnapshot:      ruleJSONMap,
		RuleStateSnapshot: stateJSONMap,
		Severity:          severity,
		TriggeredValue:    triggeredValue,
		Threshold:         *state.LastTriggeredValue,
	}

	// 保存到數據庫
	if err := s.mysql.CreateTriggeredLog(triggeredLog); err != nil {
		return fmt.Errorf("創建 TriggeredLog 失敗 [規則ID:%s, 資源:%s, 嚴重性:%s, 值:%.2f]: %w",
			string(rule.ID), rule.Target.ResourceName, severity, triggeredValue, err)
	}

	// 更新 rule_state 的 last_triggered_log_id
	newState := state
	logID := triggeredLog.ID

	// 直接將 logID 賦值給 LastTriggeredLogID
	// 注意：last_triggered_log_id 在資料庫中是 binary(16) 類型
	newState.LastTriggeredLogID = &logID

	// 更新數據庫
	if err := s.mysql.UpdateRuleStateWithUpdates(state, newState); err != nil {
		return fmt.Errorf("更新 rule_state 的 last_triggered_log_id 失敗 [規則ID:%s, 日誌ID:%v]: %w",
			string(rule.ID), logID, err)
	}

	return nil
}

// updateTriggeredLog 更新現有的 TriggeredLog
func (s *Service) updateTriggeredLog(rule models.Rule, metricRule models.MetricRule, state models.RuleState, triggeredValue float64, severity string, currentTime int64) error {
	// 獲取現有的 TriggeredLog
	triggeredLog, err := s.mysql.GetActiveTriggeredLog(rule.ID, rule.Target.ResourceName, metricRule.MetricRawName)
	if err != nil {
		return fmt.Errorf("獲取 TriggeredLog 失敗 [規則ID:%s, 資源:%s, 指標:%s]: %w",
			string(rule.ID), rule.Target.ResourceName, metricRule.MetricRawName, err)
	}

	// 如果沒有找到活動的 TriggeredLog，則創建一個新的
	if triggeredLog == nil {
		return s.createTriggeredLog(rule, metricRule, state, triggeredValue, severity, currentTime)
	}

	// 更新 TriggeredLog
	triggeredLog.LastTriggeredAt = currentTime
	triggeredLog.TriggeredValue = triggeredValue
	triggeredLog.Severity = severity

	// 保存到數據庫
	if err := s.mysql.UpdateTriggeredLog(*triggeredLog); err != nil {
		return fmt.Errorf("更新 TriggeredLog 失敗 [規則ID:%s, 資源:%s, 日誌ID:%v, 嚴重性:%s, 值:%.2f]: %w",
			string(rule.ID), rule.Target.ResourceName, triggeredLog.ID, severity, triggeredValue, err)
	}

	return nil
}

// resolveTriggeredLog 標記 TriggeredLog 為已解決
func (s *Service) resolveTriggeredLog(rule models.Rule, state models.RuleState, currentTime int64) error {
	// 獲取現有的 TriggeredLog
	triggeredLog, err := s.mysql.GetActiveTriggeredLog(rule.ID, rule.Target.ResourceName, "")
	if err != nil {
		return fmt.Errorf("獲取 TriggeredLog 失敗 [規則ID:%s, 資源:%s]: %w",
			string(rule.ID), rule.Target.ResourceName, err)
	}

	// 如果沒有找到活動的 TriggeredLog，則無需處理
	if triggeredLog == nil {
		return nil
	}

	// 標記為已解決
	triggeredLog.ResolvedAt = &currentTime

	// 保存到數據庫
	if err := s.mysql.UpdateTriggeredLog(*triggeredLog); err != nil {
		return fmt.Errorf("更新 TriggeredLog 為已解決失敗 [規則ID:%s, 資源:%s, 日誌ID:%v, 解決時間:%d]: %w",
			string(rule.ID), rule.Target.ResourceName, triggeredLog.ID, currentTime, err)
	}

	// 清除 rule_state 的 last_triggered_log_id
	newState := state
	newState.LastTriggeredLogID = nil

	// 更新數據庫
	if err := s.mysql.UpdateRuleStateWithUpdates(state, newState); err != nil {
		return fmt.Errorf("清除 rule_state 的 last_triggered_log_id 失敗 [規則ID:%s, 狀態:%s]: %w",
			string(rule.ID), newState.State, err)
	}

	return nil
}

// autoApplyContacts 自動套用通知管道
func (s *Service) autoApplyContacts(rules []models.Rule) error {
	// 獲取所有設置了 AutoApply 的通知管道
	autoApplyContacts, err := s.mysql.GetAutoApplyContacts()
	if err != nil {
		return fmt.Errorf("獲取自動套用通知管道失敗: %w", err)
	}

	if len(autoApplyContacts) == 0 {
		s.logger.Info("沒有設置自動套用的通知管道")
		return nil
	}

	for _, rule := range rules {
		// 為每個規則添加自動套用的通知管道
		for _, contact := range autoApplyContacts {
			if err := s.mysql.AddContactToRule(rule.ID, contact.ID); err != nil {
				s.logger.Error("添加通知管道到規則失敗",
					zap.Error(err),
					zap.String("rule_id", string(rule.ID)),
					zap.String("contact_id", string(contact.ID)))
				continue
			}

			s.logger.Info("成功添加通知管道到規則",
				zap.String("rule_id", string(rule.ID)),
				zap.String("contact_id", string(contact.ID)),
				zap.String("contact_name", contact.Name))
		}
	}

	return nil
}
