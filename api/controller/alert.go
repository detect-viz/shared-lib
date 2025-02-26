package controller

import (
	"github.com/detect-viz/shared-lib/alert"
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/labels"
	"github.com/detect-viz/shared-lib/mutes"
	"github.com/detect-viz/shared-lib/rules"
	"github.com/gin-gonic/gin"
)

type AlertAPI struct {
	alertService   *alert.Service
	muteService    mutes.Service
	ruleService    rules.Service
	contactService contacts.Service
	labelService   labels.Service
}

func NewAlertAPI(alertService *alert.Service) *AlertAPI {
	return &AlertAPI{
		alertService:   alertService,
		muteService:    alertService.GetMuteService(),
		ruleService:    alertService.GetRuleService(),
		contactService: alertService.GetContactService(),
		labelService:   alertService.GetLabelService(),
	}
}

// RegisterV1Routes 註冊 v1 版本 API
func RegisterV1Routes(router *gin.Engine, alertService *alert.Service) {
	v1 := router.Group("/api/v1/alert")

	// 初始化 API Controller
	alertAPI := NewAlertAPI(alertService)

	// 註冊告警相關 API
	{
		v1.GET("/state", alertAPI.ListAlertState)
		v1.GET("/history", alertAPI.ListAlertHistory)
	}

	// 註冊告警相關 API
	ruleRoutes := v1.Group("/rule")
	{
		ruleRoutes.GET("", alertAPI.ListRules)
		ruleRoutes.GET("/:id", alertAPI.GetRule)
		ruleRoutes.POST("", alertAPI.CreateRule)
		ruleRoutes.PUT("/:id", alertAPI.UpdateRule)
		ruleRoutes.DELETE("/:id", alertAPI.DeleteRule)
	}

	// 註冊抑制規則 API (mute)
	muteRoutes := v1.Group("/mute")
	{
		muteRoutes.GET("", alertAPI.ListMutes)
		muteRoutes.GET("/:id", alertAPI.GetMute)
		muteRoutes.POST("", alertAPI.CreateMute)
		muteRoutes.PUT("/:id", alertAPI.UpdateMute)
		muteRoutes.DELETE("/:id", alertAPI.DeleteMute)
	}

	// 註冊聯絡人 API
	contactRoutes := v1.Group("/contact")
	{
		contactRoutes.GET("", alertAPI.ListContacts)
		contactRoutes.GET("/:id", alertAPI.GetContact)
		contactRoutes.POST("", alertAPI.CreateContact)
		contactRoutes.PUT("/:id", alertAPI.UpdateContact)
		contactRoutes.DELETE("/:id", alertAPI.DeleteContact)
	}

	// 註冊標籤 API
	labelRoutes := v1.Group("/label")
	{
		labelRoutes.GET("", alertAPI.ListLabels)
		labelRoutes.GET("/:key", alertAPI.GetLabel)
		labelRoutes.POST("", alertAPI.CreateLabel)
		labelRoutes.PUT("/:key", alertAPI.UpdateLabel)
		labelRoutes.DELETE("/:key", alertAPI.DeleteLabel)
		labelRoutes.GET("/export", alertAPI.ExportCSV)
		labelRoutes.POST("/import", alertAPI.ImportCSV)
	}
}
