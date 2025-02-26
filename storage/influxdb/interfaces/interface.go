package interfaces

import (
	"context"
	"time"
)

// MetricPoint 指標數據點
type MetricPoint struct {
	Timestamp time.Time
	Field     string
	Value     interface{}
	Tags      map[string]string
}

// QueryResult 查詢結果
type QueryResult struct {
	Series []Series
}

// Series 數據序列
type Series struct {
	Name    string
	Tags    map[string]string
	Columns []string
	Values  [][]interface{}
}

// ClientConfig InfluxDB 客戶端配置
type ClientConfig struct {
	URL   string
	Token string
	Org   string
}

// QueryConfig 查詢配置
type QueryConfig struct {
	Bucket      string
	Start       int64
	End         int64
	Host        string
	HostTag     string
	Measurement string
	Fields      []string
	AggPeriod   string
	FilterFlux  string
	CustomFlux  string
	CreateEmpty bool
	IsGroup     bool
}

// Client InfluxDB 客戶端介面
type Database interface {
	// Write 寫入數據
	Write(ctx context.Context, bucket string, org string, points []MetricPoint) error

	// Query 查詢數據
	Query(ctx context.Context, org string, query string) (*QueryResult, error)

	// Close 關閉連接
	Close()
}
