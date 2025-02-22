# Alert Service 告警服務

此專案是一個完整的告警監控系統，用於處理各種指標的異常檢測和通知發送。

## 系統架構

```
alert-service/
├── alert/              # 告警核心邏輯
│   ├── alert.go        # 主要服務實現
│   └── mute/           # 告警抑制邏輯
├── auth/               # 認證相關
│   └── keycloak/       # Keycloak 整合
├── databases/          # 資料庫操作
│   ├── mysql.go        # MySQL 實現
│   └── event_log.go    # 事件日誌處理
├── interfaces/         # 介面定義
│   ├── database.go     # 資料庫介面
│   ├── logger.go       # 日誌介面
│   ├── notify.go       # 通知介面
│   └── keycloak.go     # 認證介面
├── logger/             # 日誌處理
│   └── logger.go       # 日誌實現
├── models/             # 資料模型
│   ├── alert/          # 告警相關模型
│   ├── common/         # 共用模型
│   ├── config/         # 配置模型
│   ├── mute/           # 抑制規則模型
│   └── notify/         # 通知模型
├── notify/             # 通知服務
│   └── service.go      # 通知實現
└── conf/              # 配置文件
    └── migrations/    # 資料庫遷移
```

## 核心功能

### 1. 告警檢測 (Alert Detection)
- 支援多種檢測方式：
  - 絕對值檢測 (Absolute)
  - 振幅檢測 (Amplitude)
- 可配置的告警等級：Info、Warning、Critical
- 支援自定義閾值和持續時間

### 2. 告警抑制 (Alert Muting)
- 支援時間範圍抑制
- 支援週期性抑制（每日、每週、每月）
- 資源群組級別的抑制規則

### 3. 通知管理 (Notification)
- 多種通知管道：
  - Email
  - Teams
  - Webhook
- 通知重試機制
- 自定義通知模板
- 分級通知策略

### 4. 狀態追蹤 (State Tracking)
- TriggerLog：記錄告警觸發
- NotifyLog：記錄通知發送
- AlertState：追蹤告警狀態

## 資料流程

1. 指標接收
   - 接收來自各種來源的指標數據
   - 進行初步的數據驗證和格式化

2. 告警檢測
   - 載入相關的告警規則
   - 執行閾值檢查
   - 產生告警觸發記錄

3. 抑制處理
   - 檢查是否符合抑制規則
   - 設置抑制時間範圍
   - 更新告警狀態

4. 通知發送
   - 產生通知內容
   - 選擇適當的通知管道
   - 執行通知發送
   - 處理重試邏輯

## 配置說明

### 告警規則配置
```yaml
alert:
  rules:
    - name: "CPU 使用率過高"
      type: "absolute"
      threshold:
        info: 70
        warn: 80
        crit: 90
      duration: 5
```

### 通知管道配置
```yaml
notify:
  channels:
    - type: "email"
      config:
        host: "smtp.example.com"
        port: 587
    - type: "teams"
      config:
        webhook_url: "https://..."
```

## 開發指南

1. 新增告警類型
   - 實現 CheckRule 介面
   - 在 alert.go 中註冊新的檢查邏輯

2. 新增通知管道
   - 實現 NotifyService 介面
   - 在 notify/service.go 中添加新的發送邏輯

3. 資料庫操作
   - 所有 SQL 操作都應通過 Database 介面
   - 使用事務確保資料一致性

## 部署要求

- Go 1.16+
- MySQL 5.7+
- Keycloak (用於認證)
- 足夠的磁碟空間用於日誌存儲

## 監控指標

- 告警觸發率
- 通知成功率
- 規則處理時間
- 資料庫操作延遲

## 注意事項

1. 所有時間操作都應使用 Unix 時間戳
2. 錯誤訊息應使用中文
3. 關鍵操作都應該記錄日誌
4. 需要定期清理歷史數據
