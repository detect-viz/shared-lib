package oracle

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"shared-lib/parser/base"
	"shared-lib/parser/common"
	"shared-lib/parser/models"
	"strconv"
	"strings"
	"time"
)

// TableSpaceParser Oracle 表空間解析器
type TableSpaceParser struct {
	*base.BaseParser
}

// NewTableSpaceParser 創建表空間解析器
func NewTableSpaceParser() *TableSpaceParser {
	return &TableSpaceParser{
		BaseParser: base.NewBaseParser(),
	}
}

// convertUnit 轉換單位
func (p *TableSpaceParser) convertUnit(value float64, rawUnit, displayUnit string) float64 {
	switch {
	case rawUnit == "bytes" && displayUnit == "GB":
		return value / common.BytesPerGB
	case rawUnit == "bytes" && displayUnit == "MB":
		return value / common.BytesPerMB
	case rawUnit == "bytes" && displayUnit == "KB":
		return value / common.BytesPerKB
	case rawUnit == "%" && displayUnit == "%":
		return value
	case rawUnit == "count" && displayUnit == "count":
		return value
	default:
		return value
	}
}

// Parse 實現 Parser 接口
func (p *TableSpaceParser) Parse(content *bytes.Buffer) (map[string]interface{}, error) {
	if content == nil {
		return nil, fmt.Errorf(common.ErrNilContent)
	}

	p.InitMetricGroups()
	reader := csv.NewReader(content)

	// 跳過標題行
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// 檢查記錄長度
		if len(record) < 15 {
			continue
		}

		// 解析基本資訊
		tablespaceName := strings.TrimSpace(record[0])
		sid := strings.TrimSpace(record[12])
		hostname := strings.TrimSpace(record[13])

		// 解析時間戳
		timestamp, err := time.Parse("200601021504", strings.TrimSpace(record[11]))
		if err != nil {
			continue
		}

		baseInfo := map[string]interface{}{
			common.Timestamp:  timestamp.Unix(),
			"hostname":        hostname,
			"sid":             sid,
			"tablespace_name": tablespaceName,
		}

		// 解析自動擴展
		if autoExt := strings.TrimSpace(record[1]); autoExt == "YES" {
			p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Tablespace.Autoextensible), copyMap(baseInfo, map[string]interface{}{
				common.Value: p.convertUnit(1.0, metricSpec.RawUnit, metricSpec.DisplayUnit),
			}))
		}

		// 解析文件數量
		if count, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64); err == nil {
			p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Tablespace.FilesCount), copyMap(baseInfo, map[string]interface{}{
				common.Value: p.convertUnit(count, metricSpec.RawUnit, metricSpec.DisplayUnit),
			}))
		}

		// 解析總空間
		if total, err := strconv.ParseFloat(strings.TrimSpace(record[3]), 64); err == nil {
			metricSpec := p.MetricsConfig.Tablespace.TotalSpaceGB
			p.AddMetric(metricSpec.AliasName, map[string]interface{}{
				common.Timestamp:  timestamp.Unix(),
				"tablespace_name": tablespaceName,
				common.Value:      p.convertUnit(total, metricSpec.RawUnit, metricSpec.DisplayUnit),
				"raw_unit":        metricSpec.RawUnit,
				"display_unit":    metricSpec.DisplayUnit,
			})
		}

		// 解析已使用空間
		if used, err := strconv.ParseFloat(strings.TrimSpace(record[4]), 64); err == nil {
			metricSpec := p.MetricsConfig.Tablespace.UsedSpaceGB
			p.AddMetric(metricSpec.AliasName, map[string]interface{}{
				common.Timestamp:  timestamp.Unix(),
				"tablespace_name": tablespaceName,
				common.Value:      p.convertUnit(used, metricSpec.RawUnit, metricSpec.DisplayUnit),
				"raw_unit":        metricSpec.RawUnit,
				"display_unit":    metricSpec.DisplayUnit,
			})
		}

		// 解析使用率百分比
		if usedPct, err := strconv.ParseFloat(strings.TrimSpace(record[6]), 64); err == nil {
			metricSpec := p.MetricsConfig.Tablespace.UsedPct
			p.AddMetric(metricSpec.AliasName, map[string]interface{}{
				common.Timestamp:  timestamp.Unix(),
				"tablespace_name": tablespaceName,
				common.Value:      p.convertUnit(usedPct, metricSpec.RawUnit, metricSpec.DisplayUnit),
				"raw_unit":        metricSpec.RawUnit,
				"display_unit":    metricSpec.DisplayUnit,
			})
		}

		// 解析自動擴展使用率
		if autoUsedPct, err := strconv.ParseFloat(strings.TrimSpace(record[10]), 64); err == nil {
			metricSpec := p.MetricsConfig.Tablespace.AutoUsedPct
			p.AddMetric(metricSpec.AliasName, map[string]interface{}{
				common.Timestamp:  timestamp.Unix(),
				"tablespace_name": tablespaceName,
				common.Value:      p.convertUnit(autoUsedPct, metricSpec.RawUnit, metricSpec.DisplayUnit),
				"raw_unit":        metricSpec.RawUnit,
				"display_unit":    metricSpec.DisplayUnit,
			})
		}

		// 解析自動擴展剩餘率
		if autoFreePct, err := strconv.ParseFloat(strings.TrimSpace(record[11]), 64); err == nil {
			metricSpec := p.MetricsConfig.Tablespace.AutoFreePct
			p.AddMetric(metricSpec.AliasName, map[string]interface{}{
				common.Timestamp:  timestamp.Unix(),
				"tablespace_name": tablespaceName,
				common.Value:      p.convertUnit(autoFreePct, metricSpec.RawUnit, metricSpec.DisplayUnit),
				"raw_unit":        metricSpec.RawUnit,
				"display_unit":    metricSpec.DisplayUnit,
			})
		}
	}

	if !p.HasData() {
		return nil, fmt.Errorf(common.ErrNoValidData)
	}

	return p.GetMetrics(), nil
}

// 複製 map 的輔助函數
func copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
