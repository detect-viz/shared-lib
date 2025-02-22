🚀 alert 的完整執行順序

1️⃣ NewService() ➝ 載入 AlertRule，建立 globalRules
2️⃣ Process() ➝ 解析 metrics，執行 CheckSingle()
3️⃣ 異常發生 (alerting) ➝ 寫入 TriggerLog，避免重複
4️⃣ 定期執行 ProcessTriggerLogs() ➝ 查詢 TriggerLog，發送 NotifyLog
5️⃣ 發送異常通知 ➝ 更新 notify_state
6️⃣ 發送恢復通知 (resolved) ➝ 更新 resolved_notify_state