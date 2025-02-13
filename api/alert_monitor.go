package api

import (
	"net/http"

	"shared-lib/models"

	"github.com/gin-gonic/gin"
)

// @Summary  當前告警
// @Tags     Alert
// @Accept   json
// @Produce  json
// @Success  200 {object} []models.CurrentAlert
// @Security ApiKeyAuth
// @Router   /get-current-alert [get]
func (h *AlertHandler) GetCurrentAlert(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	r, err := services.GetCurrentAlert(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, r)
}

// @Summary  歷史告警
// @Tags     Alert
// @Accept   json
// @Produce  json
// @Success  200 {object} []models.HistoryAlert
// @Security ApiKeyAuth
// @Router   /get-history-alert [get]
func (h *AlertHandler) GetHistoryAlert(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	r, err := h.AlertService.GetHistoryAlert(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, r)
}

// @Summary  歷史告警圖表
// @Tags     Alert
// @Accept   json
// @Produce  json
// @Param    alert body models.HistoryAlert true "history_alert"
// @Success  200 {object} []models.MetricResponse
// @Security ApiKeyAuth
// @Router   /get-history-alert-metric [post]
func (h *AlertHandler) GetHistoryAlertMetric(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	body := new(models.HistoryAlert)
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	r, err := services.GetHistoryAlertMetric(user, *body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, r)
}
