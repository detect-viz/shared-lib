package logman

import (
	"bufio"
	"bytes"
	"fmt"
	"shared-lib/models"
	"shared-lib/parser/base"
	"shared-lib/parser/common"
	"strconv"
	"strings"
	"time"
)

// Parser Logman 解析器
type Parser struct {
	*base.BaseParser
	hostname        string
	currentSnapshot string
	timeMap         map[string]int64 // 存儲快照ID和時間戳的映射
	diskNames       []string         // 存儲磁盤名稱列表
	netInterfaces   []string         // 存儲網絡接口列表
}

// NewParser 創建 Logman 解析器
func NewParser() *Parser {
	return &Parser{
		BaseParser:    base.NewBaseParser(),
		timeMap:       make(map[string]int64),
		diskNames:     make([]string, 0),
		netInterfaces: make([]string, 0),
	}
}

// Parse 實現 MetricParser 接口
func (p *Parser) Parse(content *bytes.Buffer) (map[string]interface{}, error) {
	if content == nil {
		return nil, fmt.Errorf(common.ErrNilContent)
	}

	p.InitMetricGroups()
	scanner := bufio.NewScanner(content)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		// 處理主機名
		if fields[0] == "AAA" && fields[1] == "host" {
			p.hostname = fields[2]
			continue
		}

		// 處理磁盤設備列表
		if fields[0] == "AAA" && fields[1] == "disks" {
			if len(fields) > 2 {
				devices := strings.Split(fields[2], " ")
				for _, dev := range devices {
					if dev != "" {
						p.diskNames = append(p.diskNames, strings.TrimSpace(dev))
					}
				}
			}
			continue
		}

		// 處理時間戳
		if fields[0] == "ZZZZ" {
			if len(fields) >= 4 {
				timeStr := fields[2] + " " + fields[3]
				t, err := time.Parse("15:04:05 02-Jan-2006", timeStr)
				if err == nil {
					p.timeMap[fields[1]] = t.Unix()
					p.currentSnapshot = fields[1]
				}
			}
			continue
		}

		// 處理網絡接口列表
		if fields[0] == "NET" && p.currentSnapshot == "" {
			// 跳過前兩個字段(NET和時間戳)
			for i := 2; i < len(fields); i += 2 {
				if iface := strings.TrimSpace(fields[i]); iface != "" {
					p.netInterfaces = append(p.netInterfaces, iface)
				}
			}
			continue
		}

		// 處理各類指標
		switch fields[0] {
		case "CPU_ALL":
			p.parseCPUMetrics(fields)
		case "MEM":
			p.parseMemoryMetrics(fields)
		case "DISKBUSY", "DISKREAD", "DISKWRITE":
			p.parseDiskMetrics(fields)
		case "NET":
			p.parseNetworkMetrics(fields)
		case "NETPACKET":
			p.parseNetworkPackets(fields)
		case "TOP":
			p.parseProcessMetrics(fields)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("掃描錯誤: %w", err)
	}

	if !p.HasData() {
		return nil, fmt.Errorf(common.ErrNoValidData)
	}

	return p.GetMetrics(), nil
}

// parseCPUMetrics 解析 CPU 指標
func (p *Parser) parseCPUMetrics(fields []string) {
	if len(fields) < 8 || p.currentSnapshot == "" {
		return
	}

	user, _ := strconv.ParseFloat(fields[2], 64)
	sys, _ := strconv.ParseFloat(fields[3], 64)
	idle, _ := strconv.ParseFloat(fields[4], 64)
	iowait, _ := strconv.ParseFloat(fields[5], 64)

	baseData := map[string]interface{}{
		common.Timestamp: p.timeMap[p.currentSnapshot],
		common.CPUName:   p.hostname,
	}

	// CPU 使用率
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.CPU.Usage), copyMap(baseData, map[string]interface{}{
		common.Value: 100.0 - idle,
	}))

	// 用戶態 CPU
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.CPU.User), copyMap(baseData, map[string]interface{}{
		common.Value: user,
	}))

	// 系統態 CPU
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.CPU.System), copyMap(baseData, map[string]interface{}{
		common.Value: sys,
	}))

	// IO 等待
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.CPU.IOWait), copyMap(baseData, map[string]interface{}{
		common.Value: iowait,
	}))

	// 空閒
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.CPU.Idle), copyMap(baseData, map[string]interface{}{
		common.Value: idle,
	}))
}

// parseMemoryMetrics 解析記憶體指標
func (p *Parser) parseMemoryMetrics(fields []string) {
	if len(fields) < 16 || p.currentSnapshot == "" {
		return
	}

	total, _ := strconv.ParseFloat(fields[2], 64)
	free, _ := strconv.ParseFloat(fields[3], 64)
	used, _ := strconv.ParseFloat(fields[4], 64)
	virtual, _ := strconv.ParseFloat(fields[6], 64)
	cached, _ := strconv.ParseFloat(fields[10], 64)

	baseData := map[string]interface{}{
		common.Timestamp: p.timeMap[p.currentSnapshot],
	}

	// 總記憶體
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Memory.TotalBytes), copyMap(baseData, map[string]interface{}{
		common.Value: total * common.BytesPerMB,
	}))

	// 已使用記憶體
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Memory.UsedBytes), copyMap(baseData, map[string]interface{}{
		common.Value: used * common.BytesPerMB,
	}))

	// 快取記憶體
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Memory.CacheBytes), copyMap(baseData, map[string]interface{}{
		common.Value: cached * common.BytesPerMB,
	}))

	// 虛擬記憶體
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Memory.VirtualBytes), copyMap(baseData, map[string]interface{}{
		common.Value: virtual * common.BytesPerMB,
	}))

	// 記憶體使用率
	if total > 0 {
		p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Memory.Usage), copyMap(baseData, map[string]interface{}{
			common.Value: (used / total) * 100,
		}))
	}
}

// parseDiskMetrics 解析磁盤指標
func (p *Parser) parseDiskMetrics(fields []string) {
	if len(fields) < len(p.diskNames)+2 || p.currentSnapshot == "" {
		return
	}

	metricType := fields[0]
	var metricSpec models.MetricSpec

	switch metricType {
	case "DISKBUSY":
		metricSpec = p.MetricsConfig.Disk.Busy
	case "DISKREAD":
		metricSpec = p.MetricsConfig.Disk.ReadBytes
	case "DISKWRITE":
		metricSpec = p.MetricsConfig.Disk.WriteBytes
	default:
		return
	}

	// 為每個磁盤添加指標
	for i, diskName := range p.diskNames {
		if i+2 >= len(fields) {
			break
		}

		value, err := strconv.ParseFloat(fields[i+2], 64)
		if err != nil {
			continue
		}

		// 根據指標類型進行單位轉換
		if metricType == "DISKREAD" || metricType == "DISKWRITE" {
			value *= common.BytesPerKB
		}

		p.AddMetricWithSpec(metricSpec, map[string]interface{}{
			common.Timestamp: p.timeMap[p.currentSnapshot],
			common.DiskName:  diskName,
			common.Value:     value,
		})
	}
}

// copyMap 合併兩個 map
func copyMap(base map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

// parseNetworkMetrics 解析網絡指標
func (p *Parser) parseNetworkMetrics(fields []string) {
	if len(fields) < len(p.netInterfaces)*2+2 || p.currentSnapshot == "" {
		return
	}

	// NET 格式: NET,T0001,<interface>,read KB/s,write KB/s,...
	for i, iface := range p.netInterfaces {
		baseIndex := i*2 + 2
		if baseIndex+1 >= len(fields) {
			break
		}

		read, err := strconv.ParseFloat(fields[baseIndex], 64)
		if err != nil {
			continue
		}

		write, err := strconv.ParseFloat(fields[baseIndex+1], 64)
		if err != nil {
			continue
		}

		baseData := map[string]interface{}{
			common.Timestamp:    p.timeMap[p.currentSnapshot],
			common.NetInterface: iface,
		}

		// 讀取流量
		p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Network.SentBytes), copyMap(baseData, map[string]interface{}{
			common.Value: read * common.BytesPerKB,
		}))

		// 寫入流量
		p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Network.RecvBytes), copyMap(baseData, map[string]interface{}{
			common.Value: write * common.BytesPerKB,
		}))
	}
}

// parseNetworkPackets 解析網絡封包指標
func (p *Parser) parseNetworkPackets(fields []string) {
	if len(fields) < len(p.netInterfaces)*2+2 || p.currentSnapshot == "" {
		return
	}

	// NETPACKET 格式: NETPACKET,T0001,<interface>,in packets/s,out packets/s,...
	for i, iface := range p.netInterfaces {
		baseIndex := i*2 + 2
		if baseIndex+1 >= len(fields) {
			break
		}

		packetsIn, err := strconv.ParseFloat(fields[baseIndex], 64)
		if err != nil {
			continue
		}

		packetsOut, err := strconv.ParseFloat(fields[baseIndex+1], 64)
		if err != nil {
			continue
		}

		baseData := map[string]interface{}{
			common.Timestamp:    p.timeMap[p.currentSnapshot],
			common.NetInterface: iface,
		}

		// 接收封包
		p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Network.RecvPackets), copyMap(baseData, map[string]interface{}{
			common.Value: packetsIn,
		}))

		// 發送封包
		p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Network.SentPackets), copyMap(baseData, map[string]interface{}{
			common.Value: packetsOut,
		}))
	}
}

// parseProcessMetrics 解析進程指標
func (p *Parser) parseProcessMetrics(fields []string) {
	if len(fields) < 15 || p.currentSnapshot == "" {
		return
	}

	// TOP 格式: TOP,pid,T0001,CPU%,CPU_User%,CPU_Sys%,Size,ResSize,ResText,ResData,ResShared,Command
	pid := fields[1]
	cpuTotal, _ := strconv.ParseFloat(fields[3], 64)
	memSize, _ := strconv.ParseFloat(fields[6], 64)
	memRes, _ := strconv.ParseFloat(fields[7], 64)
	command := fields[12]

	baseData := map[string]interface{}{
		common.Timestamp:   p.timeMap[p.currentSnapshot],
		common.ProcessID:   pid,
		common.ProcessName: command,
	}

	// CPU 使用率
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Process.PsCPUUsage), copyMap(baseData, map[string]interface{}{
		common.Value: cpuTotal,
	}))

	// 記憶體使用
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Process.PsMemUsage), copyMap(baseData, map[string]interface{}{
		common.Value: memRes * common.BytesPerKB,
	}))

	// 虛擬記憶體
	p.AddMetricWithSpec(models.MetricField(p.MetricsConfig.Process.PsVirtualMem), copyMap(baseData, map[string]interface{}{
		common.Value: memSize * common.BytesPerKB,
	}))
}

// validateTimestamp 驗證時間戳
func (p *Parser) validateTimestamp(timestamp int64) bool {
	now := time.Now().Unix()
	if timestamp <= 0 || timestamp > now {
		return false
	}

	// 檢查數據是否過期
	if now-timestamp > common.MaxDataAge {
		return false
	}

	return true
}
