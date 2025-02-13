package oracle

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"shared-lib/parser/common"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// AWR 報告中的表格類型
const (
	TableInformation        = "information"
	TableSQLStatistics      = "sql_statistics"
	TableTablespaceIOStats  = "tablespace_io_stats"
	TableIOStat             = "iostat"
	TableUndoSegment        = "undo_segment"
	TableMemoryStatistics   = "memory_statistics"
	TableCacheSizes         = "cache_sizes"
	TableSharedPool         = "shared_pool_statistics"
	TableInstanceEfficiency = "instance_efficiency"
	TableOperatingSystem    = "operating_system"
	TableInstanceActivity   = "instance_activity"
	TableWaitEventHistogram = "wait_event_histogram"
	TableBufferPool         = "buffer_pool_statistics"
	TableSegmentStatistics  = "segment_statistics"
	TableDefaultTables      = "default_tables"
	TableDatabaseSummary    = "database_summary"
)

// 添加常量定義
const (
	// 時間相關常量
	MinTimestamp = 946684800  // 2000-01-01 00:00:00
	MaxTimestamp = 2524579200 // 2050-01-01 00:00:00

	// 數值精度
	DefaultPrecision = 2

	// 默認配置
	DefaultConfigPath = "config/awr.yml"

	// SQL 文本相關
	MaxSQLTextLength = 4000
	MaxSQLBatchSize  = 1000
)

// 添加錯誤類型常量
const (
	errParseHTML          = "解析 HTML 失敗: %v"
	errInvalidDBInfo      = "無效的數據庫信息: %v"
	errInvalidTimestamp   = "無效的時間戳: %v"
	errInvalidMetric      = "無效的指標值: %v"
	errSaveSQLText        = "保存 SQL 文本失敗: %v"
	errTableNotFound      = "找不到指定的表格: %s"
	errInvalidTableFormat = "表格格式無效: %s"
	errEmptyContent       = "empty content"
	errInvalidConfig      = "invalid configuration: %v"
	errParseTime          = "failed to parse time: %v"
)

// AWRConfig AWR 解析器配置
type AWRConfig struct {
	EnabledTables map[string]bool    `yaml:"enabled_tables" json:"enabled_tables"`
	SQLText       SQLTextConfig      `yaml:"sql_text" json:"sql_text"`
	TimeFormat    TimeFormatConfig   `yaml:"time_format" json:"time_format"`
	NumberFormat  NumberFormatConfig `yaml:"number_format" json:"number_format"`
}

// SQLTextConfig SQL 文本相關配置
type SQLTextConfig struct {
	Enable     bool   `yaml:"enable" json:"enable"`
	MaxLength  int    `yaml:"max_length" json:"max_length"`
	BatchSize  int    `yaml:"batch_size" json:"batch_size"`
	SaveToFile bool   `yaml:"save_to_file" json:"save_to_file"`
	SaveToDB   bool   `yaml:"save_to_db" json:"save_to_db"`
	OutputPath string `yaml:"output_path" json:"output_path"`
}

// TimeFormatConfig 時間格式配置
type TimeFormatConfig struct {
	Layout      string   `yaml:"layout" json:"layout"`
	Formats     []string `yaml:"formats" json:"formats"`
	DefaultZone string   `yaml:"default_zone" json:"default_zone"`
}

// NumberFormatConfig 數值格式配置
type NumberFormatConfig struct {
	Precision    int    `yaml:"precision" json:"precision"`
	ThousandSep  string `yaml:"thousand_sep" json:"thousand_sep"`
	DecimalPoint string `yaml:"decimal_point" json:"decimal_point"`
}

// AWRParser AWR 報告解析器
type AWRParser struct {
	// 配置信息
	config    map[string]bool   // 表格解析配置
	awrConfig *AWRConfig        // AWR 特定配置
	sqlTexts  map[string]string // SQL 文本緩存
	errors    []error           // 錯誤收集
}

// NewAWRParser 創建新的 AWR 解析器
func NewAWRParser(config map[string]bool, awrConfig *AWRConfig) *AWRParser {
	if config == nil {
		config = getDefaultConfig()
	}
	if awrConfig == nil {
		awrConfig = getDefaultAWRConfig()
	}
	return &AWRParser{
		config:    config,
		awrConfig: awrConfig,
		sqlTexts:  make(map[string]string),
		errors:    make([]error, 0),
	}
}

// 獲取默認配置
func getDefaultConfig() map[string]bool {
	return map[string]bool{
		TableSQLStatistics:      true,
		TableIOStat:             true,
		TableMemoryStatistics:   true,
		TableInstanceActivity:   true,
		TableOperatingSystem:    true,
		TableWaitEventHistogram: true,
	}
}

// 獲取默認 AWR 配置
func getDefaultAWRConfig() *AWRConfig {
	return &AWRConfig{
		EnabledTables: getDefaultConfig(),
		SQLText: SQLTextConfig{
			Enable:     true,
			MaxLength:  MaxSQLTextLength,
			BatchSize:  MaxSQLBatchSize,
			SaveToFile: true,
			SaveToDB:   true,
			OutputPath: "data/sql_texts",
		},
		TimeFormat: TimeFormatConfig{
			Layout:      "02-Jan-06 15:04:05",
			Formats:     []string{"02-Jan-06 15:04:05", "2006-01-02 15:04:05"},
			DefaultZone: "Local",
		},
		NumberFormat: NumberFormatConfig{
			Precision:    DefaultPrecision,
			ThousandSep:  ",",
			DecimalPoint: ".",
		},
	}
}

// 添加錯誤處理方法
func (p *AWRParser) addError(err error) {
	if err != nil {
		p.errors = append(p.errors, err)
	}
}

// 獲取所有錯誤
func (p *AWRParser) GetErrors() []error {
	return p.errors
}

// 檢查是否有錯誤
func (p *AWRParser) HasErrors() bool {
	return len(p.errors) > 0
}

// 清除錯誤
func (p *AWRParser) ClearErrors() {
	p.errors = make([]error, 0)
}

// Parse 解析 AWR 報告
func (p *AWRParser) Parse(content *bytes.Buffer) (map[string]interface{}, error) {
	// 檢查輸入
	if content == nil || content.Len() == 0 {
		return nil, fmt.Errorf(errEmptyContent)
	}

	// 檢查配置
	if err := p.validateConfig(); err != nil {
		return nil, fmt.Errorf(errInvalidConfig, err)
	}

	// 清除之前的錯誤
	p.ClearErrors()

	// 記錄開始時間
	start := time.Now()
	p.logParseStart("AWR report")

	metrics, err := p.parseContent(content)
	if err != nil {
		p.addError(err)
		p.logError(errParseHTML, err)
		return nil, fmt.Errorf(errParseHTML, err)
	}

	// 驗證解析結果
	if err := validateMetrics(metrics); err != nil {
		p.addError(err)
		p.logError(errInvalidDBInfo, err)
		return nil, fmt.Errorf(errInvalidDBInfo, err)
	}

	// 保存 SQL 文本
	if p.awrConfig.SQLText.Enable {
		if err := p.saveSQLTexts(); err != nil {
			p.addError(err)
			p.logError(errSaveSQLText, err)
			return nil, fmt.Errorf(errSaveSQLText, err)
		}
	}

	// 標準化指標名稱
	normalizedMetrics := make(map[string]interface{})
	for key, value := range metrics {
		normalizedKey := normalizeMetricName(key)
		normalizedMetrics[normalizedKey] = value
	}

	// 記錄完成時間
	p.logParseEnd("AWR report", time.Since(start))

	// 檢查是否有錯誤發生
	if p.HasErrors() {
		return normalizedMetrics, fmt.Errorf("parsing completed with %d errors", len(p.errors))
	}

	return normalizedMetrics, nil
}

// validateConfig 驗證配置
func (p *AWRParser) validateConfig() error {
	if p.config == nil {
		return fmt.Errorf("missing table configuration")
	}
	if p.awrConfig == nil {
		return fmt.Errorf("missing AWR configuration")
	}
	if p.awrConfig.SQLText.Enable {
		if p.awrConfig.SQLText.MaxLength <= 0 {
			return fmt.Errorf("invalid SQL text max length: %d", p.awrConfig.SQLText.MaxLength)
		}
		if p.awrConfig.SQLText.BatchSize <= 0 {
			return fmt.Errorf("invalid SQL text batch size: %d", p.awrConfig.SQLText.BatchSize)
		}
	}
	return nil
}

// parseContent 實際的解析邏輯
func (p *AWRParser) parseContent(content *bytes.Buffer) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 解析 HTML
	doc, err := html.Parse(content)
	if err != nil {
		return nil, err
	}

	// 提取基本信息
	dbInfo := p.parseDBInfo(doc)
	metrics["database_info"] = dbInfo

	// 檢查時間戳
	if startTime, ok := dbInfo["start_time"].(int64); ok {
		if err := checkTimestamp(startTime); err != nil {
			return nil, err
		}
	}

	// 根據配置解析各個表格
	if p.config[TableSQLStatistics] {
		sqlStats, sqlTexts := p.parseSQLStatistics(doc)
		metrics["sql_statistics"] = sqlStats
		p.sqlTexts = sqlTexts
	}

	if p.config[TableIOStat] {
		ioStats := p.parseIOStats(doc)
		metrics["io_statistics"] = ioStats
	}

	if p.config[TableTablespaceIOStats] {
		tbsIOStats := p.parseTablespaceIOStats(doc)
		metrics["tablespace_io"] = tbsIOStats
	}

	if p.config[TableSharedPool] {
		sharedPoolStats := p.parseSharedPoolStats(doc)
		metrics["shared_pool"] = sharedPoolStats
	}

	// 添加新的解析結果
	loadProfile := p.parseLoadProfile(doc)
	metrics["load_profile"] = loadProfile

	instanceEff := p.parseInstanceEfficiency(doc)
	metrics["instance_efficiency"] = instanceEff

	memoryStats := p.parseMemoryStats(doc)
	metrics["memory_statistics"] = memoryStats

	osStats := p.parseOperatingSystem(doc)
	metrics["operating_system"] = osStats

	waitEvents := p.parseWaitEvents(doc)
	metrics["wait_events"] = waitEvents

	bufferStats := p.parseBufferPoolStats(doc)
	metrics["buffer_pool"] = bufferStats

	if p.config[TableUndoSegment] {
		undoStats := p.parseUndoSegmentStats(doc)
		metrics["undo_segment"] = undoStats
	}

	if p.config[TableInstanceActivity] {
		instanceStats := p.parseInstanceActivity(doc)
		metrics["instance_activity"] = instanceStats
	}

	if p.config[TableDatabaseSummary] {
		dbSummary := p.parseDatabaseSummary(doc)
		metrics["database_summary"] = dbSummary
	}

	if p.config[TableSegmentStatistics] {
		segStats := p.parseSegmentStatistics(doc)
		metrics["segment_statistics"] = segStats
	}

	// 添加檢查點統計
	if p.hasTable(doc, "Checkpoint Stats") {
		checkpointStats := p.parseCheckpointStats(doc)
		metrics["checkpoint_statistics"] = checkpointStats
	}

	return metrics, nil
}

// TableCell 表示表格單元格
type TableCell struct {
	Node *html.Node
	Text string
}

// NewTableCell 創建新的表格單元格
func NewTableCell(n *html.Node) TableCell {
	return TableCell{
		Node: n,
		Text: getText(n),
	}
}

// getTableRows 返回表格的所有行
func getTableRows(table *html.Node) [][]TableCell {
	var rows [][]TableCell
	var currentRow []TableCell

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "tr":
				currentRow = []TableCell{}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
				if len(currentRow) > 0 {
					rows = append(rows, currentRow)
				}
				return
			case "td", "th":
				currentRow = append(currentRow, NewTableCell(n))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(table)

	return rows
}

// 修改使用 TableCell 的地方，例如：
func (p *AWRParser) parseDBInfo(n *html.Node) map[string]interface{} {
	info := make(map[string]interface{})
	var startTime, endTime time.Time

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "DB Name") {
				rows := getTableRows(n)
				if len(rows) >= 2 { // 標題行 + 數據行
					cells := rows[1]
					if len(cells) >= 7 {
						info["dbname"] = cells[0].Text
						info["instance"] = cells[2].Text
						info["hostname"] = getHostname(n)
					}
				}
			}
			// 找到快照時間信息
			if containsText(n, "Snap Id") {
				rows := getTableRows(n)
				if len(rows) >= 3 {
					// 解析開始時間
					beginCells := rows[1]
					if len(beginCells) >= 3 {
						timeStr := beginCells[2].Text
						startTime, _ = time.Parse("02-Jan-06 15:04:05", timeStr)
					}
					// 解析結束時間
					endCells := rows[2]
					if len(endCells) >= 3 {
						timeStr := endCells[2].Text
						endTime, _ = time.Parse("02-Jan-06 15:04:05", timeStr)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	// 添加時間信息
	if !startTime.IsZero() {
		info["start_time"] = startTime.Unix()
	}
	if !endTime.IsZero() {
		info["end_time"] = endTime.Unix()
	}

	return info
}

// parseSQLStatistics 解析 SQL 統計信息
func (p *AWRParser) parseSQLStatistics(n *html.Node) ([]map[string]interface{}, map[string]string) {
	stats := make([]map[string]interface{}, 0)
	texts := make(map[string]string)
	var startTime int64

	// 從基本信息中獲取開始時間
	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	// 遍歷 HTML 找到 SQL 統計表格
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "SQL ordered by") {
				rows := getTableRows(n)
				if len(rows) > 1 { // 跳過標題行
					for _, row := range rows[1:] {
						if len(row) >= 5 {
							sqlID := row[0].Text
							executions, _ := strconv.ParseFloat(row[1].Text, 64)
							cpuTime, _ := strconv.ParseFloat(row[2].Text, 64)
							elapsedTime, _ := strconv.ParseFloat(row[3].Text, 64)

							stat := map[string]interface{}{
								"timestamp":    startTime,
								"sql_id":       sqlID,
								"executions":   executions,
								"cpu_time":     cpuTime,
								"elapsed_time": elapsedTime,
							}
							stats = append(stats, stat)

							// 收集 SQL 文本
							if sqlText := row[4].Text; sqlText != "" {
								texts[sqlID] = sqlText
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats, texts
}

// parseIOStats 解析 IO 統計信息
func (p *AWRParser) parseIOStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	// 從基本信息中獲取開始時間
	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	// 遍歷 HTML 找到 IO 統計表格
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Physical Read") || containsText(n, "Physical Write") {
				rows := getTableRows(n)
				if len(rows) > 1 { // 跳過標題行
					for _, row := range rows[1:] {
						if len(row) >= 3 {
							metric := row[0].Text
							value, _ := strconv.ParseFloat(row[1].Text, 64)

							stat := map[string]interface{}{
								"timestamp": startTime,
								"metric":    metric,
								"value":     value,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseLoadProfile 解析負載概況
func (p *AWRParser) parseLoadProfile(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	// 從基本信息中獲取開始時間
	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Load Profile") {
				rows := getTableRows(n)
				if len(rows) > 1 { // 跳過標題行
					for _, row := range rows[1:] {
						if len(row) >= 4 {
							metric := row[0].Text
							perSecond, _ := extractValue(row[1].Text)
							perTrans, _ := extractValue(row[2].Text)

							stat := map[string]interface{}{
								"timestamp":        startTime,
								"metric":           metric,
								"value_per_second": perSecond,
								"value_per_trans":  perTrans,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseInstanceEfficiency 解析實例效率
func (p *AWRParser) parseInstanceEfficiency(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Instance Efficiency Percentages") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 2 {
							metric := row[0].Text
							value, _ := extractValue(row[1].Text)

							stat := map[string]interface{}{
								"timestamp": startTime,
								"metric":    metric,
								"value":     value,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseMemoryStats 解析內存統計
func (p *AWRParser) parseMemoryStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Cache Sizes") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 2 {
							metric := row[0].Text
							value, _ := extractValue(row[1].Text)

							stat := map[string]interface{}{
								"timestamp": startTime,
								"metric":    metric,
								"value":     value,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseOperatingSystem 解析操作系統統計
func (p *AWRParser) parseOperatingSystem(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Operating System Statistics") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 4 {
							metric := row[0].Text
							perSecond, _ := extractValue(row[1].Text)
							perTrans, _ := extractValue(row[2].Text)

							stat := map[string]interface{}{
								"timestamp":        startTime,
								"metric":           metric,
								"value_per_second": perSecond,
								"value_per_trans":  perTrans,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseWaitEvents 解析等待事件統計
func (p *AWRParser) parseWaitEvents(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Wait Event") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 5 {
							event := row[0].Text
							waits, _ := extractValue(row[1].Text)
							timeouts, _ := extractValue(row[2].Text)
							totalWait, _ := extractValue(row[3].Text)
							avgWait, _ := extractValue(row[4].Text)

							stat := map[string]interface{}{
								"timestamp":    startTime,
								"event":        event,
								"waits":        waits,
								"timeouts":     timeouts,
								"total_wait":   totalWait,
								"average_wait": avgWait,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseBufferPoolStats 解析緩衝池統計
func (p *AWRParser) parseBufferPoolStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Buffer Pool Statistics") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 3 {
							metric := row[0].Text
							begin, _ := extractValue(row[1].Text)
							end, _ := extractValue(row[2].Text)

							stat := map[string]interface{}{
								"timestamp":   startTime,
								"metric":      metric,
								"begin_value": begin,
								"end_value":   end,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseTablespaceIOStats 解析表空間 IO 統計
func (p *AWRParser) parseTablespaceIOStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Tablespace IO Stats") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 6 {
							tablespace := row[0].Text
							reads, _ := extractValue(row[1].Text)
							writes, _ := extractValue(row[2].Text)
							readTime, _ := extractValue(row[3].Text)
							writeTime, _ := extractValue(row[4].Text)
							avgIO, _ := extractValue(row[5].Text)

							stat := map[string]interface{}{
								"timestamp":       startTime,
								"tablespace_name": tablespace,
								"reads":           reads,
								"writes":          writes,
								"read_time":       readTime,
								"write_time":      writeTime,
								"average_io_time": avgIO,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseSharedPoolStats 解析共享池統計
func (p *AWRParser) parseSharedPoolStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Shared Pool Statistics") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 3 {
							metric := row[0].Text
							begin, _ := extractValue(row[1].Text)
							end, _ := extractValue(row[2].Text)

							stat := map[string]interface{}{
								"timestamp":   startTime,
								"metric":      metric,
								"begin_value": begin,
								"end_value":   end,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseUndoSegmentStats 解析 Undo 段統計
func (p *AWRParser) parseUndoSegmentStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Undo Segment Summary") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 4 {
							undoSegment := row[0].Text
							activeTransactions, _ := extractValue(row[1].Text)
							maxQueryLength, _ := extractValue(row[2].Text)
							maxExtents, _ := extractValue(row[3].Text)

							stat := map[string]interface{}{
								"timestamp":           startTime,
								"undo_segment":        undoSegment,
								"active_transactions": activeTransactions,
								"max_query_length":    maxQueryLength,
								"max_extents":         maxExtents,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseInstanceActivity 解析實例活動統計
func (p *AWRParser) parseInstanceActivity(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Instance Activity Stats") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 4 {
							statistic := row[0].Text
							total, _ := extractValue(row[1].Text)
							perSecond, _ := extractValue(row[2].Text)
							perTrans, _ := extractValue(row[3].Text)

							stat := map[string]interface{}{
								"timestamp":        startTime,
								"statistic":        statistic,
								"total":            total,
								"value_per_second": perSecond,
								"value_per_trans":  perTrans,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseDatabaseSummary 解析數據庫摘要
func (p *AWRParser) parseDatabaseSummary(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Database Summary") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 3 {
							metric := row[0].Text
							value, _ := extractValue(row[1].Text)
							unit := row[2].Text

							stat := map[string]interface{}{
								"timestamp": startTime,
								"metric":    metric,
								"value":     value,
								"unit":      unit,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseSegmentStatistics 解析段統計信息
func (p *AWRParser) parseSegmentStatistics(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Segments by") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 6 {
							owner := row[0].Text
							tablespace := row[1].Text
							objectType := row[2].Text
							objectName := row[3].Text
							subObject := row[4].Text
							value, _ := extractValue(row[5].Text)

							stat := map[string]interface{}{
								"timestamp":      startTime,
								"owner":          owner,
								"tablespace":     tablespace,
								"object_type":    objectType,
								"object_name":    objectName,
								"subobject_name": subObject,
								"value":          value,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// parseCheckpointStats 解析檢查點統計
func (p *AWRParser) parseCheckpointStats(n *html.Node) []map[string]interface{} {
	stats := make([]map[string]interface{}, 0)
	var startTime int64

	dbInfo := p.parseDBInfo(n)
	if t, ok := dbInfo["start_time"].(int64); ok {
		startTime = t
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, "Checkpoint Stats") {
				rows := getTableRows(n)
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						if len(row) >= 3 {
							metric := row[0].Text
							value, _ := extractValue(row[1].Text)
							perHour, _ := extractValue(row[2].Text)

							stat := map[string]interface{}{
								"timestamp":      startTime,
								"metric":         metric,
								"value":          value,
								"value_per_hour": perHour,
							}
							stats = append(stats, stat)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return stats
}

// 輔助函數：檢查節點是否包含指定文本
func containsText(n *html.Node, text string) bool {
	var contains bool
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode && strings.Contains(n.Data, text) {
			contains = true
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return contains
}

// 輔助函數：獲取節點的文本內容
func getText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var text strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(strings.TrimSpace(n.Data))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return text.String()
}

// 修改 getHostname 函數以使用 TableCell
func getHostname(n *html.Node) string {
	var hostname string
	for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode && sibling.Data == "table" {
			if containsText(sibling, "Host Name") {
				rows := getTableRows(sibling)
				if len(rows) >= 2 && len(rows[1]) >= 1 {
					hostname = rows[1][0].Text
					break
				}
			}
		}
	}
	return hostname
}

// getCellContent 獲取單元格內容，支持數值和文本
func getCellContent(cell TableCell) (string, float64, error) {
	text := strings.TrimSpace(cell.Text)
	value, err := extractValue(text)
	return text, value, err
}

// 添加時間戳檢查輔助函數
func checkTimestamp(timestamp int64) error {
	if timestamp <= 0 {
		return fmt.Errorf(common.ErrInvalidTime, timestamp)
	}
	if timestamp > time.Now().Unix() {
		return fmt.Errorf(common.ErrInvalidTime, timestamp)
	}
	if (time.Now().Unix() - timestamp) > common.MaxDataAge {
		return fmt.Errorf(common.ErrExpiredData, timestamp)
	}
	return nil
}

// 添加輔助函數：解析大小單位
func parseSize(sizeStr string) (float64, error) {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMGT]?B)?`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	switch matches[2] {
	case "KB":
		value *= 1024
	case "MB":
		value *= 1024 * 1024
	case "GB":
		value *= 1024 * 1024 * 1024
	case "TB":
		value *= 1024 * 1024 * 1024 * 1024
	}

	return value, nil
}

// 添加輔助函數：檢查表格是否存在
func (p *AWRParser) hasTable(n *html.Node, tableName string) bool {
	var found bool
	var f func(*html.Node)
	f = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "table" {
			if containsText(n, tableName) {
				found = true
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return found
}

// 添加輔助函數：獲取表格標題
func getTableTitle(table *html.Node) string {
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "th" {
			title = getText(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(table)
	return title
}

// 添加輔助函數：驗證解析結果
func validateMetrics(metrics map[string]interface{}) error {
	if metrics == nil {
		return fmt.Errorf("metrics 為空")
	}

	// 檢查必要的字段
	dbInfo, ok := metrics["database_info"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("缺少數據庫信息")
	}

	// 檢查時間戳
	if _, ok := dbInfo["start_time"].(int64); !ok {
		return fmt.Errorf("缺少開始時間")
	}
	if _, ok := dbInfo["end_time"].(int64); !ok {
		return fmt.Errorf("缺少結束時間")
	}

	return nil
}

// 添加 SQL 文本處理相關函數
func (p *AWRParser) saveSQLTexts() error {
	if len(p.sqlTexts) == 0 {
		return nil
	}

	// 準備批量插入的數據
	values := make([]map[string]interface{}, 0, len(p.sqlTexts))
	for sqlID, sqlText := range p.sqlTexts {
		value := map[string]interface{}{
			"sql_id":     sqlID,
			"text":       sqlText,
			"created_at": time.Now().Unix(),
		}
		values = append(values, value)
	}

	// TODO: 實現數據庫插入邏輯
	return nil
}

// parseAWRTime 解析 AWR 時間
func parseAWRTime(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return time.Time{}, fmt.Errorf(errParseTime, "empty time string")
	}

	// AWR 報告中的時間格式有多種可能
	formats := []string{
		"02-Jan-06 15:04:05",
		"02-Jan-2006 15:04:05",
		"2006-01-02 15:04:05",
		"20060102150405",
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			// 檢查解析出的時間是否合理
			if t.Year() < 2000 || t.Year() > 2050 {
				lastErr = fmt.Errorf("parsed year out of range: %d", t.Year())
				continue
			}
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, fmt.Errorf(errParseTime, lastErr)
}

// 添加指標名稱標準化函數
func normalizeMetricName(name string) string {
	// 移除特殊字符
	name = regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(name, "_")
	// 轉換為小寫
	name = strings.ToLower(name)
	// 移除多餘的下劃線
	name = regexp.MustCompile(`_+`).ReplaceAllString(name, "_")
	// 移除首尾的下劃線
	name = strings.Trim(name, "_")
	return name
}

// logParseStart 記錄解析開始
func (p *AWRParser) logParseStart(filename string) {
	log.Printf("[AWR] 開始解析報告: %s", filename)
}

// logParseEnd 記錄解析完成
func (p *AWRParser) logParseEnd(filename string, duration time.Duration) {
	log.Printf("[AWR] 完成解析報告: %s, 耗時: %v", filename, duration)
}

// logError 記錄錯誤信息
func (p *AWRParser) logError(format string, args ...interface{}) {
	log.Printf("[AWR][ERROR] "+format, args...)
}

// logWarning 記錄警告信息
func (p *AWRParser) logWarning(format string, args ...interface{}) {
	log.Printf("[AWR][WARN] "+format, args...)
}

// logInfo 記錄一般信息
func (p *AWRParser) logInfo(format string, args ...interface{}) {
	log.Printf("[AWR][INFO] "+format, args...)
}

// extractValue 解析文本中的數值
func extractValue(text string) (float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf(errInvalidMetric, "empty value")
	}

	// 移除千分位分隔符
	text = strings.ReplaceAll(text, ",", "")

	// 嘗試解析百分比
	if strings.Contains(text, "%") {
		text = strings.TrimSuffix(text, "%")
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			if val < 0 || val > 100 {
				return 0, fmt.Errorf(errInvalidMetric, "percentage out of range")
			}
			return val / 100.0, nil
		}
	}

	// 嘗試解析帶單位的數值
	if strings.Contains(text, "K") || strings.Contains(text, "M") ||
		strings.Contains(text, "G") || strings.Contains(text, "T") {
		return parseSize(text)
	}

	// 普通數值解析
	re := regexp.MustCompile(`(-?\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return strconv.ParseFloat(matches[1], 64)
	}

	return 0, fmt.Errorf(errInvalidMetric, "invalid number format")
}
