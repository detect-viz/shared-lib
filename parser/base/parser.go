package base

import (
	"bytes"
	"shared-lib/models"
	"shared-lib/parser/common"
)

// MetricParser 定義解析器接口
type MetricParser interface {
	Parse(content *bytes.Buffer) (map[string]interface{}, error)
	InitMetricGroups()
	AddMetric(name string, data map[string]interface{})
	GetMetrics() map[string]interface{}
}

// BaseParser 基礎解析器
type BaseParser struct {
	metrics       map[string]interface{}
	MetricsConfig models.MetricSpecConfig
}

// NewBaseParser 創建基礎解析器
func NewBaseParser() *BaseParser {
	return &BaseParser{
		metrics:       make(map[string]interface{}),
		MetricsConfig: models.MetricSpecConfig{},
	}
}

// convertUnit 轉換單位
func (p *BaseParser) convertUnit(value float64, rawUnit, displayUnit string) float64 {
	switch {
	case rawUnit == "bytes" && displayUnit == "GB":
		return value / common.BytesPerGB
	case rawUnit == "bytes" && displayUnit == "MB":
		return value / common.BytesPerMB
	case rawUnit == "bytes" && displayUnit == "KB":
		return value / common.BytesPerKB
	case rawUnit == "bytes" && displayUnit == "bytes":
		return value
	case rawUnit == "%" && displayUnit == "%":
		return value
	case rawUnit == "count" && displayUnit == "count":
		return value
	case rawUnit == "pages" && displayUnit == "pages":
		return value
	case rawUnit == "transfers" && displayUnit == "transfers":
		return value
	case rawUnit == "queues" && displayUnit == "queues":
		return value
	case rawUnit == "errors" && displayUnit == "errors":
		return value
	case rawUnit == "seconds" && displayUnit == "seconds":
		return value
	default:
		return value
	}
}

// AddMetricWithSpec 使用指標規格添加指標
func (p *BaseParser) AddMetricWithSpec(spec models.MetricField, data map[string]interface{}) {
	// 使用 alias_name 作為指標名稱
	metricName := spec.AliasName

	// 轉換單位
	if value, ok := data[common.Value].(float64); ok {
		data[common.Value] = p.convertUnit(value, spec.RawUnit, spec.DisplayUnit)
	}

	// 添加單位信息
	data["raw_unit"] = spec.RawUnit
	data["display_unit"] = spec.DisplayUnit

	// 添加指標
	p.AddMetric(metricName, data)
}

// AddMetricGroup 添加指標組
func (p *BaseParser) AddMetricGroup(name string) {
	p.metrics[name] = []map[string]interface{}{}
}

// AddMetricGroups 批量添加指標組
func (p *BaseParser) AddMetricGroups(names []string) {
	for _, name := range names {
		p.AddMetricGroup(name)
	}
}

// InitMetricGroups 初始化指標組
func (p *BaseParser) InitMetricGroups() {
	// 直接初始化所有可能的指標組
	p.AddMetricGroup("cpu_usage")
	p.AddMetricGroup("system_usage")
	p.AddMetricGroup("mem_total_bytes")
	p.AddMetricGroup("mem_used_bytes")
}

// AddMetric 添加指標數據
func (p *BaseParser) AddMetric(name string, data map[string]interface{}) {
	if metrics, ok := p.metrics[name].([]map[string]interface{}); ok {
		p.metrics[name] = append(metrics, data)
	}
}

// HasData 檢查是否有解析到數據
func (p *BaseParser) HasData() bool {
	for _, data := range p.metrics {
		if metrics, ok := data.([]map[string]interface{}); ok && len(metrics) > 0 {
			return true
		}
	}
	return false
}

// GetMetrics 獲取所有指標
func (p *BaseParser) GetMetrics() map[string]interface{} {
	return p.metrics
}

// ValidateTimestamp 驗證時間戳
func (p *BaseParser) ValidateTimestamp(timestamp int64) error {
	return common.ValidateTimestamp(timestamp)
}

// ValidateName 驗證名稱
func (p *BaseParser) ValidateName(name string) error {
	return common.ValidateName(name)
}
