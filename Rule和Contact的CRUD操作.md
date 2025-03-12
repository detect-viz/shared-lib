# `Rule` 和 `Contact` 的 CRUD 操作

✅ `Rule` 和 `Contact` 的 CRUD 操作後，都應該觸發 `LoadGlobalRules()` 以確保即時更新監控規則！\*\*

---

## **🔹 為什麼 CRUD 需要 `LoadGlobalRules()`？**

### **1️⃣ `Rule` 變更**

- 新增 (`CreateRule()`)：新規則加入後，系統要即時開始監控
- 更新 (`UpdateRule()`)：如 `threshold`, `duration`, `silence_period` 變更，系統要重新套用
- 刪除 (`DeleteRule()`)：避免已刪除的 `Rule` 繼續影響 `AlertService`

📌 **影響範圍**
✅ **更新 `rules` 表**  
✅ **同步 `rule_states` 表**  
✅ **影響 `ProcessAlert()` 檢查邏輯**

---

### **2️⃣ `Contact` 變更**

- 新增 (`CreateContact()`)：新的聯絡人可接收通知
- 更新 (`UpdateContact()`)：變更 `severity`、`retry_delay` 等屬性
- 刪除 (`DeleteContact()`)：刪除後系統應該自動忽略對應的通知

📌 **影響範圍**
✅ **影響 `notify_logs`**（變更 `contact_id` 後的影響）  
✅ **影響 `rule_contacts` 關聯**（如果 `contact_id` 被移除，則規則可能失效）  
✅ **影響 `ProcessNotifyLog()`**（可能需要重新整理可通知對象）

---

## **🔹 優化 `LoadGlobalRules()`**

### **1️⃣ `Rule` 變更 → 影響 `rule_states`**

```go
func ReloadRules() {
    rules := db.LoadAllRules()
    ruleStates := db.LoadRuleStates()

    for _, rule := range rules {
        if _, exists := ruleStates[rule.ID]; !exists {
            db.CreateRuleState(rule)  // 確保 rule_states 存在
        }
    }
}
```

---

### **2️⃣ `Contact` 變更 → 影響 `rule_contacts`**

```go
func ReloadContacts() {
    contacts := db.LoadAllContacts()
    ruleContacts := db.LoadRuleContacts()

    for _, contact := range contacts {
        if contact.Deleted {  // 若聯絡人被刪除
            db.RemoveRuleContact(contact.ID)  // 移除所有關聯
        }
    }
}
```

---

## **🔹 總結**

✅ **`Rule` & `Contact` CRUD 變更後，應該觸發 `LoadGlobalRules()`**  
✅ **`ReloadRules()` 確保 `Rule` 變更後狀態即時更新**  
✅ **`ReloadContacts()` 確保 `Contact` 變更後不影響通知發送**  
✅ **`AlertService` & `NotificationService` 都應該監聽變更，確保數據一致**

## 當 `Contact` 被 `disabled` 或 `deleted`，需要檢查以下內容，以確保通知機制不受影響：

---

## **🔹 `Contact.Disabled = true` (禁用聯絡人)**

📌 **影響範圍**

- **影響 `rule_contacts` 關聯**
  - 如果某 `Rule` 仍綁定該 `Contact`，應該跳過該聯絡人
- **影響 `ProcessNotifyLog()`**
  - 發送通知時應該排除 `disabled` 的聯絡人
- **影響 `retryFailedNotifications()`**
  - 若 `Contact` 被禁用，應該停止 `retry`

📌 **處理方式** 1️⃣ `ProcessNotifyLog()` 跳過 `disabled` 聯絡人：

```go
func ProcessNotifyLog() {
    triggeredLogs := db.FindPendingTriggeredLogs()
    for _, log := range triggeredLogs {
        contacts := db.GetContactsForRule(log.RuleID)

        for _, contact := range contacts {
            if contact.Disabled {
                continue  // 忽略已禁用的聯絡人
            }
            SendNotification(contact, log)
        }
    }
}
```

2️⃣ `retryFailedNotifications()` 應該排除 `disabled` 聯絡人：

```go
func retryFailedNotifications() {
    failedLogs := db.FindFailedNotifications()
    for _, log := range failedLogs {
        contact := db.GetContactByID(log.ContactID)
        if contact.Disabled {
            log.State = "final_failed"  // 停止重試
            db.UpdateNotifyLog(log)
            continue
        }
        SendNotification(contact, log)
    }
}
```

---

## **🔹 `Contact.Deleted = true` (刪除聯絡人)**

📌 **影響範圍**

- **影響 `rule_contacts`**
  - `rule_contacts` 應該刪除該 `ContactID` 相關的記錄

📌 **處理方式** 1️⃣ 刪除 `Contact` 時，自動移除 `rule_contacts`：

```go
func DeleteContact(contactID string) {
    db.RemoveRuleContact(contactID)  // 刪除關聯
    db.MarkContactDeleted(contactID) // 標記已刪除
}
```

---

## **🔹 檢查重點**

✅ **`Contact.Disabled`**

- **影響 `ProcessNotifyLog()`，應該跳過通知**
- **影響 `retryFailedNotifications()`，應該停止重試**

✅ **`Contact.Deleted`**

- **影響 `rule_contacts`，應該刪除關聯**

- **影響 `ProcessNotifyLog()`，不應該再發送**

🚀 **現在可以測試 `Contact` 被 `Disabled` / `Deleted` 後，是否能正確影響 `Notify` 流程！**

## Response:

### **🔹 當 `Contact` 被 `Disabled` 或 `Deleted`，是否允許？是否應該警告或拒絕？**

根據不同場景，應該區分 **「允許但警告 ⚠️」** 或 **「直接拒絕 ❌」** 來處理。

---

## **1️⃣ `Contact.Disabled`（允許，但警告 ⚠️）**

✅ **允許的情境**

- `Contact` 被 **短暫停用**（例如：維護、測試、不想接收通知）
- `Rule` 仍然綁定該 `Contact`，但 `Contact` 被 `Disabled`
- `ProcessNotifyLog()` 執行時，應該跳過 `disabled` 聯絡人

⚠️ **應該警告的情境**

- 用戶嘗試手動觸發測試通知 (`Test Notification`)，但 `Contact` 被禁用
- `Rule` **唯一的聯絡人** 是 `disabled`，這將導致該 `Rule` 觸發後無法發送通知

📌 **警告方式** 1️⃣ 在 **`Test Notification`** API 回應 **Warning** 2️⃣ 在 **`Rule` 設定頁面** 顯示 **「該規則綁定的聯絡人已禁用」** 3️⃣ `ProcessNotifyLog()` 記錄 **Warning Log**

```go
if contact.Disabled {
    log.Warn("聯絡人 %s 被禁用，跳過通知", contact.ID)
    continue
}
```

---

## **2️⃣ `Contact.Deleted`（拒絕 ❌）**

❌ **應該拒絕的情境**

- 用戶試圖手動將已刪除的 `Contact` 加回 `Rule`
- `ProcessNotifyLog()` 仍然嘗試發送給已刪除的 `Contact`
- `Rule` **唯一的 `Contact` 已刪除**，這導致該 `Rule` 沒有通知對象

📌 **拒絕方式** 1️⃣ **刪除 `Contact` 時，應該同步刪除 `rule_contacts` 關聯** 2️⃣ **API 回應 `400 Bad Request`**

```go
if contact.Deleted {
    return errors.New("該聯絡人已刪除，無法使用")
}
```

3️⃣ **UI 上 `Rule` 設定應該過濾掉已刪除的 `Contact`**

---

## **🔹 總結**

| 狀況                 | 允許 | 警告 ⚠️                            | 拒絕 ❌ |
| -------------------- | ---- | ---------------------------------- | ------- |
| **Contact Disabled** | ✅   | **若是唯一 `Contact`，則警告**     | ❌      |
| **Contact Deleted**  | ❌   | **不能手動添加已刪除的 `Contact`** | ✅      |

📌 **`Disabled` 狀態：允許，但警告** 📌 **`Deleted` 狀態：應該拒絕，並清除關聯**  
📌 **如果 `Rule` 只綁定 `Disabled` 或 `Deleted` 的 `Contact`，應該警告用戶**

# **🔹 當 `Contact` 被刪除時，應該做什麼？**

✅ **影響 `RuleContact`** → 刪除 `RuleContact` 關聯，確保未來不會發送通知  
✅ **影響 `ProcessNotifyLog()`** → 跳過已刪除的 `Contact`  
✅ **影響 `retryFailedNotifications()`** → 停止對已刪除的 `Contact` 進行 `retry`

📌 **刪除 `Contact` 時**

```go
func DeleteContact(contactID []byte) {
    db.RemoveRuleContact(contactID) // 移除 Rule 關聯
}
```

📌 **`ProcessNotifyLog()` 排除已刪除 `Contact`**

```go
func ProcessNotifyLog() {
    logs := db.FindPendingNotifyLogs()
    for _, log := range logs {
        contact := db.GetContactByID(log.ContactID)
        if contact == nil { // Contact 被刪除
            continue // 不影響歷史 NotifyLog，但不會再發送
        }
        SendNotification(log)
    }
}
```

📌 **`RetryFailedNotifications()` 停止對刪除聯絡人的重試**

```go
func RetryFailedNotifications() {
    failedLogs := db.FindFailedNotifyLogs()
    for _, log := range failedLogs {
        contact := db.GetContactByID(log.ContactID)
        if contact == nil { // Contact 被刪除
            log.NotifyState = "final_failed" // 終止重試
            db.UpdateNotifyLog(log)
            continue
        }
        SendNotification(log)
    }
}
```

---

## **🔹 總結**

✅ **當 `Contact` 被刪除後，新的 `NotifyLog` 不會產生，但舊的 `NotifyLog` 仍然可以查詢**  
✅ **`ProcessNotifyLog()` 應該跳過已刪除 `Contact`，避免無效發送**  
✅ **`RetryFailedNotifications()` 應該終止對已刪除 `Contact` 的重試**

🚀 **這樣的設計確保 `NotifyLog` 仍能保留發送歷史，並且 `Contact` 被刪除後不會影響通知機制！**
