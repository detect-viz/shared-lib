package v2

import (
	"context"
	"fmt"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/storage/influxdb/interfaces"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// ClientConfig InfluxDB 客戶端配置
type ClientConfig struct {
	URL   string
	Token string
	Org   string
}

type clientV2 struct {
	client influxdb2.Client
	org    string
}

// NewClient 創建新的 InfluxDB 客戶端
func NewClient(config *models.InfluxDBConfig) (interfaces.Database, error) {
	if config.URL == "" || config.Token == "" {
		return nil, fmt.Errorf("missing required InfluxDB configuration")
	}

	client := influxdb2.NewClient(config.URL, config.Token)
	return &clientV2{
		client: client,
		org:    config.Org,
	}, nil
}

func (c *clientV2) Write(ctx context.Context, bucket string, org string, points []interfaces.MetricPoint) error {
	writeAPI := c.client.WriteAPIBlocking(org, bucket)
	for _, point := range points {
		p := influxdb2.NewPoint(point.Field, point.Tags, map[string]interface{}{"value": point.Value}, point.Timestamp)
		if err := writeAPI.WritePoint(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (c *clientV2) Query(ctx context.Context, org string, query string) (*interfaces.QueryResult, error) {
	// 解析 Flux 查詢字符串為 QueryConfig
	config := interfaces.QueryConfig{
		Bucket:      extractBucket(query),
		Start:       extractTimeRange(query, "start"),
		End:         extractTimeRange(query, "stop"),
		Host:        extractFilter(query, "host"),
		HostTag:     "host",
		Measurement: extractMeasurement(query),
		Fields:      extractFields(query),
		AggPeriod:   extractAggPeriod(query),
		FilterFlux:  extractCustomFilters(query),
		CustomFlux:  "",
		CreateEmpty: true,
		IsGroup:     false,
	}

	// 使用 QueryInfluxDB 執行查詢
	results, err := QueryInfluxDB(c.client, org, config)
	if err != nil {
		return nil, err
	}

	// 轉換結果為 QueryResult 格式
	queryResult := &interfaces.QueryResult{
		Series: make([]interfaces.Series, 0),
	}

	// 使用 map 來分組不同的 measurement
	seriesMap := make(map[string]*interfaces.Series)

	// 將結果轉換為 Series
	for _, result := range results {
		if response, ok := result.(DefaultMetricResponse); ok {
			measurement := response.GetMetric()

			series, exists := seriesMap[measurement]
			if !exists {
				series = &interfaces.Series{
					Name:    measurement,
					Tags:    make(map[string]string),
					Columns: []string{"time", "_field", "_value"},
					Values:  make([][]interface{}, 0),
				}
				seriesMap[measurement] = series
			}

			series.Values = append(series.Values, []interface{}{
				response.GetTimestamp(),
				response.GetMetric(),
				response.GetValue(),
			})
		}
	}

	// 將 map 轉換為 slice
	for _, series := range seriesMap {
		queryResult.Series = append(queryResult.Series, *series)
	}

	return queryResult, nil
}

// 輔助函數用於解析 Flux 查詢字符串
func extractBucket(query string) string {
	// 實現從 query 中提取 bucket 名稱的邏輯
	return ""
}

func extractTimeRange(query string, rangeType string) int64 {
	// 實現從 query 中提取時間範圍的邏輯
	return 0
}

func extractFilter(query string, filterName string) string {
	// 實現從 query 中提取過濾條件的邏輯
	return ""
}

func extractMeasurement(query string) string {
	// 實現從 query 中提取 measurement 的邏輯
	return ""
}

func extractFields(query string) []string {
	// 實現從 query 中提取字段列表的邏輯
	return nil
}

func extractAggPeriod(query string) string {
	// 實現從 query 中提取聚合週期的邏輯
	return ""
}

func extractCustomFilters(query string) string {
	// 實現從 query 中提取自定義過濾條件的邏輯
	return ""
}

func (c *clientV2) Close() {
	c.client.Close()
}
