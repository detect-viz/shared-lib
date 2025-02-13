package api

import (
	"ipoc-golang-api/services/alerts"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Get Metric Rule By ID
// @Description only admin
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Success 200 {object} models.MetricRule
// @Security ApiKeyAuth
// @Router /get-metric-rule-detail/{id} [get]
func (h *AlertHandler) GetMetricRuleByID(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	res, err := alerts.GetMetricRuleByID(id)
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary  取得 MetricRule 所需參數名
// @Tags     MetricRule
// @Accept   json
// @Produce  json
// @Param id path int true "metric_id"
// @Success  200 {object} []string
// @Router   /get-metric-rule-keys/{id} [get]
func (h *AlertHandler) GetKeysByMetricRuleID(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	res := alerts.GetKeysByMetricRuleID(id)
	c.JSON(http.StatusOK, res)
}

// @Summary  取得 metric rule 下拉選單 by category
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param category path string true "category"
// @Success  200 {object} models.OptionRes
// @Router   /get-metric-rule/{category} [get]
func (h *AlertHandler) GetMetricRuleByCategory(c *gin.Context) {

	category := c.Param("category")

	res := alerts.GetMetricRuleByCategory(category)

	c.JSON(http.StatusOK, res)
}
