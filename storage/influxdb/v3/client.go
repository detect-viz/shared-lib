package v3

import (
	"context"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/storage/influxdb/interfaces"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

type clientV3 struct {
	client *influxdb3.Client
	org    string
}

func NewClient(config *models.InfluxDBConfig) (interfaces.Database, error) {
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:  config.URL,
		Token: config.Token,
	})
	if err != nil {
		return nil, err
	}
	return &clientV3{
		client: client,
		org:    config.Org,
	}, nil
}

// Write 實現寫入數據
func (c *clientV3) Write(ctx context.Context, bucket string, org string, points []interfaces.MetricPoint) error {
	influxPoints := make([]*influxdb3.Point, len(points))

	for i, point := range points {
		p := influxdb3.NewPointWithMeasurement(point.Field)

		for k, v := range point.Tags {
			p.SetTag(k, v)
		}

		p.SetField("value", point.Value)
		p.SetTimestamp(point.Timestamp)

		influxPoints[i] = p
	}

	return c.client.WritePoints(ctx, influxPoints,
		influxdb3.WithPrecision(lineprotocol.Nanosecond))
}

// Query 實現查詢數據
func (c *clientV3) Query(ctx context.Context, org string, query string) (*interfaces.QueryResult, error) {
	result, err := c.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	queryResult := &interfaces.QueryResult{
		Series: make([]interfaces.Series, 0),
	}

	// 使用 map 來分組不同的 measurement
	seriesMap := make(map[string]*interfaces.Series)

	for result.Next() {
		value := result.Value()
		measurement := value["_measurement"].(string)

		// 獲取或創建新的 series
		series, exists := seriesMap[measurement]
		if !exists {
			series = &interfaces.Series{
				Name:    measurement,
				Tags:    make(map[string]string),
				Columns: make([]string, 0),
				Values:  make([][]interface{}, 0),
			}
			// 設置列名
			for k := range value {
				series.Columns = append(series.Columns, k)
			}
			seriesMap[measurement] = series
		}

		// 添加數據行
		row := make([]interface{}, len(series.Columns))
		for i, col := range series.Columns {
			row[i] = value[col]
		}
		series.Values = append(series.Values, row)
	}

	// 將 map 轉換為 slice
	for _, series := range seriesMap {
		queryResult.Series = append(queryResult.Series, *series)
	}

	return queryResult, nil
}

// Close 實現關閉連接
func (c *clientV3) Close() {
	c.client.Close()
}
