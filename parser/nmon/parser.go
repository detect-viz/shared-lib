package nmon

import (
	"bufio"
	"bytes"
	"fmt"
	"shared-lib/parser/base"
	"shared-lib/parser/common"
	"strconv"
	"strings"
	"time"
)

// Parser NMON 解析器
type Parser struct {
	*base.BaseParser
}

// NewParser 創建 NMON 解析器
func NewParser() *Parser {
	return &Parser{
		BaseParser: base.NewBaseParser(),
	}
}

// Parse 實現 MetricParser 接口
func (p *Parser) Parse(content *bytes.Buffer) (map[string]interface{}, error) {
	if content == nil {
		return nil, fmt.Errorf(common.ErrNilContent)
	}

	p.InitMetricGroups()
	scanner := bufio.NewScanner(content)

	// 用於存儲時間戳映射
	timeMap := make(map[string]int64)

	// 第一次掃描：收集時間戳
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ZZZZ") {
			parts := strings.Split(line, ",")
			if len(parts) >= 4 {
				timeKey := parts[1]
				timeStr := fmt.Sprintf("%s %s", parts[3], parts[2])
				if t, err := time.Parse("02-Jan-2006 15:04:05", timeStr); err == nil {
					timeMap[timeKey] = t.Unix()
				}
			}
		}
	}

	// 修改掃描器重置的部分
	// 原來的:
	// content.Seek(0, 0)
	// scanner = bufio.NewScanner(content)

	// 改為:
	content = bytes.NewBuffer(content.Bytes())
	scanner = bufio.NewScanner(content)

	// 第二次掃描：收集所有指標數據
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")

		switch {
		case strings.HasPrefix(line, "CPU_ALL"):
			if len(parts) >= 8 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}

				user, err := strconv.ParseFloat(parts[2], 64)
				if err != nil {
					continue
				}
				sys, err := strconv.ParseFloat(parts[3], 64)
				if err != nil {
					continue
				}
				wait, err := strconv.ParseFloat(parts[4], 64)
				if err != nil {
					continue
				}
				idle, err := strconv.ParseFloat(parts[5], 64)
				if err != nil {
					continue
				}
				steal, err := strconv.ParseFloat(parts[6], 64)
				if err != nil {
					continue
				}
				nice := 0.0 // NMON 可能沒有 nice 值

				// CPU 指標
				data := map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     user + sys + nice + steal + wait,
				}
				p.AddMetric(common.CPUUsage, data)
				p.AddMetric(common.IdleUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     idle,
				})
				p.AddMetric(common.SystemUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     sys,
				})
				p.AddMetric(common.UserUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     user,
				})
				p.AddMetric(common.IOWaitUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     wait,
				})
				p.AddMetric(common.NiceUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     nice,
				})
				p.AddMetric(common.StealUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.CPUName:   "total",
					common.Value:     steal,
				})
			}

		case strings.HasPrefix(line, "MEM"):
			if len(parts) >= 7 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}

				total, err := strconv.ParseFloat(parts[2], 64)
				if err != nil {
					continue
				}
				if total <= 0 {
					return nil, fmt.Errorf(common.ErrInvalidValue, total)
				}
				free, err := strconv.ParseFloat(parts[5], 64)
				if err != nil {
					continue
				}
				cache, err := strconv.ParseFloat(parts[6], 64)
				if err != nil {
					continue
				}
				buffer, err := strconv.ParseFloat(parts[7], 64)
				if err != nil {
					continue
				}

				used := total - free - cache - buffer
				if used < 0 {
					return nil, fmt.Errorf(common.ErrInvalidValue, used)
				}

				// 記憶體指標
				data := map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     total * common.BytesPerMB,
				}
				p.AddMetric(common.MemTotalBytes, data)
				p.AddMetric(common.MemUsedBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     used * common.BytesPerMB,
				})
				p.AddMetric(common.CacheBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     cache * common.BytesPerMB,
				})
				p.AddMetric(common.BufferBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     buffer * common.BytesPerMB,
				})
				p.AddMetric(common.MemUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     (used / total) * common.PercentMultiplier,
				})

				// 添加快取和緩衝區使用率
				p.AddMetric(common.CacheUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     (cache / total) * common.PercentMultiplier,
				})
				p.AddMetric(common.BufferUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     (buffer / total) * common.PercentMultiplier,
				})
			}

		case strings.HasPrefix(line, "MEMNEW"):
			if len(parts) >= 4 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}

				swapTotal, err := strconv.ParseFloat(parts[2], 64)
				if err != nil {
					continue
				}
				if swapTotal <= 0 {
					return nil, fmt.Errorf(common.ErrInvalidValue, swapTotal)
				}
				swapFree, err := strconv.ParseFloat(parts[3], 64)
				if err != nil {
					continue
				}
				swapUsed := swapTotal - swapFree
				if swapUsed < 0 {
					return nil, fmt.Errorf(common.ErrInvalidValue, swapUsed)
				}

				p.AddMetric(common.SwapTotalBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     swapTotal * common.BytesPerMB, // MB to bytes
				})
				p.AddMetric(common.SwapUsedBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     swapUsed * common.BytesPerMB,
				})
				p.AddMetric(common.SwapUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     (swapUsed / swapTotal) * common.PercentMultiplier,
				})
			}

		case strings.HasPrefix(line, "VM"):
			if len(parts) >= 3 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}

				pgin, err := strconv.ParseFloat(parts[2], 64)
				if err != nil {
					continue
				}
				pgout, err := strconv.ParseFloat(parts[3], 64)
				if err != nil {
					continue
				}

				p.AddMetric(common.PgIn, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     pgin,
				})
				p.AddMetric(common.PgOut, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     pgout,
				})
				p.AddMetric(common.PagingUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     pgin + pgout,
				})
			}

		case strings.HasPrefix(line, "NET"):
			if len(parts) >= 8 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}

				NetworkName := strings.TrimSpace(parts[2])
				if err := p.ValidateName(NetworkName); err != nil {
					return nil, err
				}

				read, err := strconv.ParseFloat(parts[3], 64)
				if err != nil {
					continue
				}
				write, err := strconv.ParseFloat(parts[4], 64)
				if err != nil {
					continue
				}
				readPkts, err := strconv.ParseFloat(parts[5], 64)
				if err != nil {
					continue
				}
				writePkts, err := strconv.ParseFloat(parts[6], 64)
				if err != nil {
					continue
				}
				readErrs, err := strconv.ParseFloat(parts[7], 64)
				if err != nil {
					continue
				}
				writeErrs, err := strconv.ParseFloat(parts[8], 64)
				if err != nil {
					continue
				}

				// 網絡指標
				p.AddMetric(common.RecvBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					"network_name":   NetworkName,
					common.Value:     read,
				})
				p.AddMetric(common.SentBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					"network_name":   NetworkName,
					common.Value:     write,
				})
				p.AddMetric(common.RecvPackets, map[string]interface{}{
					common.Timestamp: timestamp,
					"network_name":   NetworkName,
					common.Value:     readPkts,
				})
				p.AddMetric(common.SentPackets, map[string]interface{}{
					common.Timestamp: timestamp,
					"network_name":   NetworkName,
					common.Value:     writePkts,
				})
				p.AddMetric(common.RecvErrs, map[string]interface{}{
					common.Timestamp: timestamp,
					"network_name":   NetworkName,
					common.Value:     readErrs,
				})
				p.AddMetric(common.SentErrs, map[string]interface{}{
					common.Timestamp: timestamp,
					"network_name":   NetworkName,
					common.Value:     writeErrs,
				})
			}

		case strings.HasPrefix(line, "DISK"):
			if len(parts) >= 8 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}

				diskName := strings.TrimSpace(parts[2])
				if err := p.ValidateName(diskName); err != nil {
					return nil, err
				}

				reads, err := strconv.ParseFloat(parts[3], 64)
				if err != nil {
					continue
				}
				writes, err := strconv.ParseFloat(parts[4], 64)
				if err != nil {
					continue
				}
				readBytes, err := strconv.ParseFloat(parts[5], 64)
				if err != nil {
					continue
				}
				writeBytes, err := strconv.ParseFloat(parts[6], 64)
				if err != nil {
					continue
				}
				readTime, err := strconv.ParseFloat(parts[7], 64)
				if err != nil {
					continue
				}
				writeTime, err := strconv.ParseFloat(parts[8], 64)
				if err != nil {
					continue
				}

				// 磁盤指標
				p.AddMetric(common.Reads, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     reads,
				})
				p.AddMetric(common.Writes, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     writes,
				})
				p.AddMetric(common.ReadBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     readBytes * common.BytesPerKB, // KB to bytes
				})
				p.AddMetric(common.WriteBytes, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     writeBytes * common.BytesPerKB,
				})
				p.AddMetric(common.QueueLength, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     (reads + writes) / 2, // 估算值
				})

				// 添加讀寫時間指標
				p.AddMetric(common.ReadTime, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     readTime,
				})
				p.AddMetric(common.WriteTime, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     writeTime,
				})
				p.AddMetric(common.IOTime, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     readTime + writeTime,
				})

				// 添加隊列相關指標
				p.AddMetric(common.Rqueue, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     reads,
				})
				p.AddMetric(common.Wqueue, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     writes,
				})
				p.AddMetric(common.Xfers, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     reads + writes,
				})
				p.AddMetric(common.Busy, map[string]interface{}{
					common.Timestamp: timestamp,
					"disk_name":      diskName,
					common.Value:     ((readTime + writeTime) / common.MilliToSec) * common.PercentMultiplier,
				})
			}

		case strings.HasPrefix(line, "JFSFILE"):
			if len(parts) >= 5 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}
				fsName := strings.TrimSpace(parts[2])
				if err := p.ValidateName(fsName); err != nil {
					return nil, err
				}
				total, err := strconv.ParseFloat(parts[3], 64)
				if err != nil {
					continue
				}
				if total == 0 {
					return nil, fmt.Errorf(common.ErrDivideByZero)
				}
				free, err := strconv.ParseFloat(parts[4], 64)
				if err != nil {
					continue
				}
				used := total - free

				// 檔案系統指標
				p.AddMetric(common.FSTotalBytes, map[string]interface{}{
					common.Timestamp:  timestamp,
					"filesystem_name": fsName,
					common.Value:      total * common.BytesPerMB,
				})
				p.AddMetric(common.FSFreeBytes, map[string]interface{}{
					common.Timestamp:  timestamp,
					"filesystem_name": fsName,
					common.Value:      free * common.BytesPerMB,
				})
				p.AddMetric(common.FSUsedBytes, map[string]interface{}{
					common.Timestamp:  timestamp,
					"filesystem_name": fsName,
					common.Value:      used * common.BytesPerMB,
				})
				p.AddMetric(common.FSUsage, map[string]interface{}{
					common.Timestamp:  timestamp,
					"filesystem_name": fsName,
					common.Value:      (used / total) * common.PercentMultiplier,
				})
			}

		case strings.HasPrefix(line, "TOP"):
			if len(parts) >= 6 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}
				pid, err := strconv.ParseFloat(parts[2], 64)
				if err != nil {
					continue
				}
				procName := strings.TrimSpace(parts[3])
				if err := p.ValidateName(procName); err != nil {
					return nil, err
				}
				cpuUsage, err := strconv.ParseFloat(parts[4], 64)
				if err != nil {
					continue
				}
				memUsage, err := strconv.ParseFloat(parts[5], 64)
				if err != nil {
					continue
				}

				// 進程 CPU 使用率
				p.AddMetric(common.PSCPUUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					"process_name":   procName,
					"pid":            pid,
					common.Value:     cpuUsage,
				})

				// 進程記憶體使用率
				p.AddMetric(common.PSMemUsage, map[string]interface{}{
					common.Timestamp: timestamp,
					"process_name":   procName,
					common.Value:     memUsage,
				})

				// 進程 ID
				p.AddMetric(common.PID, map[string]interface{}{
					common.Timestamp: timestamp,
					"process_name":   procName,
					"pid":            pid,
				})
			}

		case strings.HasPrefix(line, "SYS"):
			if len(parts) >= 3 && timeMap[parts[1]] != 0 {
				timestamp := timeMap[parts[1]]
				if err := p.ValidateTimestamp(timestamp); err != nil {
					return nil, err
				}
				uptime, err := strconv.ParseFloat(parts[2], 64)
				if err != nil {
					continue
				}
				if uptime < 0 {
					return nil, fmt.Errorf(common.ErrInvalidValue, uptime)
				}

				// 系統運行時間
				p.AddMetric(common.Uptime, map[string]interface{}{
					common.Timestamp: timestamp,
					common.Value:     uptime,
				})
			}
		}
	}

	// 添加掃描錯誤檢查
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(common.ErrScannerError, err)
	}

	// 檢查是否有解析到數據
	if !p.HasData() {
		return nil, fmt.Errorf(common.ErrNoValidData)
	}

	return p.GetMetrics(), nil
}

// 添加名稱檢查輔助函數
func (p *Parser) ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf(common.ErrInvalidValue, "empty name")
	}
	return nil
}
