package api

import (
	"ipoc-golang-api/global"
	"shared-lib/alert"
	"shared-lib/alert/contact"
	"shared-lib/alert/mute"
	"shared-lib/databases"
	"shared-lib/label"

	"github.com/gin-gonic/gin"
)

// AlertHandler 處理所有告警相關的 API
type AlertHandler struct {
	AlertService   *alert.Service
	MuteService    *mute.Service
	LabelService   *label.Service
	ContactService *contact.Service
}

// NewAlertHandler 創建 AlertHandler
func NewAlertHandler(alertSvc *alert.Service, muteSvc *mute.Service, labelSvc *label.Service) *AlertHandler {
	return &AlertHandler{
		AlertService: alertSvc,
		MuteService:  muteSvc,
		LabelService: labelSvc,
	}
}

// AlertAPI 告警相關 API
type AlertAPI struct {
	service *alert.Service
	mute    *mute.Service
	handler *AlertHandler
	label   *label.Service
}

// NewAlertAPI 創建告警 API
func NewAlertAPI() *AlertAPI {
	databases := databases.NewMySQL(global.Mysql)
	alertService := alert.NewService(databases)
	muteService := mute.NewService(global.Mysql, global.Logger, global.Crontab)
	labelService := label.NewService(global.Mysql)
	handler := NewAlertHandler(alertService, muteService, labelService)

	return &AlertAPI{
		service: alertService,
		mute:    muteService,
		handler: handler,
		label:   labelService,
	}
}

// RegisterRoutes 註冊路由
func (api *AlertAPI) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/api/v1/alert")

	{
		//*** label ***//
		g.POST("alert-label/create", api.handler.Create)
		g.GET("alert-label/get-all", api.handler.List)
		g.GET("alert-label/:key", api.handler.GetLabelValueByKey)
		g.PUT("alert-label/update/:key", api.handler.Update)
		g.DELETE("/alert-label/delete/:key", api.handler.Delete)

		//*** alert select ***//
		g.GET("/get-alert-unit/:id", api.handler.GetAlertUnitByMetricRuleID)
		g.GET("/get-alert-operator/:id", api.handler.GetAlertOperatorByMetricRuleID)
		g.GET("/get-resource-partition", api.handler.GetResourcePartitionsByMetricRuleID)
		g.GET("/get-resource-partition/:name", api.handler.GetPartitionByResourceName)
		g.GET("/get-alert-option", api.handler.GetAlertOption)

		//*** metric rule ***//
		g.GET("/get-metric-rule/:category", api.handler.GetMetricRuleByCategory)
		g.GET("/get-metric-rule-detail/:id", api.handler.GetMetricRuleByID)
		g.GET("/get-metric-rule-keys/:id", api.handler.GetKeysByMetricRuleID)

		//*** alert dashboard ***//
		g.GET("/get-current-alert", api.handler.GetCurrentAlert)
		g.GET("/get-history-alert", api.handler.GetHistoryAlert)
		g.POST("/get-history-alert-metric", api.handler.GetHistoryAlertMetric)

		//*** alert mute ***//
		g.POST("/alert-mute/create", api.handler.CreateAlertMute)
		g.GET("/alert-mute/get-all", api.handler.GetAllAlertMute)
		g.PUT("/alert-mute/update", api.handler.UpdateAlertMute)
		g.DELETE("/alert-mute/delete/:id", api.handler.DeleteAlertMuteByID)

		//*** alert rule ***//
		g.POST("/alert-rule/create", api.handler.CreateAlertRule)
		g.GET("/alert-rule/get-all", api.handler.GetAllAlertRule)
		g.GET("/alert-rule/:id", api.handler.GetAlertRuleByID)
		g.PUT("/alert-rule/update", api.handler.UpdateAlertRule)
		g.DELETE("/alert-rule/delete/:id", api.handler.DeleteAlertRuleByID)
		g.GET("/copy-alert-rule/:id", api.handler.CopyAlertRuleByID)
		g.POST("/alert-rule/test", api.handler.TestAlertRule)

		//*** alert contact ***//
		g.POST("/alert-contact/test", api.handler.AlertContactTest)
		g.POST("/alert-contact/create", api.handler.CreateAlertContact)
		g.GET("/alert-contact/get-all", api.handler.GetAllAlertContact)
		g.PUT("/alert-contact/update", api.handler.UpdateAlertContact)
		g.DELETE("/alert-contact/delete/:id", api.handler.DeleteAlertContactByID)
	}
}
