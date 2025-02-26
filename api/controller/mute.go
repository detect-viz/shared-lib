package controller

import (
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取抑制規則列表
// @Description 取得所有抑制規則
// @Tags Mute
// @Accept json
// @Produce json
// @Success 200 {array} models.Mute "成功回應"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/mute [get]
func (a *AlertAPI) ListMutes(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	mutes, err := a.muteService.List(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, mutes)
}

// @Summary 獲取單一抑制規則
// @Description 根據 ID 獲取特定的抑制規則
// @Tags Mute
// @Accept json
// @Produce json
// @Param id path int true "抑制規則 ID"
// @Success 200 {object} models.Mute "成功回應"
// @Failure 400 {object} response.ErrorResponse "無效的 ID"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/mute/{id} [get]
func (a *AlertAPI) GetMute(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	mute, err := a.muteService.Get(id)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, mute)
}

// @Summary 創建抑制規則
// @Description 新增一條抑制規則
// @Tags Mute
// @Accept json
// @Produce json
// @Param mute body models.Mute true "抑制規則內容"
// @Success 201 {object} models.Mute "成功創建"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/mute [post]
func (a *AlertAPI) CreateMute(c *gin.Context) {
	var mute models.Mute
	if err := c.ShouldBindJSON(&mute); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.muteService.Create(&mute)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONCreated(c, mute)
}

// @Summary 更新抑制規則
// @Description 根據 ID 更新抑制規則內容
// @Tags Mute
// @Accept json
// @Produce json
// @Param id path int true "抑制規則 ID"
// @Param mute body models.Mute true "抑制規則內容"
// @Success 200 {object} models.Mute "成功更新"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/mute/{id} [put]
func (a *AlertAPI) UpdateMute(c *gin.Context) {
	var mute models.Mute
	if err := c.ShouldBindJSON(&mute); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.muteService.Update(&mute)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, mute)
}

// @Summary 刪除抑制規則
// @Description 根據 ID 刪除抑制規則
// @Tags Mute
// @Accept json
// @Produce json
// @Param id path int true "抑制規則 ID"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.ErrorResponse "無效的 ID"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/mute/{id} [delete]
func (a *AlertAPI) DeleteMute(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	err = a.muteService.Delete(id)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, gin.H{"message": "刪除成功"})
}
