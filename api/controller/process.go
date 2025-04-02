package controller

import (
	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 立即通知
// @Description 批次通知
// @Tags Alert
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "成功回應"
// @Router /alert/run-notify [post]
func (a *AlertAPI) ProcessNotifyLog(c *gin.Context) {
	err := a.alertService.ProcessNotifyLog()
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, gin.H{"message": "通知處理成功"})
}

// @Summary 告警處理
// @Description 告警處理
// @Tags Alert
// @Accept json
// @Param alertPayload body models.AlertPayload true "檢測數據"
// @Success 200 {object} response.Response "成功回應"
// @Router /alert/run-alert [post]
func (a *AlertAPI) ProcessAlert(c *gin.Context) {
	// 實現告警處理邏輯
	var alertPayload models.AlertPayload
	if err := c.ShouldBindJSON(&alertPayload); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}
	err := a.alertService.ProcessAlert(alertPayload)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, gin.H{"message": "告警處理成功"})
}
