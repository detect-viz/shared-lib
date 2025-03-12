package controller

import (
	"github.com/detect-viz/shared-lib/alert"
	"github.com/detect-viz/shared-lib/api/middleware"
	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AlertAPI struct {
	alertService   *alert.Service
	ruleService    rules.Service
	contactService contacts.Service
	notifyService  notifier.Service
	logger         *zap.Logger
}

func NewAlertAPI(alertService *alert.Service) *AlertAPI {
	return &AlertAPI{
		alertService:   alertService,
		ruleService:    alertService.GetRuleService(),
		contactService: alertService.GetContactService(),
		notifyService:  alertService.GetNotifyService(),
		logger:         alertService.GetLogger(),
	}
}

// RegisterV1Routes 註冊 v1 版本 API
func RegisterV1Routes(router *gin.Engine, alertService *alert.Service, keycloak keycloak.Client) {
	v1 := router.Group("/api/v1/alert")

	// 初始化 API Controller
	alertAPI := NewAlertAPI(alertService)
	v1.Use(middleware.GetUserInfo(keycloak, alertService, &gin.Context{}))

	{
		v1.POST("/run-alert", alertAPI.ProcessAlert)
		v1.POST("/run-notify", alertAPI.ProcessNotifyLog)
	}
	// 註冊告警相關 API
	{
		v1.GET("/state", alertAPI.ListRuleState)
		v1.GET("/history", alertAPI.ListAlertHistory)
		v1.GET("/rule/metric-rule/:uid", alertAPI.GetMetricRule)
		v1.GET("/rule/metric-rule-options/:category", alertAPI.GetMetricRuleOptions)
		v1.GET("/rule/metric-rule-category-options", alertAPI.GetMetricRuleCategoryOptions)
	}

	// 註冊告警相關 API
	ruleRoutes := v1.Group("/rule")
	{
		ruleRoutes.GET("", alertAPI.ListRules)
		ruleRoutes.GET("/:id", alertAPI.GetRule)
		ruleRoutes.POST("", alertAPI.CreateRule)
		ruleRoutes.PUT("/:id", alertAPI.UpdateRule)
		ruleRoutes.DELETE("/:id", alertAPI.DeleteRule)
		ruleRoutes.POST("/manual-notify", alertAPI.ManualNotify)
	}

	// 註冊聯絡人 API
	contactRoutes := v1.Group("/contact")
	{
		contactRoutes.GET("", alertAPI.ListContacts)
		contactRoutes.GET("/:id", alertAPI.GetContact)
		contactRoutes.POST("", alertAPI.CreateContact)
		contactRoutes.PUT("/:id", alertAPI.UpdateContact)
		contactRoutes.DELETE("/:id", alertAPI.DeleteContact)
		contactRoutes.POST("/test", alertAPI.TestContact)
		contactRoutes.GET("/notify-methods", alertAPI.GetNotifyMethods)
		contactRoutes.GET("/notify-options", alertAPI.GetNotifyOptions)
	}
}
