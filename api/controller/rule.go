package controller

import (
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取規則列表
// @Description 取得所有規則
// @Tags Rule
// @Accept json
// @Produce json
// @Success 200 {array} models.Rule "成功回應"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule [get]
func (a *AlertAPI) ListRules(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	rules, err := a.ruleService.List(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, rules)
}

// @Summary 獲取單一規則
// @Description 根據 ID 獲取特定的規則
// @Tags Rule
// @Accept json
// @Produce json
// @Param id path int true "規則 ID"
// @Success 200 {object} models.Rule "成功回應"
// @Failure 400 {object} response.ErrorResponse "無效的 ID"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/{id} [get]
func (a *AlertAPI) GetRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	rule, err := a.ruleService.Get(id)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, rule)
}

// @Summary 創建規則
// @Description 新增一條規則
// @Tags Rule
// @Accept json
// @Produce json
// @Param rule body models.Rule true "規則內容"
// @Success 201 {object} models.Rule "成功創建"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule [post]
func (a *AlertAPI) CreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.ruleService.Create(&rule)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONCreated(c, rule)
}

// @Summary 更新規則
// @Description 根據 ID 更新規則內容
// @Tags Rule
// @Accept json
// @Produce json
// @Param id path int true "規則 ID"
// @Param rule body models.Rule true "規則內容"
// @Success 200 {object} models.Rule "成功更新"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/{id} [put]
func (a *AlertAPI) UpdateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.ruleService.Update(&rule)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, rule)
}

// @Summary 刪除規則
// @Description 根據 ID 刪除規則
// @Tags Rule
// @Accept json
// @Produce json
// @Param id path int true "規則 ID"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.ErrorResponse "無效的 ID"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/{id} [delete]
func (a *AlertAPI) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	err = a.ruleService.Delete(id)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, gin.H{"message": "刪除成功"})
}
