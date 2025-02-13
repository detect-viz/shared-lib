package api

import (
	"ipoc-golang-api/models"

	"ipoc-golang-api/services/alerts"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary  取得所有群組每個主機的Partition
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param    metric_rule_id query int true "metric_rule_id"
// @Param    resource_group_id query int true "resource_group_id"
// @Success  200 {object} []models.ResourcePartition
// @Security ApiKeyAuth
// @Router   /get-resource-partition [get]
func (h *AlertHandler) GetResourcePartitionsByMetricRuleID(c *gin.Context) {

	user := c.Keys["user"].(models.SSOUser)

	metric_rule_id, _ := strconv.Atoi(c.Query("metric_rule_id"))
	resource_group_id, _ := strconv.Atoi(c.Query("resource_group_id"))

	res := alerts.GetResourcePartitionsByMetricRule(user.OrgID, metric_rule_id, resource_group_id)
	c.JSON(http.StatusOK, res)
}

// @Summary  取得 資源分區 By resource_name + metric_rule_id
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param    name path string true "resource_name"
// @Param    metric_rule_id query int true "metric_rule_id"
// @Success  200 {object} models.OptionRes
// @Security ApiKeyAuth
// @Router   /get-resource-partition/:name [get]
func (h *AlertHandler) GetPartitionByResourceName(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)

	metric_rule_id, err := strconv.Atoi(c.Query("metric_rule_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	resources := alerts.GetPartitionsByMetricRuleAndResource(user.OrgID, metric_rule_id, c.Param("name"))

	res := services.List2PullDownRes(resources)

	c.JSON(http.StatusOK, res)
}

// @Summary  取得 Alert 下拉選單
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param    metric query string true "metric_rule_category,alert_rule,alert_contact,alert_contact_type,resource_group"
// @Success  200 {object} models.OptionRes
// @Security ApiKeyAuth
// @Router   /get-alert-option [get]
func (h *AlertHandler) GetAlertOption(c *gin.Context) {

	realm := c.Keys["user"].(models.SSOUser).Realm
	//realm := "master"

	res := alerts.GetAlertOption(realm, c.Query("metric"))
	c.JSON(http.StatusOK, res)
}

// @Summary  取得告警單位下拉選單
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param id path int true "metric_id"
// @Success  200 {object} models.OptionRes
// @Router   /get-alert-unit/{id} [get]
func (h *AlertHandler) GetAlertUnitByMetricRuleID(c *gin.Context) {

	//id, err := strconv.Atoi(c.Param("id"))
	// if err != nil {
	// 	c.JSON(http.StatusBadGateway, err.Error())
	// 	return
	// }
	//res := alerts.GetAlertUnitByMetricRuleID(id)
	c.JSON(http.StatusOK, "")
}

// @Summary  取得告警規則的運算符號下拉選單
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param id path int true "metric_id"
// @Success  200 string string
// @Router   /get-alert-operator/{id} [get]
func (h *AlertHandler) GetAlertOperatorByMetricRuleID(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	res := alerts.GetAlertOperatorByMetricRuleID(id)
	c.JSON(http.StatusOK, res)
}
