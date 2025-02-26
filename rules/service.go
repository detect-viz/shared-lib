package rules

import (
	"sync"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var RuleSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)), // 綁定接口和實現
)

var (
	// Rules 告警規則映射表 realm -> rules
	Rules = make(map[int64][]models.Rule)
	// MetricRules 指標規則映射表 id -> rule
	MetricRules = make(map[int64]models.MetricRule)
	ruleMutex   sync.RWMutex
)

// 將 Service 改名為 serviceImpl
type serviceImpl struct {
	mysql  *mysql.Client
	logger *zap.Logger
}

func NewService(mysql *mysql.Client, logger *zap.Logger) *serviceImpl {
	return &serviceImpl{
		mysql:  mysql,
		logger: logger,
	}
}

func (s *serviceImpl) Create(rule *models.Rule) error {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	return s.mysql.CreateRule(rule)
}

func (s *serviceImpl) Get(id int64) (models.Rule, error) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rule, err := s.mysql.GetRule(id)
	if err != nil {
		return models.Rule{}, err
	}
	return *rule, nil
}

func (s *serviceImpl) List(realm string) ([]models.Rule, error) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rules, err := s.mysql.ListRules(realm)
	if err != nil {
		return []models.Rule{}, err
	}
	return rules, nil
}

func (s *serviceImpl) Update(rule *models.Rule) error {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	return s.mysql.UpdateRule(rule)
}

func (s *serviceImpl) Delete(id int64) error {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	return s.mysql.DeleteRule(id)
}

// GetMetricRule 根據 ID 獲取指標規則
func (s *serviceImpl) GetMetricRule(id int64) (models.MetricRule, bool) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rule, exists := MetricRules[id]
	return rule, exists
}
