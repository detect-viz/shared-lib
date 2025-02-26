package controller

import (
	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取告警狀態列表
// @Description 取得所有告警狀態
// @Tags Alert
// @Accept json
// @Produce json
// @Success 200 {array} models.AlertState "成功回應"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/state [get]
func (a *AlertAPI) ListAlertState(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	alertState, err := a.alertService.ListAlertState(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, alertState)
}

// @Summary 獲取告警歷史列表
// @Description 取得所有告警歷史
// @Tags Alert
// @Accept json
// @Produce json
// @Success 200 {array} models.AlertHistory "成功回應"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/history [get]
func (a *AlertAPI) ListAlertHistory(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	alertHistory, err := a.alertService.ListAlertHistory(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, alertHistory)
}
