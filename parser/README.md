# Parser
1. parser 負責 解析數據並轉換成標準化 Map，Key 為 metric
2. alert 負責 處理 map[string]interface{}，確保 快速查找 timestamp & value
3. 異常檢測 效能更高，避免 重複解析 JSON, CSV, LOG

`upload_path` : global.EnvConfig.Server.Parser.UploadPath
`error_path` : global.EnvConfig.Server.Parser.ErrorPath
`backup_path` : global.EnvConfig.Server.Parser.BackupPath

## 流程
1. 每 n 分鐘檢查 `upload_path` 有沒有檔案，有的話送到 function 取得 fileinfo 包含 datasource / hostname / date / content 等等，如果不支援種類或出錯就移動到 error_path 

2. partition 過濾
   - disk_name 可以指定過濾 /^veth/
   - 過濾 過去 變化率不大的 partition
   - label by partition 可以顯示 partition 自定義名稱

3. custom_label 可以自定義 rule/ hostname / partition 的 label 一起顯示

3. error_path 可 根據錯誤類型分類，讓 log 格式異常 也能 自動歸類，減少人工排查時間可 根據錯誤類型分類，讓 log 格式異常 也能 自動歸類，減少人工排查時間

4. ProcessData 開始根據 datasource 匹配對應的 function 進行解析
   - `conn` : upload/oracleMonitor/SRVECCDV01_202410181000_ipoc_conn.log  / ipoc_conn.err
   - `tableSpace` : upload/oracleMonitor/SRVECCDV01_202410270000_tableSpace.csv
   - `awrrpt` : upload/adv_121010_121011_202408310700_202408310800.html
   - `logmon` : upload/AL2SUB_08310000.csv
   - `nmon` : upload/SAPLinuxAP01_202408310230.nmon
   - `njmon` : upload/SRVIPOCNEW_20240831_2000.json
  
5. 傳 map[string]interface{} 給 alert 進行異常檢測
   
6. 成功後移到 backup_path 根據 datasource / date 保存

# Performance Data Parsers

這個目錄包含了各種性能數據格式的解析器:

## 支援的數據格式

1. AWR Report (Oracle AWR 報告)
   - 解析 HTML 格式的 AWR 報告
   - 提取 CPU 和記憶體使用率指標
   - 位置: `awr/parser.go`

2. Windows Logman CSV
   - 解析 Windows 性能監控 CSV 數據
   - 提取 CPU、記憶體等系統指標
   - 位置: `logman/parser.go`

3. Linux NMON
   - 解析 Linux NMON 格式數據
   - 提取系統性能指標
   - 位置: `nmon/parser.go`

4. NJMON JSON
   - 解析 NJMON 的 JSON 格式數據
   - 提取系統性能指標
   - 位置: `njmon/parser.go`

## 指標說明

所有解析器都會提取以下標準指標:

### CPU 指標
- cpu_usage: CPU 總使用率
- idle_usage: CPU 空閒率
- system_usage: 系統使用率
- user_usage: 用戶使用率
- iowait_usage: IO 等待率
- nice_usage: Nice 值
- steal_usage: Steal 值

### 記憶體指標
- mem_total_bytes: 總記憶體(bytes)
- mem_used_bytes: 已使用記憶體(bytes)
- cache_bytes: 快取大小(bytes)
- buffer_bytes: 緩衝區大小(bytes)
- mem_usage: 記憶體使用率(%)

## 使用方式

```go
import "shared-lib/parser"

func main() {
    // 解析數據
    metrics, err := parser.ProcessData(fileInfo)
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用解析後的指標數據
    cpuUsage := metrics["cpu_usage"].([]map[string]interface{})
    memUsage := metrics["mem_usage"].([]map[string]interface{})
}
```


```json
{
    "cpu_usage":[
         {
            "timestamp": 1717776000,
            "cpu_name": "total",
            "value": 45.32
         }
    ],
    "mem_usage":[
        {
            "timestamp": 1717776000,
            "value": 45.32
        }
    ],
    "read_bytes":[
        {
            "timestamp": 1717776000,
            "disk_name": "sda",
            "value": 45.32
        }
    ],
    "ps_pcu_usage":[
        {
            "timestamp": 1717776000,
            "pid": 1234,
            "process_name": "java",
            "value": 45.32
        }
    ],
    "fs_usage":[
        {
            "timestamp": 1717776000,
            "filesystem_name": "/",
            "value": 45.32
        }
    ],
    "recv_bytes":[
        {
            "timestamp": 1717776000,
            "network_name": "eth0",
            "value": 45.32
        }
    ]
}
```
