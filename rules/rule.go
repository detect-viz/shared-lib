package rules

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Controller 規則控制器
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController 創建規則控制器
func NewController(service Service, logger *zap.Logger) *Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// CreateRule 創建規則
// @Summary 創建規則
// @Description 創建一個新的規則
// @Tags rules
// @Accept json
// @Produce json
// @Param rule body models.RuleResponse true "規則信息"
// @Success 200 {object} models.RuleResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rules [post]
func (c *Controller) CreateRule(ctx *gin.Context) {
	realm := ctx.GetString("realm")
	if realm == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("realm 不能為空"))
		return
	}

	var ruleResp models.RuleResponse
	if err := ctx.ShouldBindJSON(&ruleResp); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}

	// 調用服務層創建規則
	createdRule, err := c.service.Create(realm, &ruleResp)
	if err != nil {
		c.logger.Error("創建規則失敗", zap.Error(err))
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	response.JSONSuccess(ctx, createdRule)
}

// GetRule 獲取規則
// @Summary 獲取規則
// @Description 根據 metricRuleUID 獲取規則
// @Tags rules
// @Accept json
// @Produce json
// @Param metricRuleUID path string true "指標規則 UID"
// @Success 200 {object} models.RuleResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rules/{metricRuleUID} [get]
func (c *Controller) GetRule(ctx *gin.Context) {
	realm := ctx.GetString("realm")
	if realm == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("realm 不能為空"))
		return
	}

	metricRuleUID := ctx.Param("metricRuleUID")
	if metricRuleUID == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("metricRuleUID 不能為空"))
		return
	}

	// 解析 ruleOverview 參數
	var ruleOverview models.RuleOverview
	overviewStr := ctx.Query("overview")
	if overviewStr != "" {
		if err := json.Unmarshal([]byte(overviewStr), &ruleOverview); err != nil {
			response.JSONError(ctx, http.StatusBadRequest, err)
			return
		}
	}

	// 調用服務層獲取規則
	rule, err := c.service.Get(realm, metricRuleUID, ruleOverview)
	if err != nil {
		c.logger.Error("獲取規則失敗", zap.Error(err))
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	if rule == nil {
		response.JSONError(ctx, http.StatusNotFound, errors.New("規則不存在"))
		return
	}

	response.JSONSuccess(ctx, rule)
}

// ListRules 列出規則
// @Summary 列出規則
// @Description 列出所有規則
// @Tags rules
// @Accept json
// @Produce json
// @Param cursor query int false "游標"
// @Param limit query int false "限制"
// @Success 200 {object} response.Response{data=[]models.MetricRuleOverview}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rules [get]
func (c *Controller) ListRules(ctx *gin.Context) {
	realm := ctx.GetString("realm")
	if realm == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("realm 不能為空"))
		return
	}

	// 解析分頁參數
	cursorStr := ctx.DefaultQuery("cursor", "0")
	limitStr := ctx.DefaultQuery("limit", "10")

	cursor, err := strconv.ParseInt(cursorStr, 10, 64)
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}

	// 調用服務層列出規則
	rules, nextCursor, err := c.service.List(realm, cursor, limit)
	if err != nil {
		c.logger.Error("列出規則失敗", zap.Error(err))
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	// 創建包含游標的響應
	result := gin.H{
		"data":       rules,
		"nextCursor": nextCursor,
	}
	response.JSONSuccess(ctx, result)
}

// UpdateRule 更新規則
// @Summary 更新規則
// @Description 更新現有規則
// @Tags rules
// @Accept json
// @Produce json
// @Param id path string true "規則 ID"
// @Param rule body models.RuleResponse true "規則信息"
// @Success 200 {object} models.RuleResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rules/{id} [put]
func (c *Controller) UpdateRule(ctx *gin.Context) {
	realm := ctx.GetString("realm")
	if realm == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("realm 不能為空"))
		return
	}

	id := ctx.Param("id")
	if id == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("id 不能為空"))
		return
	}

	var ruleResp models.RuleResponse
	if err := ctx.ShouldBindJSON(&ruleResp); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}

	// 設置 ID
	ruleResp.ID = id

	// 調用服務層更新規則
	updatedRule, err := c.service.Update(realm, &ruleResp)
	if err != nil {
		c.logger.Error("更新規則失敗", zap.Error(err))
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	response.JSONSuccess(ctx, updatedRule)
}

// DeleteRule 刪除規則
// @Summary 刪除規則
// @Description 刪除現有規則
// @Tags rules
// @Accept json
// @Produce json
// @Param id path string true "規則 ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rules/{id} [delete]
func (c *Controller) DeleteRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("id 不能為空"))
		return
	}

	// 調用服務層刪除規則
	err := c.service.Delete(id)
	if err != nil {
		c.logger.Error("刪除規則失敗", zap.Error(err))
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	response.JSONSuccess(ctx, nil)
}

// GetAvailableTargets 獲取可用的監控目標
// @Summary 獲取可用的監控目標
// @Description 獲取可用的監控目標
// @Tags rules
// @Accept json
// @Produce json
// @Param metricRuleUID path string true "指標規則 UID"
// @Success 200 {object} response.Response{data=[]models.Target}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/rules/{metricRuleUID}/targets [get]
func (c *Controller) GetAvailableTargets(ctx *gin.Context) {
	realm := ctx.GetString("realm")
	if realm == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("realm 不能為空"))
		return
	}

	metricRuleUID := ctx.Param("metricRuleUID")
	if metricRuleUID == "" {
		response.JSONError(ctx, http.StatusBadRequest, errors.New("metricRuleUID 不能為空"))
		return
	}

	// 調用服務層獲取可用目標
	targets, err := c.service.GetAvailableTarget(realm, metricRuleUID)
	if err != nil {
		c.logger.Error("獲取可用目標失敗", zap.Error(err))
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	response.JSONSuccess(ctx, targets)
}
