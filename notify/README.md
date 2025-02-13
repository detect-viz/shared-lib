# Notify 模組

通知模組負責處理告警通知的發送、重試和日誌管理。

## 功能特點

### 1. 通知發送
- 支持多種通知渠道
  - Email
  - Teams
  - Line
  - Slack
  - Discord
  - Webex
  - Webhook

### 2. 生命週期管理
- 通知狀態追蹤
  - `notify_path`: 待發送通知
  - `send_path`: 發送成功
  - `failed_path`: 發送失敗
- 日誌命名規則
  ```
  {timestamp}_{resource}_{rule}_{uuid}_{status}_notify.log
  ```
  - `status`: alerting/resolved

### 3. 錯誤處理
- 智能重試機制
  - 最大重試次數: 3次
  - 指數退避策略
  - 初始等待: 1秒
  - 最大等待: 30秒

### 4. 日誌管理
- 自動壓縮
  - 每小時執行
  - 壓縮1小時前的日誌
  - 格式: tar.gz
- 自動清理
  - 每天凌晨3點執行
  - 保留30天
  - 單目錄最大500MB
  - 保持1GB剩餘空間

## 目錄結構
```
notify/
├── constants/     # 常量定義
├── errors/        # 錯誤類型
├── log/           # 日誌管理
├── channels/     # 通知實現
│   ├── email.go
│   ├── teams.go
│   ├── line.go
│   ├── slack.go
│   ├── discord.go
│   ├── webex.go
│   └── webhook.go
├── template/      # 通知模板
├── utils/         # 工具函數
└── service.go     # 核心服務
```

## 處理流程

1. 接收告警
   - 從 alert 模組接收 `NotificationLog`
   - 根據 Type 和 Status 選擇模板

2. 發送通知
   - 選擇對應的通知器
   - 渲染通知模板
   - 執行發送
   ```go
   channel, err := NewSender(notification)
   if err != nil {
       return err
   }
   return channel.Send(message)
   ```

3. 結果處理
   - 發送成功
     - 移動到 send_path
     - 更新 MySQL 記錄
   - 發送失敗
     - 重試次數 +1
     - 超過最大重試次數移至 failed_path
     - 記錄錯誤日誌

4. 日誌管理
   - 自動壓縮歸檔
   - 定期清理過期日誌
   - 監控磁碟空間

## 使用範例

```go
// 創建通知服務
service := notify.NewService(db)

// 發送通知
notification := models.NotificationLog{
    Type:   "email",
    Status: "alerting",
    Subject: "CPU 使用率過高",
    Body:    "Server CPU 使用率達到 90%",
}

if err := service.Send(notification); err != nil {
    logger.Error("發送通知失敗", zap.Error(err))
}
```

## 注意事項

1. 確保通知配置正確
2. 監控重試次數和失敗率
3. 定期檢查日誌狀態
4. 注意磁碟空間使用