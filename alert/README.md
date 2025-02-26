🚀 alert 的完整執行順序

1️⃣ NewService() ➝ 載入 AlertRule，建立 globalRules
2️⃣ Process() ➝ 解析 metrics，執行 CheckSingle()
3️⃣ 異常發生 (alerting) ➝ 寫入 TriggerLog，避免重複
4️⃣ 定期執行 ProcessTriggerLogs() ➝ 查詢 TriggerLog，發送 NotifyLog
5️⃣ 發送異常通知 ➝ 更新 notify_state
6️⃣ 發送恢復通知 (resolved) ➝ 更新 resolved_notify_state


🔹 每次執行檢查：
   1️⃣ 從 `AlertState` 提取 `CheckRule`
   2️⃣ 計算 `stack_duration`
   3️⃣ **若超過 `rule.Duration`，則**
      ➜ `TriggerLog` 只寫入一次
      ➜ 更新 `AlertState` 為 `alerting`
   4️⃣ **若恢復 (`resolved`)，則**
      ➜ `stack_duration = 0`
      ➜ `FirstTriggerTime = 0`
   5️⃣ **發送通知時**
      ➜ `NotifyLog` 只寫入 `sent` 或 `failed`
✅ **目前 `TriggerLog` 記錄了異常的「觸發時間 (`Timestamp`)」，但沒有明確記錄「結束時間」。**  
✅ **如果要完整追蹤告警的「開始時間 & 結束時間」，你可能需要補充 `ResolvedTime` 欄位。**

---

## **📌 1️⃣ 目前 `TriggerLog` 的欄位**
```go
type TriggerLog struct {
    ID                uint     `json:"id" gorm:"primaryKey"`
    UUID              string   `json:"uuid"`
    Timestamp         int64    `json:"timestamp"`   // 異常發生時間（目前只有這個）
    RuleID            int64    `json:"rule_id"`
    ResourceName      string   `json:"resource_name"`
    MetricName        string   `json:"metric_name"`
    Value             float64  `json:"value"`
    Threshold         float64  `json:"threshold"`
    Severity          string   `json:"severity"`
    Status            string   `json:"status"`       // alerting / resolved
    NotifyState       string   `json:"notify_state"` // pending / sent / failed
}
```
📌 **`Timestamp` 只記錄異常發生時間，沒有記錄異常結束 (`resolved`) 時間。**

---

## **📌 2️⃣ `ResolvedTime` 欄位**
🚀 **可以新增 `ResolvedTime` 來記錄異常何時恢復**
```go
type TriggerLog struct {
    ID                uint     `json:"id" gorm:"primaryKey"`
    UUID              string   `json:"uuid"`
    Timestamp         int64    `json:"timestamp"`   // 異常發生時間
    ResolvedTime      *int64   `json:"resolved_time,omitempty"` // 異常結束時間（新增）
    RuleID            int64    `json:"rule_id"`
    ResourceName      string   `json:"resource_name"`
    MetricName        string   `json:"metric_name"`
    Value             float64  `json:"value"`
    Threshold         float64  `json:"threshold"`
    Severity          string   `json:"severity"`
    Status            string   `json:"status"`       // alerting / resolved
    NotifyState       string   `json:"notify_state"` // pending / sent / failed
}
```
📌 **這樣 `TriggerLog` 就能同時記錄異常的「開始時間 (`Timestamp`)」和「結束時間 (`ResolvedTime`)」。**

---

## **📌 3️⃣ 如何更新 `ResolvedTime`**
當 `AlertState` **從 `alerting` 變成 `resolved`** 時，應該更新 `TriggerLog.ResolvedTime`：
```go
if state.RuleState == "resolved" {
    db.Model(&TriggerLog{}).
        Where("rule_id = ? AND status = ?", state.AlertRuleDetailID, "alerting").
        Update("resolved_time", time.Now().Unix())

    db.Model(&TriggerLog{}).
        Where("rule_id = ? AND status = ?", state.AlertRuleDetailID, "alerting").
        Update("status", "resolved")
}
```
✅ **這樣 `TriggerLog` 就能記錄異常何時開始 (`Timestamp`)，何時結束 (`ResolvedTime`)，確保歷史數據完整！**

---

## **📌 4️⃣ `TriggerLog` 更新邏輯**
| **狀態變化** | **`TriggerLog.Timestamp` (開始時間)** | **`TriggerLog.ResolvedTime` (結束時間)** | **`TriggerLog.Status`** |
|-------------|----------------|----------------|----------------|
| **異常發生 (`alerting`)** | `Timestamp = Now()` | `NULL` | `alerting` |
| **異常恢復 (`resolved`)** | `不變` | `ResolvedTime = Now()` | `resolved` |

---

## **📌 5️⃣ `TriggerLog` 新增 `ResolvedTime` 的好處**
✅ **可以完整記錄「異常持續時間」**（`ResolvedTime - Timestamp`）  
✅ **可以幫助 `告警歷史頁面` 顯示異常的完整時間範圍**  
✅ **可以讓 `RecoveryNotify` 知道何時異常結束，以發送恢復通知**  
✅ **不影響 `AlertState` 的即時計算，但能夠提供更完整的歷史記錄**

---

## **📌 6️⃣ `告警歷史 API` 如何查詢異常持續時間**
📌 **修改 `/api/alert-history` API，加入 `ResolvedTime`**
```sql
SELECT 
    t.timestamp AS trigger_time,
    t.resolved_time AS resolved_time,
    t.resource_name,
    t.metric_name,
    t.value AS metric_value,
    t.threshold AS threshold,
    t.severity,
    t.status AS alert_status,
    n.status AS notify_status,
    n.contact_name,
    n.contact_type,
    n.sent_at
FROM trigger_logs t
LEFT JOIN notify_logs n ON t.notify_log_id = n.id
ORDER BY t.timestamp DESC;
```
📌 **前端可以顯示異常持續時間**
```go
duration := trigger.ResolvedTime - trigger.Timestamp
```

📌 **顯示範例**
| **異常時間** | **恢復時間** | **持續時間** | **設備** | **指標** | **數值** | **閾值** | **狀態** |
|---------|---------|---------|--------|--------|--------|--------|--------|
| `15:45:00` | `15:50:00` | `5 分鐘` | `server01` | `cpu_usage` | `95.2%` | `90%` | `已解決` |
| `14:30:00` | `NULL` | `進行中` | `server02` | `mem_usage` | `88.0%` | `85%` | `告警中` |

✅ **這樣 `TriggerLog` 既能記錄異常歷史，也能確保恢復時間可查詢！ 🎯**



{

"fs_usage:/oracle/client": [
		{
			"timestamp": "1725072303",
			"value": "9.0"
		},
		{
			"timestamp": "1725072363",
			"value": "9.0"
		}
],
"fs_free_usage:/oracle/client": [
		{
			"timestamp": "1725072303",
			"value": "9.0"
		},
		{
			"timestamp": "1725072363",
			"value": "9.0"
		}
],
}

監控送過來資料host會放在參數，所以可以區別，剩下因為有分區，所以為:區隔當作Key，我的告警規則如何自動匹配，告警觸發時訊息如何呈現，原本只有單一分區、沒有分區，但是像是tablespace有兩層分區，我應該要如何靈活處理
host,cpu
host,memory
host,disk,disk_name
host,filesystem,filesystem_name
host,network,network_name
host,database
host,database,tablespace_name

## 分區處理
查找 metric_rules 時，使用 `category` `metric_name` 來精確匹配
數據檢查時根據 `:` 分割來解析 `partition_value`
生成告警訊息時，根據 `partition_tag` 組合適當的資源名稱


分區使用 map[string]string

例如 :
category: `tablespace`
metric_name: `free_bytes`
display_name: `剩餘空間`
scale: `134217728`
unit: `GB`
partition_tag: `host,database,tablespace_name`
partition_tag: `主機,資料庫,表空間`
operator: `<`

[{{Severity}}] 主機 {{host}} 資料庫 {{database}} 表空間 {{tablespace_name}} 剩餘空間低於 {{Threshold}} {{unit}} (當前值: {{TriggerValue}} {{unit}})

[Warning] 主機 `SAP01` 資料庫 `/oracle/client` 表空間 `tbs1` 剩餘空間低於 5 GB (當前值: 3 GB)

"free_bytes:/oracle/client": [
		{
			"timestamp": "1725072303",
			"value": "9.0"
		},
		{
			"timestamp": "1725072363",
			"value": "9.0"
		}
]