# 告警測試 Payload

本目錄包含根據 `conf/conf.d/metric_rule.yaml` 中定義的告警規則生成的測試 payload。每個規則都有兩個對應的 JSON 檔案：

- `{RULE_ID}_normal.json`: 正常值，不會觸發告警
- `{RULE_ID}_alert.json`: 異常值，會觸發告警

## 使用方法

可以使用 curl 命令將這些 payload 發送到告警服務進行測試：

```bash
curl -X POST -H "Content-Type: application/json" -d @CPU-001_alert.json http://localhost:8080/api/v1/alert
```

或者使用帶有開發模式的 API：

```bash
curl --location --request POST 'http://localhost:8080/api/v1/alert/run-alert' \
--header 'Realm: master' \
--header 'X-Dev-Mode: admin-test' \
--header 'Content-Type: application/json' \
--data @conf/test/CPU-001_alert.json
```

## 規則說明

以下是各規則的簡要說明：

### CPU 相關規則

- CPU-001: CPU 使用率持續高負載 (閾值: 95%)
- CPU-002: IO 等待時間持續過高 (閾值: 70%)
- CPU-003: CPU 使用率突發增長 (閾值: 300%)

### 記憶體相關規則

- MEMORY-001: 記憶體使用量持續過高 (閾值: 2000MB)
- MEMORY-002: 記憶體使用率突發增長 (閾值: 250%)

### 磁碟相關規則

- DISK-001: 磁碟繁忙率持續過高 (閾值: 90%)

### 檔案系統相關規則

- FILESYSTEM-001: 檔案系統使用率持續過高 (閾值: 100%)
- FILESYSTEM-002: 檔案系統可用空間過低 (閾值: 1GB)

### 網路相關規則

- NETWORK-001: 網路發送流量持續過高 (閾值: 400MB/s)

### 資料庫相關規則

- DATABASE-001: 資料庫連線拒絕過多 (閾值: 3 次)
- TABLESPACE-001: 資料庫表可用空間過低 (閾值: 5GB)

## 自動套用通知管道

系統現在支持自動套用通知管道到新創建的規則。要啟用此功能，請按照以下步驟操作：

1. 在資料庫中執行遷移腳本 `000020_add_auto_apply_to_contacts.up.sql`
2. 在管理界面中，為需要自動套用的通知管道啟用 "自動套用" 選項
3. 當系統自動創建新的監控規則時，這些通知管道將自動與規則關聯

這樣，當新的監控對象被添加並自動應用規則時，系統也會自動套用通知管道，實現完整的自動化監控流程。

## 注意事項

1. 測試時請確保告警服務已正確配置並運行
2. 這些測試 payload 中的時間戳是固定的，實際使用時可能需要更新為當前時間
3. 測試後可以檢查資料庫中的 `rule_states` 表，確認 `last_check_value` 是否已更新
4. 自動套用通知管道功能需要資料庫遷移支持，請確保已執行相關遷移腳本
