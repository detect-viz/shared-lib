package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"os"
	"path/filepath"
	"strings"
	"time"

	"shared-lib/models"
	"shared-lib/parser/common"
	"shared-lib/parser/logman"
	"shared-lib/parser/njmon"
	"shared-lib/parser/nmon"
	"shared-lib/parser/oracle"
)

// 定義日期格式常量
const (
	FormatYYYYMMDDHHMM = "200601021504"
	FormatYYYYMMDDHH   = "2006013115"
	FormatYYYYMMDD     = "20060102"
)

func (s *Service) DetectDataSourceType(filename string, content string) (sourceType string, hostname string) {
	// 1. AWR Report
	if strings.Contains(content, "AWR Report for DB:") {
		sourceType = s.config.MappingTag.Source.OracleAwrrpt
		// 從文件名提取 hostname，例如: adv_121022_121023_202408311900_202408312000.html
		hostname = strings.Split(filename, "_")[0]
		return
	}

	// 2. Windows Logman
	if strings.Contains(content, "PDH-CSV") && strings.Contains(content, "System\\") {
		sourceType = s.config.MappingTag.Source.Logman
		// 從文件名提取 hostname，例如: AL2SUB_08310000.csv
		hostname = strings.Split(filename, "_")[0]
		return
	}

	// 3. Linux NMON
	if strings.Contains(content, "AAA,progname,nmon") {
		sourceType = s.config.MappingTag.Source.Nmon
		// 從 AAA,host 行提取主機名
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "AAA,host,") {
				hostname = strings.TrimPrefix(line, "AAA,host,")
				break
			}
		}
		return
	}

	// 4. Linux NJMON
	if strings.Contains(content, "njmon_version") && strings.Contains(content, "njmon4Linux") {
		sourceType = s.config.MappingTag.Source.Njmon
		// 從 identity.hostname 提取主機名
		var data map[string]interface{}
		json.Unmarshal([]byte(content), &data)
		if identity, ok := data["identity"].(map[string]interface{}); ok {
			hostname = identity["hostname"].(string)
		}
		return
	}

	// 5. Oracle TableSpace
	if strings.Contains(content, "tablespace_name,autoextensible,files_in_tablespace") {
		sourceType = s.config.MappingTag.Source.OracleTablespace
		// 從最後一列提取 hostname
		lines := strings.Split(content, "\n")
		if len(lines) > 1 {
			fields := strings.Split(lines[1], ",")
			if len(fields) > 12 {
				hostname = strings.TrimSpace(fields[12])
			}
		}
		return
	}

	// 6. Oracle Connection
	if strings.Contains(filename, "ipoc_conn") {
		sourceType = s.config.MappingTag.Source.OracleConnection
		// 從文件名提取 hostname，例如: SRVECCDV01_202412040945_ipoc_conn.err
		parts := strings.Split(filename, "_")
		if len(parts) > 0 {
			hostname = parts[0]
		}
		return
	}

	return "unknown", ""
}

func CheckFile(filename string) error {
	svc := &Service{} // 如果需要配置，則傳入配置
	// 1. 先獲取文件資訊
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("stat file error: %v", err)
	}

	// 2. 打開文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	if fileInfo.Size() > 100*1024*1024 { // 大於 100MB
		// 使用 Scanner 處理大文件
		scanner := bufio.NewScanner(file)
		if fileInfo.Size() > 512*1024*1024 { // 大於 512MB 設置更大的 buffer
			buf := make([]byte, 1024*1024)
			scanner.Buffer(buf, 1024*1024)
		}

		for scanner.Scan() {
			buffer.Write(scanner.Bytes())
			buffer.WriteByte('\n')
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan file error: %v", err)
		}
	} else {
		// 小文件使用 bufio.NewReader
		reader := bufio.NewReader(file)
		_, err = io.Copy(&buffer, reader)
		if err != nil {
			return fmt.Errorf("read file error: %v", err)
		}
	}

	sourceType, hostname := svc.DetectDataSourceType(
		filepath.Base(filename),
		buffer.String(),
	)
	fmt.Printf("檔案類型: %s, 主機名: %s\n", sourceType, hostname)
	return nil
}

func DetectFileInfo(filename string, content string) (models.FileInfo, error) {
	var info models.FileInfo

	// 1. 從檔名提取日期
	extractDate := func(filename string) time.Time {
		var t time.Time
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		parts := strings.Split(name, "_")

		for _, part := range parts {
			// 嘗試不同的日期格式
			formats := []string{
				FormatYYYYMMDDHHMM, // 202412040945
				FormatYYYYMMDDHH,   // 2024083100
				FormatYYYYMMDD,     // 20240831
			}

			for _, format := range formats {
				if len(part) == len(format) {
					if parsed, err := time.ParseInLocation(format, part, time.Local); err == nil {
						return parsed
					}
				}
			}
		}
		return t
	}

	// 2. 判斷數據源類型和主機名
	switch {
	case strings.Contains(content, "AWR Report for DB:"):
		// adv_121022_121023_202408311900_202408312000.html
		info.SourceName = "awrrpt"
		info.Hostname = strings.Split(filename, "_")[0]
		if t := extractDate(filename); !t.IsZero() {
			info.Timestamp = t
		}

	case strings.Contains(content, "PDH-CSV") && strings.Contains(content, "System\\"):
		// AL2SUB_08310000.csv
		info.SourceName = "logman"
		info.Hostname = strings.Split(filename, "_")[0]
		if t := extractDate(filename); !t.IsZero() {
			info.Timestamp = t
		}

	case strings.Contains(content, "AAA,progname,nmon"):
		// SAPLinuxAP01_202408310000.nmon
		info.SourceName = "nmon"
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "AAA,host,") {
				info.Hostname = strings.TrimPrefix(line, "AAA,host,")
				break
			}
		}
		if t := extractDate(filename); !t.IsZero() {
			info.Timestamp = t
		}

	case strings.Contains(content, "njmon_version"):
		// SRVIPOCNEW_20240831_2300.json
		info.SourceName = "njmon"
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(content), &data); err == nil {
			if identity, ok := data["identity"].(map[string]interface{}); ok {
				info.Hostname = identity["hostname"].(string)
			}
			if timestamp, ok := data["timestamp"].(map[string]interface{}); ok {
				if dt, ok := timestamp["datetime"].(string); ok {
					if t, err := time.Parse("2006-01-02T15:04:05", dt); err == nil {
						info.Timestamp = t
					}
				}
			}
		}

	case strings.Contains(content, "tablespace_name,autoextensible"):
		// SRVECCDV01_202408310000_tableSpace.csv
		info.SourceName = "tablespace"
		lines := strings.Split(content, "\n")
		if len(lines) > 1 {
			fields := strings.Split(lines[1], ",")
			if len(fields) > 12 {
				info.Hostname = strings.TrimSpace(fields[12])
			}
		}
		if t := extractDate(filename); !t.IsZero() {
			info.Timestamp = t
		}

	case strings.Contains(filename, "ipoc_conn"):
		// SRVECCDV01_202412040945_ipoc_conn.err
		info.SourceName = "conn"
		parts := strings.Split(filename, "_")
		if len(parts) > 0 {
			info.Hostname = parts[0]
		}
		if t := extractDate(filename); !t.IsZero() {
			info.Timestamp = t
		}
	}

	// 添加錯誤檢查
	if info.SourceName == "" {
		return info, fmt.Errorf("unknown source type for file: %s", filename)
	}
	if info.Hostname == "" {
		return info, fmt.Errorf("cannot detect hostname from file: %s", filename)
	}
	if info.Timestamp.IsZero() {
		return info, fmt.Errorf("cannot parse timestamp from file: %s", filename)
	}

	return info, nil
}

func CheckFileInfo(filename string) {
	content, _ := os.ReadFile(filename)
	fileInfo, err := DetectFileInfo(filepath.Base(filename), string(content))
	if err != nil {
		fmt.Printf("檢查檔案資訊時發生錯誤: %v\n", err)
		return
	}

	fmt.Printf("檔案類型: %s\n", fileInfo.SourceName)
	fmt.Printf("主機名: %s\n", fileInfo.Hostname)
	fmt.Printf("時間戳: %s\n", fileInfo.Timestamp.Format("2006-01-02 15:04:05"))
}

// ProcessData 根據數據源類型解析數據
func (s *Service) ProcessData(file models.FileInfo) (map[string]interface{}, error) {
	switch file.SourceName {
	case s.config.MappingTag.Source.OracleAwrrpt,
		s.config.MappingTag.Source.OracleTablespace,
		s.config.MappingTag.Source.OracleConnection:
		parser := oracle.NewParser(file.SourceName)
		return parser.Parse(file.Content)

	case s.config.MappingTag.Source.Logman:
		return logman.Parse(file.Content)

	case s.config.MappingTag.Source.Nmon:
		return nmon.Parse(file.Content)

	case s.config.MappingTag.Source.Njmon:
		return njmon.Parse(file.Content)

	default:
		return nil, fmt.Errorf("unsupported data source type: %s", file.SourceName)
	}
}

// MetricGroup 定義一組相關的指標
type MetricGroup struct {
	Name    string
	Metrics map[string][]map[string]interface{}
}

// InitMetricGroups 初始化所有指標組
func InitMetricGroups() []MetricGroup {
	return []MetricGroup{
		{
			Name: "CPU",
			Metrics: map[string][]map[string]interface{}{
				common.CPUUsage:    {},
				common.IdleUsage:   {},
				common.SystemUsage: {},
				common.UserUsage:   {},
				common.IOWaitUsage: {},
				common.NiceUsage:   {},
				common.StealUsage:  {},
			},
		},
		{
			Name: "Memory",
			Metrics: map[string][]map[string]interface{}{
				common.MemTotalBytes:  {},
				common.MemUsedBytes:   {},
				common.CacheBytes:     {},
				common.BufferBytes:    {},
				common.MemUsage:       {},
				common.CacheUsage:     {},
				common.BufferUsage:    {},
				common.SwapTotalBytes: {},
				common.SwapUsedBytes:  {},
				common.SwapUsage:      {},
				common.PagingUsage:    {},
				common.PgIn:           {},
				common.PgOut:          {},
			},
		},
		{
			Name: "Network",
			Metrics: map[string][]map[string]interface{}{
				common.SentBytes:   {},
				common.RecvBytes:   {},
				common.SentPackets: {},
				common.RecvPackets: {},
				common.SentErrs:    {},
				common.RecvErrs:    {},
			},
		},
		{
			Name: "Disk",
			Metrics: map[string][]map[string]interface{}{
				common.Busy:        {},
				common.ReadBytes:   {},
				common.WriteBytes:  {},
				common.Reads:       {},
				common.Writes:      {},
				common.Xfers:       {},
				common.Rqueue:      {},
				common.Wqueue:      {},
				common.QueueLength: {},
				common.IOTime:      {},
				common.ReadTime:    {},
				common.WriteTime:   {},
			},
		},
		{
			Name: "Filesystem",
			Metrics: map[string][]map[string]interface{}{
				common.FSTotalBytes: {},
				common.FSFreeBytes:  {},
				common.FSUsedBytes:  {},
				common.FSUsage:      {},
			},
		},
		{
			Name: "Process",
			Metrics: map[string][]map[string]interface{}{
				common.PID:        {},
				common.PSCPUUsage: {},
				common.PSMemUsage: {},
			},
		},
		{
			Name: "System",
			Metrics: map[string][]map[string]interface{}{
				common.Uptime: {},
			},
		},
	}
}
