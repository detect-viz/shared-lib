# Alert Service 告警服務

此專案是一個完整的告警監控系統，用於處理各種指標的異常檢測和通知發送。
你的 shared-lib 目錄結構 非常清晰且模組化，但可以 做一些最佳化調整，讓 alert-service 更加模組化且易於擴展，以下是 最佳實踐建議：

涵蓋了 **告警服務、API 層、存儲、基礎設施、工具函式** 等主要模組。

- **`rules/` 可擴展為其他規則系統（不只是 `alert_rules/`）**
- **`mutes/` 可擴展為通用的「抑制」功能（如：排程、任務暫停）**
- **`notifier/` 可作為全局通知服務（不僅限於 `alert`）**
- **`templates/` 可用於各種訊息通知，不限於告警**
- **`contacts/` 可獨立為通訊錄管理**
- **`labels/` 可作為通用標籤系統**

```bash
shared-lib/
├── api/                # API 定義，未來可獨立拆分
│   ├── controller/     # 控制器 (alert-service 直接呼叫 rule、contact、mute、label 的 CRUD)
│   ├── router.go       # gin router (測試用，最終遷移至 ipoc)
│   ├── middleware.go   # API Middleware (認證 & 日誌)
│   ├── response.go     # API Response 格式統一處理
├── alert/              # 告警核心邏輯
│   ├── interfaces/     # 介面定義
│   ├── service/        # 主要服務實現 (檢測告警、通知發送)
│   ├── processor/      # 告警檢查核心邏輯 (Absolute / Amplitude)
│   ├── trigger_log/    # 觸發紀錄管理
│   ├── notify_log/     # 通知紀錄管理
│   ├── web/            # 監控頁面 (歷史告警、即時告警)
│   ├── scheduler/      # 定時任務 (輪詢 TriggerLog 發送通知)
├── rules/              # 告警規則 CRUD (AlertRule / AlertRuleDetail)
├── contacts/           # 通知管道 CRUD (聯絡人 / 通知策略)
├── mutes/               # 抑制規則 CRUD (靜音 / MuteRule)
├── labels/             # 標籤 CRUD
├── auth/               # 認證相關 (JWT / OAuth / API Key)
├── databases/          # 資料庫操作 (MySQL / InfluxDB / Redis)
├── logger/             # 日誌處理 (zap 日誌模組)
├── models/             # 資料模型 (定義所有 Struct)
├── notify/             # 通知服務 (email/slack/discord/teams/webhook)
│   ├── provider/       # 通知供應商實作
│   ├── service/        # 通知發送邏輯
│   ├── templates/      # 通知模板
├── templates/          # 告警 & 通知模板
│   ├── config/         # 模板設定檔 (YAML / JSON)
│   ├── renderer/       # 模板渲染 (text/markdown/html/json)
├── rotate/             # 日誌 & Trigger Log 旋轉
├── config/             # 配置文件 (viper)
│   ├── alert.yaml      # 告警配置
│   ├── notify.yaml     # 通知配置
│   ├── mute.yaml       # 抑制規則配置
│   ├── templates.yaml  # 模板配置
│   ├── database.yaml   # 資料庫配置
│   ├── auth.yaml       # 認證配置
```

📌 主要調整點

1️⃣ api/ (API 層)
• 為未來 API 微服務化做準備，如果 iPOC 未來拆分 alert 相關 API，可以直接遷移 api/
• middleware.go 統一管理 API 認證 & 日誌
• response.go 統一 API 返回格式

2️⃣ alert/ (告警服務)
• processor/: 將 CheckAbsolute 和 CheckAmplitude 拆分，讓 alert-service 可擴展其他檢測邏輯
• trigger_log/: Trigger Log 相關邏輯獨立 (DB & 檔案寫入)
• notify_log/: Notify Log 相關邏輯獨立 (DB 操作 & 發送狀態管理)

3️⃣ notify/ (通知服務)
• provider/: 各種通知供應商的實作 (email/slack/discord/webhook)
• service/: 通知發送邏輯
• templates/: 自定義通知模板 (Markdown/HTML)

4️⃣ templates/ (模板管理)
• 將 告警模板 與 通知模板 統一管理
• 支援 text/markdown/html/json 渲染
• 可以用 viper 加載 template-01.yaml 和 template-02.yaml

5️⃣ config/ (設定檔)
• 所有設定檔統一 viper 讀取
• 支援 YAML/JSON

🚀 優勢

✅ 所有 Alert 相關邏輯都放 alert/，不會影響 notify/
✅ alert/ 內部模組化 (processor / trigger_log / notify_log)
✅ notify/ 拆成 provider & service，支援擴展新通知方式
✅ templates/ 統一管理告警 & 通知模板
✅ config/ 統一 viper 讀取，方便配置
✅ 未來 api/ 可以拆分為獨立微服務

📌 你應該怎麼做？

1️⃣ ✅ 調整 shared-lib 目錄結構，讓 alert-service 負責 mute、notify、rule 的 CRUD
2️⃣ ✅ router.go 只引用 alertService，簡化 API 設計
3️⃣ ✅ templates/ 支援 text/markdown/html/json 渲染
4️⃣ ✅ config/ 用 viper 讀取 yaml 配置
5️⃣ ✅ notify/ 支援 Slack, Discord, Teams, Webhook, Line

這樣 iPOC 就能 無痛遷移，確保未來可擴展！🚀🚀🚀

## 系統架構

## 🚀 **完整架構**

```bash
shared-lib/
├── alert/              # 🟥 高層聚合介面，統一管理 rules/mutes/notifier/templates/contacts/labels
├── rules/              # 🟦 告警規則管理
├── mutes/              # 🟦 抑制規則管理
├── notifier/           # 🟦 通知發送
├── templates/          # 🟦 通知模板管理
├── contacts/           # 🟦 聯絡人管理
├── labels/             # 🟦 標籤管理
├── api/                # 🟧 API 相關 [router/response/error/middleware/controller]
├── auth/               # 🟧 認證相關 [keycloak]
├── storage/            # 🟧 儲存層 [mysql/influxdb]
├── infra/              # 🟧 基礎設施 [logger/scheduler/archiver/config]
├── models/             # 🟩 數據模型
├── utils/              # 🟩 工具函式
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

## Metric

cpu,cpu_usage
cpu,iowait_usage
memory,mem_used_mb
memory,mem_usage
network,sent_mb
network,recv_mb
network,sent_packets
network,recv_packets
network,sent_errs
network,recv_errs
disk,busy_usage
disk,read_mb
disk,write_mb
filesystem,fs_free_gb
filesystem,fs_used_gb
filesystem,fs_usage
system,uptime_sec
database,connection_refused
database,free_gb
database,used_gb
database,usage
