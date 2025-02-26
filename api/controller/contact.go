package controller

import (
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取聯絡人列表
// @Description 取得所有聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Success 200 {array} models.Contact "成功回應"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact [get]
func (a *AlertAPI) ListContacts(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	contacts, err := a.contactService.List(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, contacts)
}

// @Summary 獲取單一聯絡人
// @Description 根據 ID 獲取特定的聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "聯絡人 ID"
// @Success 200 {object} models.Contact "成功回應"
// @Failure 400 {object} response.ErrorResponse "無效的 ID"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/{id} [get]
func (a *AlertAPI) GetContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	contact, err := a.contactService.Get(id)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, contact)
}

// @Summary 創建聯絡人
// @Description 新增一條聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Param contact body models.Contact true "聯絡人內容"
// @Success 201 {object} models.Contact "成功創建"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact [post]
func (a *AlertAPI) CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.contactService.Create(&contact)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONCreated(c, contact)
}

// @Summary 更新聯絡人
// @Description 根據 ID 更新聯絡人內容
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "聯絡人 ID"
// @Param contact body models.Contact true "聯絡人內容"
// @Success 200 {object} models.Contact "成功更新"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/{id} [put]
func (a *AlertAPI) UpdateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.contactService.Update(&contact)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, contact)
}

// @Summary 刪除聯絡人
// @Description 根據 ID 刪除聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "聯絡人 ID"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.ErrorResponse "無效的 ID"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/{id} [delete]
func (a *AlertAPI) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	err = a.contactService.Delete(id)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, gin.H{"message": "刪除成功"})
}
