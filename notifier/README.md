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

## 用戶資訊如何提供

如果用戶已經有 channel_token 或 smtp_host 相關資訊，則應該在 tenantConfig 內提供：

```go
tenantConfig := map[string]string{
    "smtp_host": "smtp.example.com",
    "smtp_port": "587",
    "smtp_user": "user@example.com",
    "smtp_pass": "supersecret",
    "channel_token": "your-line-token",
}

```
