package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取聯絡人列表
// @Description 取得所有聯絡人 [ID, Name, Type, Enabled]
// @Tags Contact
// @Accept json
// @Produce json
// @Param cursor query int false "created_at"
// @Param limit query int false "最大筆數"
// @Success 200 {object} response.Response "成功回應"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact [get]
func (a *AlertAPI) ListContacts(c *gin.Context) {
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

	contactResponses, nextCursor, err := a.contactService.List(user.Realm, cursor, limit)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	response.JSONResponse(c, 200, gin.H{
		"contacts":    contactResponses,
		"next_cursor": nextCursor,
	}, "success")
}

// @Summary 獲取單一聯絡人
// @Description 根據 ID 獲取特定的聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "聯絡人 ID"
// @Success 200 {object} models.ContactResponse "成功回應"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/{id} [get]
func (a *AlertAPI) GetContact(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	contactResponse, err := a.contactService.Get(idStr)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}

	response.JSONSuccess(c, contactResponse)
}

// @Summary 創建聯絡人
// @Description 新增一條聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Param contact body models.ContactResponse true "聯絡人內容"
// @Success 201 {object} models.ContactResponse "成功創建"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact [post]
func (a *AlertAPI) CreateContact(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	var contactResp models.ContactResponse
	if err := c.ShouldBindJSON(&contactResp); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	newContactResp, err := a.contactService.Create(user.Realm, &contactResp)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	response.JSONCreated(c, newContactResp)
}

// @Summary 更新聯絡人
// @Description 根據 ID 更新聯絡人內容
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "聯絡人 ID"
// @Param contact body models.ContactResponse true "聯絡人內容"
// @Success 200 {object} models.ContactResponse "成功更新"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/{id} [put]
func (a *AlertAPI) UpdateContact(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	idStr := c.Param("id")
	if idStr == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	var contactResp models.ContactResponse
	if err := c.ShouldBindJSON(&contactResp); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	// 確保 contactResp.ID 來自 path
	contactResp.ID = idStr

	updatedContactResp, err := a.contactService.Update(user.Realm, &contactResp)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}

	response.JSONSuccess(c, updatedContactResp)
}

// @Summary 刪除聯絡人
// @Description 根據 ID 刪除聯絡人
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "聯絡人 ID"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/{id} [delete]
func (a *AlertAPI) DeleteContact(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	err := a.contactService.Delete(idStr)
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

// @Summary 測試聯絡人
// @Description 測試聯絡人是否能夠收到通知
// @Tags Contact
// @Accept json
// @Produce json
// @Param contact body models.ContactResponse true "聯絡人內容"
// @Success 200 {object} models.Response "成功回應"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/test [post]
func (a *AlertAPI) TestContact(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	var contactResp models.ContactResponse
	if err := c.ShouldBindJSON(&contactResp); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	// 確保 Config 存在
	if contactResp.Config == nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	err := a.contactService.NotifyTest(user.Realm, &contactResp)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}

	response.JSONSuccess(c, gin.H{"message": "測試成功"})
}

// @Summary 獲取通知方法
// @Description 取得所有通知方法
// @Tags Contact
// @Accept json
// @Produce json
// @Success 200 {array} string "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/notify-methods [get]
func (a *AlertAPI) GetNotifyMethods(c *gin.Context) {
	methods := a.contactService.GetNotifyMethods()
	response.JSONSuccess(c, methods)
}

// @Summary 獲取通知選項
// @Description 取得所有通知選項
// @Tags Contact
// @Accept json
// @Produce json
// @Success 200 {array} string "成功回應"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/contact/notify-options [get]
func (a *AlertAPI) GetNotifyOptions(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	options, err := a.contactService.GetNotifyOptions(context.Background(), user.Realm)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, err)
		return
	}
	response.JSONSuccess(c, options)
}
