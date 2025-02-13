package api

import (
	"ipoc-golang-api/global"
	"shared-lib/databases"
	"shared-lib/interfaces"
	"shared-lib/notify"
	"time"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	service *notify.Service
}

func NewContactHandler(db interfaces.Database) *ContactHandler {
	return &ContactHandler{
		service: notify.NewService(db),
	}
}

// NotifyAPI 通知相關 API
type NotifyAPI struct {
	service *notify.Service
}

// NewNotifyAPI 創建通知 API
func NewNotifyAPI() *NotifyAPI {
	databases := databases.NewMySQL(global.Mysql)
	return &NotifyAPI{
		service: notify.NewService(databases),
	}
}

// RegisterRoutes 註冊路由
func (api *NotifyAPI) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/api/notify")
	{
		g.POST("/process", api.ProcessNotifications) // 手動觸發處理通知
	}
}

// ProcessNotifications 處理通知
func (api *NotifyAPI) ProcessNotifications(c *gin.Context) {
	if err := api.service.ProcessNotifications(); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "處理完成",
		"time":    time.Now().Format(time.RFC3339),
	})
}
