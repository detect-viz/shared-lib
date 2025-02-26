package v2

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/detect-viz/shared-lib/storage/influxdb/interfaces"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// MetricResponse 定義指標回應的介面
type MetricResponse interface {
	GetTimestamp() int64
	GetMetric() string
	GetValue() interface{}
}

// DefaultMetricResponse 預設的指標回應實現
type DefaultMetricResponse struct {
	Timestamp int64       `json:"timestamp"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
}

func (r DefaultMetricResponse) GetTimestamp() int64 {
	return r.Timestamp
}

func (r DefaultMetricResponse) GetMetric() string {
	return r.Metric
}

func (r DefaultMetricResponse) GetValue() interface{} {
	return r.Value
}

// QueryInfluxDB 從 InfluxDB 查詢資料
func QueryInfluxDB(client influxdb2.Client, org string, config interfaces.QueryConfig) ([]interface{}, error) {
	queryAPI := client.QueryAPI(org)

	// 建構 Flux 查詢
	fluxQuery := buildFluxQuery(config)

	// 執行查詢
	result, err := executeQuery(queryAPI, fluxQuery)
	if err != nil {
		return nil, err
	}

	// 處理結果
	return processQueryResult(result)
}

// buildFluxQuery 建構 Flux 查詢語句
func buildFluxQuery(config interfaces.QueryConfig) string {
	var fluxParts []string

	// 基本查詢
	baseQuery := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %v, stop: %v)
		|> filter(fn: (r) => r["%s"] == "%s")
		|> filter(fn: (r) => r["_measurement"] == "%s")
		%s
		|> filter(fn: (r) => `,
		config.Bucket, config.Start, config.End,
		config.HostTag, config.Host, config.Measurement, config.FilterFlux)
	fluxParts = append(fluxParts, baseQuery)

	// Fields 過濾
	for i, field := range config.Fields {
		if i == len(config.Fields)-1 {
			fluxParts = append(fluxParts, fmt.Sprintf(`r["_field"] == "%s")`, field))
		} else {
			fluxParts = append(fluxParts, fmt.Sprintf(`r["_field"] == "%s" or `, field))
		}
	}

	// 時間聚合
	if config.AggPeriod != "" {
		aggregateQuery := fmt.Sprintf(`
		|> aggregateWindow(every: %s, fn: mean, createEmpty: %v)`,
			config.AggPeriod, config.CreateEmpty)
		fluxParts = append(fluxParts, aggregateQuery)
	}

	// Group by
	if config.IsGroup {
		groupQuery := `
		|> group(columns:["_time","_field"])
		|> sum(column:"_value")
		|> group(columns:["_field"])`
		fluxParts = append(fluxParts, groupQuery)
	}

	if config.CustomFlux != "" {
		fluxParts = append(fluxParts, config.CustomFlux)
	}

	// 最後處理
	lastQuery := `
		|> fill(value: 0.0)
		|> toFloat()`
	fluxParts = append(fluxParts, lastQuery)

	return strings.Join(fluxParts, "\n")
}

// executeQuery 執行查詢
func executeQuery(queryAPI api.QueryAPI, fluxQuery string) (*api.QueryTableResult, error) {
	ctx := context.Background()
	return queryAPI.Query(ctx, fluxQuery)
}

// processQueryResult 處理查詢結果
func processQueryResult(result *api.QueryTableResult) ([]interface{}, error) {
	var responses []interface{}

	for result.Next() {
		record := result.Record()
		value := record.Value()
		if v, ok := value.(float64); ok {
			value = math.Round(v*100) / 100
		}
		response := DefaultMetricResponse{
			Timestamp: record.Time().Unix(),
			Metric:    record.Field(),
			Value:     value,
		}
		responses = append(responses, response)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return responses, nil
}
