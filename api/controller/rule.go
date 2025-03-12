package controller

import (
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary 獲取規則列表
// @Description 取得所有規則 [ID, Name, Type, Enabled]
// @Tags Rule
// @Accept json
// @Produce json
// @Param cursor query int false "created_at"
// @Param limit query int false "最大筆數"
// @Success 200 {object} response.Response "成功回應"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule [get]
func (a *AlertAPI) ListRules(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	cursor := int64(0) // 預設從 0 開始
	limit := 10        // 預設取 10 筆
	var err error

	if c.Query("cursor") != "" {
		cursor, err = strconv.ParseInt(c.Query("cursor"), 10, 64)
		if err != nil || cursor < 0 {
			response.JSONError(c, 400, apierrors.ErrInvalidID)
			return
		}
	}

	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			response.JSONError(c, 400, apierrors.ErrInvalidID)
			return
		}
	}

	rules, nextCursor, err := a.ruleService.List(user.Realm, cursor, limit)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONResponse(c, 200, gin.H{
		"rules":       rules,
		"next_cursor": nextCursor,
	}, "success")
}

// @Summary 獲取單一規則
// @Description 根據 ID 獲取特定的規則
// @Tags Rule
// @Accept json
// @Produce json
// @Param id path string true "規則 ID"
// @Success 200 {object} models.Rule "成功回應"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/{id} [get]
func (a *AlertAPI) GetRule(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	idStr := c.Param("id")
	if idStr == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	// 將 ID 轉換為 []byte
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	rule, err := a.ruleService.Get(user.Realm, string(id[:]), models.RuleOverview{})
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
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
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule [post]
func (a *AlertAPI) CreateRule(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	if user.Realm == "" {
		response.JSONError(c, 401, apierrors.ErrInvalidRealm)
		return
	}
	var rule models.RuleResponse
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	//* 創建規則
	newRule, err := a.ruleService.Create(user.Realm, &rule)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONCreated(c, newRule)
}

// @Summary 更新規則
// @Description 根據 ID 更新規則內容
// @Tags Rule
// @Accept json
// @Produce json
// @Param id path string true "規則 ID"
// @Param rule body models.Rule true "規則內容"
// @Success 200 {object} models.Rule "成功更新"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/{id} [put]
func (a *AlertAPI) UpdateRule(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	idStr := c.Param("id")
	if idStr == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	var rule models.RuleResponse
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	// 確保 rule.ID 來自 path
	rule.ID = idStr

	updatedRule, err := a.ruleService.Update(user.Realm, &rule)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}

	response.JSONSuccess(c, updatedRule)
}

// @Summary 刪除規則
// @Description 根據 ID 刪除規則
// @Tags Rule
// @Accept json
// @Produce json
// @Param id path string true "規則 ID"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/{id} [delete]
func (a *AlertAPI) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	// 將 ID 轉換為 []byte
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	err = a.ruleService.Delete(string(id[:]))
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}

	response.JSONSuccess(c, gin.H{"message": "刪除成功"})
}

// ManualNotifyRequest 手動通知請求
type ManualNotifyRequest struct {
	RuleIDs []string `json:"ruleIds" binding:"required"`
}

// ManualNotify 手動觸發通知
// @Summary 手動觸發通知
// @Description 手動觸發指定規則的通知
// @Tags Rule
// @Accept json
// @Produce json
// @Param request body ManualNotifyRequest true "手動通知請求"
// @Success 200 {object} map[string]string "通知已觸發"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/rule/manual-notify [post]
func (a *AlertAPI) ManualNotify(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	if user.Realm == "" {
		response.JSONError(c, 401, apierrors.ErrInvalidRealm)
		return
	}

	var req ManualNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	// 檢查請求參數
	if len(req.RuleIDs) == 0 {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	// 調用服務層手動觸發通知
	err := a.ruleService.ManualNotify(user.Realm, req.RuleIDs)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	response.JSONSuccess(c, gin.H{"message": "通知已觸發"})
}
