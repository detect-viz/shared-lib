package interfaces

import (
	"bytes"

	"github.com/detect-viz/shared-lib/models"
)

// 解析器介面
type Parser interface {
	Parse(content *bytes.Buffer) (map[string]interface{}, error)
}

// 解析服務介面
type ParserService interface {
	SetConfig(config *models.ParserConfig)
	Run() error
	ProcessFiles() error
}

// 指標配置介面
type MetricConfig interface {
	GetMetricField(category, name string) (models.MetricField, bool)
	ConvertUnit(value float64) float64
}
