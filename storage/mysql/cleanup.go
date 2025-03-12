package mysql

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 執行 SQL 檔案
func (c *Client) executeSQLFile(filePath string) error {
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	sql := string(sqlBytes)
	result := c.db.Exec(sql)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Cleanup completed: %d rows affected", result.RowsAffected)
	return nil
}

// 定期清理
func (c *Client) StartExecuteSQLFile() {
	ticker := time.NewTicker(24 * time.Hour) // 每 24 小時執行一次
	defer ticker.Stop()

	for range ticker.C {
		err := c.executeSQLFile("/path/to/cleanup.sql") // 指定 SQL 檔案位置
		if err != nil {
			log.Println("Error executing cleanup SQL:", err)
		} else {
			log.Println("Cleanup job executed successfully.")
		}
	}
}

// 清理日誌
func (c *Client) cleanupLogs() {
	expiration := time.Now().AddDate(0, 0, -90).Unix() // 90 天前的時間戳
	result := c.db.Exec("DELETE FROM notify_logs WHERE created_at < ?", expiration)
	if result.Error != nil {
		log.Println("Failed to cleanup logs:", result.Error)
	} else {
		log.Println("Cleanup completed. Rows affected:", result.RowsAffected)
	}
}
func (c *Client) StartCleanupCron() {
	// 每天凌晨執行
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupLogs()
	}
}

// 定期自動增加分區  MySQL PARTITION BY RANGE（可快速刪除舊數據）
func (c *Client) checkAndAddPartition() {
	var latestPartition int64
	c.db.Raw("SELECT MAX(PARTITION_DESCRIPTION) FROM INFORMATION_SCHEMA.PARTITIONS WHERE TABLE_NAME = 'notify_logs'").Scan(&latestPartition)

	now := time.Now().Unix()
	nextPartition := latestPartition + 31536000 // 加一年
	if now > latestPartition-31536000 {         // 提前一年新增
		sql := fmt.Sprintf("ALTER TABLE notify_logs ADD PARTITION (PARTITION p%d VALUES LESS THAN (%d))", nextPartition, nextPartition)
		c.db.Exec(sql)
	}
}
