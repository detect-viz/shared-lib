package rules

import (
	"github.com/detect-viz/shared-lib/models"
)

type Service interface {
	Create(rule *models.Rule) error
	Get(id int64) (models.Rule, error)
	List(realm string) ([]models.Rule, error)
	Update(rule *models.Rule) error
	Delete(id int64) error
	GetMetricRule(id int64) (models.MetricRule, bool)
}
