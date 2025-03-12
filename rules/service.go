package rules

import (
	"fmt"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var RuleSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// RuleChangeCallback 規則變更回調函數類型
type RuleChangeCallback func(rule models.Rule, operation string)

// Service 規則服務
type serviceImpl struct {
	global           *models.GlobalConfig
	mysql            *mysql.Client
	logger           logger.Logger
	onRuleChangeFunc RuleChangeCallback
}

// NewService 創建規則服務
func NewService(mysql *mysql.Client, logger logger.Logger, global *models.GlobalConfig) *serviceImpl {
	return &serviceImpl{
		global: global,
		mysql:  mysql,
		logger: logger,
	}
}

// SetRuleChangeCallback 設置規則變更回調函數
func (s *serviceImpl) SetRuleChangeCallback(callback RuleChangeCallback) {
	s.onRuleChangeFunc = callback
}

// Create 創建規則
func (s *serviceImpl) Create(realm string, ruleResp *models.RuleResponse) (*models.RuleResponse, error) {
	// 將 RuleResponse 轉換為 Rule
	rule := s.FromResponse(*ruleResp)
	rule.RealmName = realm // 設置 realm

	// 創建規則
	createdRule, err := s.mysql.CreateRule(&rule)
	if err != nil {
		s.logger.Error("創建規則失敗", zap.Error(err))
		return nil, err
	}

	// 創建成功後，觸發回調函數
	if s.onRuleChangeFunc != nil {
		s.onRuleChangeFunc(*createdRule, "create")
	}

	// 將 Rule 轉換回 RuleResponse
	response := s.ToResponse(*createdRule)
	return &response, nil
}

// Get 獲取規則
func (s *serviceImpl) Get(realm string, metricRuleUID string, ruleOverview models.RuleOverview) (*models.RuleResponse, error) {
	// 根據 metricRuleUID 查詢規則
	rules, _, err := s.mysql.ListRules(realm, 0, 100)
	if err != nil {
		return nil, err
	}

	// 查找匹配的規則
	for _, rule := range rules {
		if rule.MetricRuleUID == metricRuleUID {
			// 將 Rule 轉換為 RuleResponse
			response := s.ToResponse(rule)
			return &response, nil
		}
	}

	return nil, nil
}

// List 獲取規則列表
func (s *serviceImpl) List(realm string, cursor int64, limit int) ([]models.MetricRuleOverview, int64, error) {
	// 獲取規則列表
	rules, nextCursor, err := s.mysql.ListRules(realm, cursor, limit)
	if err != nil {
		return nil, 0, err
	}

	// 將 Rule 按 MetricRuleUID 分組
	rulesByMetricRuleUID := make(map[string][]models.Rule)
	for _, rule := range rules {
		rulesByMetricRuleUID[rule.MetricRuleUID] = append(rulesByMetricRuleUID[rule.MetricRuleUID], rule)
	}

	// 將分組後的 Rule 轉換為 MetricRuleOverview
	var overviews []models.MetricRuleOverview
	for metricRuleUID, rulesForMetricRule := range rulesByMetricRuleUID {
		// 獲取 MetricRule 信息
		metricRule, ok := s.global.MetricRules[metricRuleUID]
		if !ok {
			s.logger.Warn("獲取 MetricRule 失敗", zap.String("uid", metricRuleUID))
			continue
		}

		// 計算統計信息
		targetCount := 0
		partitionCount := 0
		contactCount := 0
		var ruleResponses []models.RuleResponse

		for _, rule := range rulesForMetricRule {
			targetCount++
			if rule.Target.PartitionName != "" {
				partitionCount++
			}
			contactCount += len(rule.Contacts)

			// 創建 RuleResponse
			ruleResponse := models.RuleResponse{
				ID:            string(rule.ID),
				AutoApply:     rule.AutoApply,
				Enabled:       rule.Enabled,
				InfoThreshold: rule.InfoThreshold,
				WarnThreshold: rule.WarnThreshold,
				CritThreshold: rule.CritThreshold,
				Times:         rule.Times,
				Duration:      rule.Duration,
				SilencePeriod: rule.SilencePeriod,
				Target:        rule.Target,
				// 需要從其他地方獲取 MetricRule 和 Contacts
			}
			ruleResponses = append(ruleResponses, ruleResponse)
		}

		// 創建 MetricRuleOverview
		overview := models.MetricRuleOverview{
			Category:       metricRule.Category,
			MetricRuleUID:  metricRuleUID,
			MetricRuleName: metricRule.Name,
			RuleCount:      len(rulesForMetricRule),
			TargetCount:    targetCount,
			PartitionCount: partitionCount,
			ContactCount:   contactCount,
			Rules:          ruleResponses,
		}
		overviews = append(overviews, overview)
	}

	return overviews, nextCursor, nil
}

// Update 更新規則
func (s *serviceImpl) Update(realm string, ruleResp *models.RuleResponse) (*models.RuleResponse, error) {
	// 獲取現有規則
	existingRule, err := s.mysql.GetRule([]byte(ruleResp.ID))
	if err != nil {
		s.logger.Error("獲取規則失敗", zap.Error(err))
		return nil, err
	}

	// 將 RuleResponse 轉換為 Rule
	updatedRule := s.FromResponse(*ruleResp)
	updatedRule.RealmName = realm // 設置 realm

	// 保留原有的 ID
	updatedRule.ID = existingRule.ID

	// 更新規則
	savedRule, err := s.mysql.UpdateRule(&updatedRule)
	if err != nil {
		s.logger.Error("更新規則失敗", zap.Error(err))
		return nil, err
	}

	// 更新成功後，觸發回調函數
	if s.onRuleChangeFunc != nil {
		s.onRuleChangeFunc(*savedRule, "update")
	}

	// 將 Rule 轉換回 RuleResponse
	response := s.ToResponse(*savedRule)
	return &response, nil
}

// Delete 刪除規則
func (s *serviceImpl) Delete(id string) error {
	// 獲取規則
	rule, err := s.mysql.GetRule([]byte(id))
	if err != nil {
		s.logger.Error("獲取規則失敗", zap.Error(err))
		return err
	}

	// 刪除規則
	if err := s.mysql.DeleteRule([]byte(id)); err != nil {
		s.logger.Error("刪除規則失敗", zap.Error(err))
		return err
	}

	// 刪除成功後，觸發回調函數
	if s.onRuleChangeFunc != nil {
		s.onRuleChangeFunc(*rule, "delete")
	}

	return nil
}

// GetAvailableTarget 獲取可用的監控對象
func (s *serviceImpl) GetAvailableTarget(realm, metricRuleUID string) ([]models.Target, error) {
	// 獲取所有規則
	rules, _, err := s.mysql.ListRules(realm, 0, 1000)
	if err != nil {
		return nil, err
	}

	// 過濾出可用的監控對象
	var availableTargets []models.Target
	for _, rule := range rules {
		if rule.MetricRuleUID != metricRuleUID {
			availableTargets = append(availableTargets, rule.Target)
		}
	}

	return availableTargets, nil
}

// getMetricRuleByUID 根據 UID 獲取 MetricRule
func (s *serviceImpl) getMetricRuleByUID(uid string) (*models.MetricRule, error) {
	// 從 global.MetricRules 中查找對應的 MetricRule
	metricRule, ok := s.global.MetricRules[uid]
	if !ok {
		return nil, fmt.Errorf("metric rule with UID %s not found", uid)
	}

	// 返回匹配的 MetricRule
	return &metricRule, nil
}

// ToResponse 將 Rule 轉換為 RuleResponse
func (s *serviceImpl) ToResponse(rule models.Rule) models.RuleResponse {
	// 獲取 MetricRule 信息
	metricRule, ok := s.global.MetricRules[rule.MetricRuleUID]
	if !ok {
		s.logger.Warn("獲取 MetricRule 失敗", zap.String("uid", rule.MetricRuleUID))
		// 如果找不到對應的 MetricRule，創建一個空的
		metricRule = models.MetricRule{
			UID: rule.MetricRuleUID,
		}
	}

	// 轉換聯絡人
	var contactsResponse []models.RuleContactResponse
	for _, contact := range rule.Contacts {
		contactsResponse = append(contactsResponse, models.RuleContactResponse{
			ID:   string(contact.ID),
			Name: contact.Name,
			Type: contact.ChannelType,
		})
	}

	// 創建 RuleResponse
	response := models.RuleResponse{
		ID:            string(rule.ID),
		AutoApply:     rule.AutoApply,
		Enabled:       rule.Enabled,
		InfoThreshold: rule.InfoThreshold,
		WarnThreshold: rule.WarnThreshold,
		CritThreshold: rule.CritThreshold,
		Times:         rule.Times,
		Duration:      rule.Duration,
		SilencePeriod: rule.SilencePeriod,
		Target:        rule.Target,
		MetricRule:    metricRule,
		Contacts:      contactsResponse,
	}

	return response
}

// FromResponse 將 RuleResponse 轉換為 Rule
func (s *serviceImpl) FromResponse(ruleResp models.RuleResponse) models.Rule {
	// 創建 Rule
	rule := models.Rule{
		RealmName:     "", // 需要在調用時設置
		MetricRuleUID: ruleResp.MetricRule.UID,
		Target: models.Target{
			ResourceName:   ruleResp.Target.ResourceName,
			PartitionName:  ruleResp.Target.PartitionName,
			DatasourceName: ruleResp.Target.DatasourceName,
		},
		AutoApply:     ruleResp.AutoApply,
		Enabled:       ruleResp.Enabled,
		InfoThreshold: ruleResp.InfoThreshold,
		WarnThreshold: ruleResp.WarnThreshold,
		CritThreshold: ruleResp.CritThreshold,
		Times:         ruleResp.Times,
		Duration:      ruleResp.Duration,
		SilencePeriod: ruleResp.SilencePeriod,
	}

	// 如果有 ID，則設置
	if ruleResp.ID != "" {
		id, err := s.parseID(ruleResp.ID)
		if err == nil {
			rule.ID = id
		}
	}

	// 處理聯絡人
	for _, contactResp := range ruleResp.Contacts {
		id, err := s.parseID(contactResp.ID)
		if err != nil {
			s.logger.Warn("解析聯絡人 ID 失敗", zap.Error(err))
			continue
		}
		contact := models.Contact{
			ID:          id,
			Name:        contactResp.Name,
			ChannelType: contactResp.Type,
		}
		rule.Contacts = append(rule.Contacts, contact)
	}

	return rule
}

// parseID 將字符串 ID 轉換為 []byte
func (s *serviceImpl) parseID(idStr string) ([]byte, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// ManualNotify 手動觸發通知
func (s *serviceImpl) ManualNotify(realm string, ruleIDs []string) error {
	// 檢查參數
	if len(ruleIDs) == 0 {
		return fmt.Errorf("規則 ID 列表不能為空")
	}

	// 獲取規則
	var rules []models.Rule
	for _, idStr := range ruleIDs {
		id, err := s.parseID(idStr)
		if err != nil {
			s.logger.Warn("解析規則 ID 失敗", zap.Error(err), zap.String("id", idStr))
			continue
		}

		rule, err := s.mysql.GetRule(id)
		if err != nil {
			s.logger.Warn("獲取規則失敗", zap.Error(err), zap.String("id", idStr))
			continue
		}

		// 檢查規則是否屬於指定的 realm
		if rule.RealmName != realm {
			s.logger.Warn("規則不屬於指定的 realm", zap.String("realm", realm), zap.String("ruleRealm", rule.RealmName))
			continue
		}

		rules = append(rules, *rule)
	}

	if len(rules) == 0 {
		return fmt.Errorf("沒有找到有效的規則")
	}

	// 觸發手動通知
	for _, rule := range rules {
		// 使用回調函數通知告警服務
		if s.onRuleChangeFunc != nil {
			// 使用特殊的操作類型 "manual_notify" 表示手動觸發
			s.onRuleChangeFunc(rule, "manual_notify")
		}
	}

	return nil
}
