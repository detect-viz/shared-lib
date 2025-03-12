package rules

import (
	"github.com/detect-viz/shared-lib/models"
)

// Service 規則服務接口
type Service interface {
	// SetRuleChangeCallback 設置規則變更回調函數
	SetRuleChangeCallback(callback RuleChangeCallback)

	// Create 創建規則
	Create(realm string, ruleResp *models.RuleResponse) (*models.RuleResponse, error)

	// Get 獲取規則
	Get(realm string, metricRuleUID string, ruleOverview models.RuleOverview) (*models.RuleResponse, error)

	// List 獲取規則列表
	List(realm string, cursor int64, limit int) ([]models.MetricRuleOverview, int64, error)

	// Update 更新規則
	Update(realm string, ruleResp *models.RuleResponse) (*models.RuleResponse, error)

	// Delete 刪除規則
	Delete(id string) error

	// GetAvailableTarget 獲取可用的監控對象
	GetAvailableTarget(realm, metricRuleUID string) ([]models.Target, error)

	// ToResponse 將 Rule 轉換為 RuleResponse
	ToResponse(rule models.Rule) models.RuleResponse

	// FromResponse 將 RuleResponse 轉換為 Rule
	FromResponse(ruleResp models.RuleResponse) models.Rule

	// ManualNotify 手動觸發通知
	ManualNotify(realm string, ruleIDs []string) error
}
