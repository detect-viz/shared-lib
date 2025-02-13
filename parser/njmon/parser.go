package njmon

import (
	"bufio"
	"bytes"
	"encoding/json"

	"shared-lib/parser/common"
	"time"
)

// Parse 解析 NJMON JSON 文件
func Parse(content *bytes.Buffer) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	scanner := bufio.NewScanner(content)

	// 初始化指標組
	var (
		cpuMetrics = map[string][]map[string]interface{}{
			common.CPUUsage:    {},
			common.IdleUsage:   {},
			common.SystemUsage: {},
			common.UserUsage:   {},
			common.IOWaitUsage: {},
			common.NiceUsage:   {},
			common.StealUsage:  {},
		}
		memMetrics = map[string][]map[string]interface{}{
			common.MemTotalBytes: {},
			common.MemUsedBytes:  {},
			common.CacheBytes:    {},
			common.BufferBytes:   {},
			common.MemUsage:      {},
		}
		diskMetrics = map[string][]map[string]interface{}{
			common.Reads:       {},
			common.Writes:      {},
			common.ReadBytes:   {},
			common.WriteBytes:  {},
			common.QueueLength: {},
			common.ReadTime:    {},
			common.WriteTime:   {},
			common.IOTime:      {},
			common.Rqueue:      {},
			common.Wqueue:      {},
			common.Xfers:       {},
			common.Busy:        {},
		}
		networkMetrics = map[string][]map[string]interface{}{
			common.RecvBytes:   {},
			common.SentBytes:   {},
			common.RecvPackets: {},
			common.SentPackets: {},
			common.RecvErrs:    {},
			common.SentErrs:    {},
		}
		fsMetrics = map[string][]map[string]interface{}{
			common.FSTotalBytes: {},
			common.FSFreeBytes:  {},
			common.FSUsedBytes:  {},
			common.FSUsage:      {},
		}
		processMetrics = map[string][]map[string]interface{}{
			common.PSCPUUsage: {},
			common.PSMemUsage: {},
			common.PID:        {},
		}
		systemMetrics = map[string][]map[string]interface{}{
			common.Uptime: {},
		}
	)

	// 逐行讀取並解析 JSON
	for scanner.Scan() {
		line := scanner.Text()

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// 解析時間戳
		var timestamp int64
		if ts, ok := data["timestamp"].(map[string]interface{}); ok {
			if dt, ok := ts["datetime"].(string); ok {
				if t, err := time.Parse("2006-01-02T15:04:05", dt); err == nil {
					timestamp = t.Unix()
				}
			}
		}

		// 解析 CPU 指標
		if cpu, ok := data["cpu"].(map[string]interface{}); ok {
			if total, ok := cpu["cpu_total"].(map[string]interface{}); ok {
				// 用戶使用率
				if user, ok := total["user"].(float64); ok {
					cpuMetrics[common.UserUsage] = append(cpuMetrics[common.UserUsage], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     user,
					})
				}
				// 系統使用率
				if sys, ok := total["sys"].(float64); ok {
					cpuMetrics[common.SystemUsage] = append(cpuMetrics[common.SystemUsage], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     sys,
					})
				}
				// 空閒率
				if idle, ok := total["idle"].(float64); ok {
					cpuMetrics[common.IdleUsage] = append(cpuMetrics[common.IdleUsage], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     idle,
					})
				}
				// IO 等待率
				if wait, ok := total["wait"].(float64); ok {
					cpuMetrics[common.IOWaitUsage] = append(cpuMetrics[common.IOWaitUsage], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     wait,
					})
				}
				// nice 值
				if nice, ok := total["nice"].(float64); ok {
					cpuMetrics[common.NiceUsage] = append(cpuMetrics[common.NiceUsage], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     nice,
					})
				}
				// steal 值
				if steal, ok := total["steal"].(float64); ok {
					cpuMetrics[common.StealUsage] = append(cpuMetrics[common.StealUsage], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     steal,
					})
				}
				// 總 CPU 使用率
				if user, ok := total["user"].(float64); ok {
					if sys, ok := total["sys"].(float64); ok {
						cpuMetrics[common.CPUUsage] = append(cpuMetrics[common.CPUUsage], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     user + sys,
						})
					}
				}
			}
		}

		// 解析記憶體指標
		if mem, ok := data["memory"].(map[string]interface{}); ok {
			if real, ok := mem["Real"].(map[string]interface{}); ok {
				// 總記憶體
				if total, ok := real["total_MB"].(float64); ok {
					memMetrics[common.MemTotalBytes] = append(memMetrics[common.MemTotalBytes], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     total * 1024 * 1024, // MB 轉 bytes
					})
				}
				// 已使用記憶體
				if used, ok := real["used_MB"].(float64); ok {
					memMetrics[common.MemUsedBytes] = append(memMetrics[common.MemUsedBytes], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     used * 1024 * 1024,
					})
				}
				// 快取
				if cache, ok := real["cached_MB"].(float64); ok {
					memMetrics[common.CacheBytes] = append(memMetrics[common.CacheBytes], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     cache * 1024 * 1024,
					})
				}
				// 緩衝區
				if buffer, ok := real["buffers_MB"].(float64); ok {
					memMetrics[common.BufferBytes] = append(memMetrics[common.BufferBytes], map[string]interface{}{
						common.Timestamp: timestamp,
						common.Value:     buffer * 1024 * 1024,
					})
				}
				// 記憶體使用率
				if total, ok := real["total_MB"].(float64); ok {
					if used, ok := real["used_MB"].(float64); ok {
						memMetrics[common.MemUsage] = append(memMetrics[common.MemUsage], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     (used / total) * 100,
						})
					}
				}
			}
		}

		// 解析網路指標
		if networks, ok := data["network"].(map[string]interface{}); ok {
			for nicName, nicData := range networks {
				if nic, ok := nicData.(map[string]interface{}); ok {
					// 接收字節數
					if recv, ok := nic["read_KB"].(float64); ok {
						networkMetrics[common.RecvBytes] = append(networkMetrics[common.RecvBytes], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     recv * 1024,
							"network_name":   nicName,
						})
					}
					// 發送字節數
					if sent, ok := nic["write_KB"].(float64); ok {
						networkMetrics[common.SentBytes] = append(networkMetrics[common.SentBytes], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     sent * 1024,
							"network_name":   nicName,
						})
					}
					// 接收封包數
					if recvPkts, ok := nic["read_packets"].(float64); ok {
						networkMetrics[common.RecvPackets] = append(networkMetrics[common.RecvPackets], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     recvPkts,
							"network_name":   nicName,
						})
					}
					// 發送封包數
					if sentPkts, ok := nic["write_packets"].(float64); ok {
						networkMetrics[common.SentPackets] = append(networkMetrics[common.SentPackets], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     sentPkts,
							"network_name":   nicName,
						})
					}
					// 接收錯誤數
					if recvErrs, ok := nic["read_errors"].(float64); ok {
						networkMetrics[common.RecvErrs] = append(networkMetrics[common.RecvErrs], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     recvErrs,
							"network_name":   nicName,
						})
					}
					// 發送錯誤數
					if sentErrs, ok := nic["write_errors"].(float64); ok {
						networkMetrics[common.SentErrs] = append(networkMetrics[common.SentErrs], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     sentErrs,
							"network_name":   nicName,
						})
					}
				}
			}
		}

		// 解析磁盤指標
		if disks, ok := data["disks"].(map[string]interface{}); ok {
			for diskName, diskData := range disks {
				if disk, ok := diskData.(map[string]interface{}); ok {
					// 讀取字節數
					if readBytes, ok := disk["rkb"].(float64); ok {
						diskMetrics[common.ReadBytes] = append(diskMetrics[common.ReadBytes], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     readBytes * 1024, // KB 轉 bytes
							"disk_name":      diskName,
						})
					}
					// 寫入字節數
					if writeBytes, ok := disk["wkb"].(float64); ok {
						diskMetrics[common.WriteBytes] = append(diskMetrics[common.WriteBytes], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     writeBytes * 1024,
							"disk_name":      diskName,
						})
					}
					// 讀取次數
					if reads, ok := disk["reads"].(float64); ok {
						diskMetrics[common.Reads] = append(diskMetrics[common.Reads], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     reads,
							"disk_name":      diskName,
						})
					}
					// 寫入次數
					if writes, ok := disk["writes"].(float64); ok {
						diskMetrics[common.Writes] = append(diskMetrics[common.Writes], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     writes,
							"disk_name":      diskName,
						})
					}
					// IO 時間
					if ioTime, ok := disk["busy"].(float64); ok {
						diskMetrics[common.IOTime] = append(diskMetrics[common.IOTime], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     ioTime,
							"disk_name":      diskName,
						})
					}
					// 讀取時間
					if readTime, ok := disk["rmsec"].(float64); ok {
						diskMetrics[common.ReadTime] = append(diskMetrics[common.ReadTime], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     readTime,
							"disk_name":      diskName,
						})
					}
					// 寫入時間
					if writeTime, ok := disk["wmsec"].(float64); ok {
						diskMetrics[common.WriteTime] = append(diskMetrics[common.WriteTime], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     writeTime,
							"disk_name":      diskName,
						})
					}
					// 隊列長度
					if queueLen, ok := disk["avgqu-sz"].(float64); ok {
						diskMetrics[common.QueueLength] = append(diskMetrics[common.QueueLength], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     queueLen,
							"disk_name":      diskName,
						})
					}
				}
			}
		}

		// 解析文件系統指標
		if fs, ok := data["filesystems"].(map[string]interface{}); ok {
			for fsName, fsData := range fs {
				if filesystem, ok := fsData.(map[string]interface{}); ok {
					// 總容量
					if total, ok := filesystem["size_MB"].(float64); ok {
						totalBytes := total * 1024 * 1024
						fsMetrics[common.FSTotalBytes] = append(fsMetrics[common.FSTotalBytes], map[string]interface{}{
							common.Timestamp:  timestamp,
							common.Value:      totalBytes,
							"filesystem_name": fsName,
						})
					}
					// 已使用空間
					if used, ok := filesystem["used_MB"].(float64); ok {
						usedBytes := used * 1024 * 1024
						fsMetrics[common.FSUsedBytes] = append(fsMetrics[common.FSUsedBytes], map[string]interface{}{
							common.Timestamp:  timestamp,
							common.Value:      usedBytes,
							"filesystem_name": fsName,
						})
					}
					// 剩餘空間
					if free, ok := filesystem["free_MB"].(float64); ok {
						freeBytes := free * 1024 * 1024
						fsMetrics[common.FSFreeBytes] = append(fsMetrics[common.FSFreeBytes], map[string]interface{}{
							common.Timestamp:  timestamp,
							common.Value:      freeBytes,
							"filesystem_name": fsName,
						})
					}
					// 使用率
					if total, ok := filesystem["size_MB"].(float64); ok {
						if used, ok := filesystem["used_MB"].(float64); ok {
							fsMetrics[common.FSUsage] = append(fsMetrics[common.FSUsage], map[string]interface{}{
								common.Timestamp:  timestamp,
								common.Value:      (used / total) * 100,
								"filesystem_name": fsName,
							})
						}
					}
				}
			}
		}

		// 解析進程指標
		if procs, ok := data["processes"].(map[string]interface{}); ok {
			for procName, procData := range procs {
				if proc, ok := procData.(map[string]interface{}); ok {
					// 進程 ID
					if pid, ok := proc["pid"].(float64); ok {
						processMetrics[common.PID] = append(processMetrics[common.PID], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     pid,
							"process_name":   procName,
						})
					}
					// CPU 使用率
					if cpuUsage, ok := proc["pcpu"].(float64); ok {
						processMetrics[common.PSCPUUsage] = append(processMetrics[common.PSCPUUsage], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     cpuUsage,
							"process_name":   procName,
						})
					}
					// 記憶體使用率
					if memUsage, ok := proc["pmem"].(float64); ok {
						processMetrics[common.PSMemUsage] = append(processMetrics[common.PSMemUsage], map[string]interface{}{
							common.Timestamp: timestamp,
							common.Value:     memUsage,
							"process_name":   procName,
						})
					}
				}
			}
		}

		// 解析系統指標
		if uptime, ok := data["uptime"].(float64); ok {
			systemMetrics[common.Uptime] = append(systemMetrics[common.Uptime], map[string]interface{}{
				common.Timestamp: timestamp,
				common.Value:     uptime,
			})
		}
	}

	// 將所有指標添加到返回的 map 中
	for name, data := range cpuMetrics {
		metrics[name] = data
	}
	for name, data := range memMetrics {
		metrics[name] = data
	}
	for name, data := range diskMetrics {
		metrics[name] = data
	}
	for name, data := range networkMetrics {
		metrics[name] = data
	}
	for name, data := range fsMetrics {
		metrics[name] = data
	}
	for name, data := range processMetrics {
		metrics[name] = data
	}
	for name, data := range systemMetrics {
		metrics[name] = data
	}

	return metrics, nil
}
