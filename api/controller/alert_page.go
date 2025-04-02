package controller

import (
	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取告警狀態列表
// @Description 取得所有告警狀態
// @Tags Alert
// @Accept json
// @Produce json
// @Success 200 {array} models.RuleState "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/state [get]
func (a *AlertAPI) ListRuleState(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	ruleState, err := a.alertService.ListRuleState(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, ruleState)
}

// @Summary 獲取告警歷史列表
// @Description 取得所有告警歷史
// @Tags Alert
// @Accept json
// @Produce json
// @Success 200 {array} []models.TriggeredLog "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
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

// @Summary 獲取告警規則列表
// @Description 取得所有告警規則
// @Tags Alert
// @Accept json
// @Produce json
// @Param category path string true "規則類別"
// @Success 200 {array} models.Rule "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/metric-rule-options/{category} [get]
func (a *AlertAPI) GetMetricRuleOptions(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		err := apierrors.ErrInvalidID
		response.JSONError(c, 400, err)
		return
	}

	metricRuleOptions, err := a.alertService.GetMetricRuleOptions(category)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}
	response.JSONSuccess(c, metricRuleOptions)
}

// @Summary 獲取告警規則列表
// @Description 取得所有告警規則
// @Tags Alert
// @Accept json
// @Produce json
// @Success 200 {array} models.Rule "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/metric-rule-category-options [get]
func (a *AlertAPI) GetMetricRuleCategoryOptions(c *gin.Context) {

	metricRuleCategoryOptions, err := a.alertService.GetMetricRuleCategoryOptions()
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, metricRuleCategoryOptions)
}

// @Summary 獲取規則詳細
// @Description 取得規則詳細
// @Tags Alert
// @Accept json
// @Produce json
// @Param uid path string true "規則 UID"
// @Success 200 {array} models.Rule "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/metric-rule/{uid} [get]
func (a *AlertAPI) GetMetricRule(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		err := apierrors.ErrInvalidID
		response.JSONError(c, 400, err)
		return
	}
	metricRule, err := a.alertService.GetMetricRule(uid)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, metricRule)
}
