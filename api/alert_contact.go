package api

import (
	"net/http"
	"shared-lib/alert"
	"shared-lib/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Get All Alert Contact
// @Description only admin
// @Tags AlertContact
// @Accept  json
// @Produce  json
// @Success 200 {object} []models.AlertContactRes
// @Security ApiKeyAuth
// @Router /alert-contact/get-all [get]
func (h *AlertHandler) GetAllAlertContact(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	contacts, err := h.ContactService.List(user.Realm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

// @Summary Test AlertContact
// @Tags AlertContact
// @Accept  json
// @Produce  json
// @Param contact body models.AlertContactRes true "contact"
// @Success 200 {object} models.Response
// @Security ApiKeyAuth
// @Router /alert-contact/test [post]
func (h *AlertHandler) AlertContactTest(c *gin.Context) {
	var contact models.AlertContact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	contact.RealmName = c.Keys["user"].(models.SSOUser).Realm
	res := h.ContactService.NotifyTest(contact)
	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Create AlertContact
// @Tags AlertContact
// @Accept  json
// @Produce  json
// @Param contact body models.AlertContactRes true "contact"
// @Success 200 {object} models.AlertContactRes
// @Security ApiKeyAuth
// @Router /alert-contact/create [post]
func (h *AlertHandler) CreateAlertContact(c *gin.Context) {
	var contact models.AlertContact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	contact.RealmName = c.Keys["user"].(models.SSOUser).Realm

	// 檢查名稱
	check, err := h.ContactService.CheckName(contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}
	if !check.Success {
		c.JSON(http.StatusBadRequest, check)
		return
	}

	// 創建通知管道
	if err := h.ContactService.Create(&contact); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	// 處理告警等級
	for i, lvl := range contact.Severities {
		contact.Levels[i] = alert.GetSeverityByName(lvl.Severity)
	}
	c.JSON(http.StatusOK, contact)
}

// @Summary Update Alert Contact
// @Description only admin
// @Tags AlertContact
// @Accept  json
// @Produce  json
// @Param group body models.AlertContactRes true "update"
// @Success 200 {object} models.AlertContactRes
// @Security ApiKeyAuth
// @Router /alert-contact/update [put]
func (h *AlertHandler) UpdateAlertContact(c *gin.Context) {
	var contact models.AlertContact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	contact.RealmName = c.Keys["user"].(models.SSOUser).Realm

	// 檢查名稱
	check, err := h.ContactService.CheckName(contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}
	if !check.Success {
		c.JSON(http.StatusBadRequest, check)
		return
	}

	// 更新通知管道
	if err := h.ContactService.Update(&contact); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	// 處理告警等級
	for i, lvl := range contact.Severities {
		contact.Levels[i] = alert.GetSeverityByName(lvl.Severity)
	}

	c.JSON(http.StatusOK, contact)
}

// @Summary Delete Alert Contact By ID
// @Description only admin
// @Tags AlertContact
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Param confirm query bool false "Confirm deletion"
// @Success 200 {object} models.Response
// @Security ApiKeyAuth
// @Router /alert-contact/delete/{id} [delete]
func (h *AlertHandler) DeleteAlertContactByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Msg:     "無效的 ID",
		})
		return
	}

	confirm := c.Query("confirm")

	// 檢查是否被使用
	isUsed, err := h.ContactService.IsUsedByRules(int(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	if isUsed && confirm != "true" {
		c.JSON(http.StatusConflict, models.Response{
			Success: false,
			Msg:     "該通知管道正在使用，確認是否繼續刪除？",
		})
		return
	}

	// 執行刪除
	userID := c.Keys["user"].(models.SSOUser).ID
	res := h.ContactService.Delete(id, userID)
	if !res.Success {
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Msg:     "通知管道刪除成功",
	})
}
