package api

import (
	"net/http"
	"shared-lib/models"

	"github.com/gin-gonic/gin"
)

// @Summary Create AlertLabel
// @Tags AlertLabel
// @Accept  json
// @Produce  json
// @Param label body []models.Label true "label"
// @Success 200 {object} models.Label
// @Security ApiKeyAuth
// @Router /alert-label/create [post]
func (h *AlertHandler) Create(c *gin.Context) {
	var labels []models.Label
	if err := c.ShouldBindJSON(&labels); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	realm := c.Keys["user"].(models.SSOUser).Realm
	if err := h.LabelService.CreateLabel(realm, labels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, labels)
}

// @Summary  取得 alert label 下拉選單 by key
// @Tags     AlertOption
// @Accept   json
// @Produce  json
// @Param key path string true "key"
// @Success  200 {object} models.OptionRes
// @Router   /get-alert-label/{key} [get]
func (h *AlertHandler) GetLabelValueByKey(c *gin.Context) {
	realm := c.Keys["user"].(models.SSOUser).Realm
	key := c.Param("key")

	res, err := h.LabelService.GetLabelsByKey(realm, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Get All Custom Label
// @Description only admin
// @Tags AlertLabel
// @Accept  json
// @Produce  json
// @Success 200 {object} []models.Label
// @Security ApiKeyAuth
// @Router /alert-label/get-all [get]
func (h *AlertHandler) List(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	res := h.LabelService.GetAllLabels(user.Realm)
	c.JSON(http.StatusOK, res)
}

// @Summary Update Custom Label
// @Description only admin
// @Tags AlertLabel
// @Accept  json
// @Produce  json
// @Param key path string true "key"
// @Param group body []models.Label true "update"
// @Success 200 {object} models.Label
// @Security ApiKeyAuth
// @Router /alert-label/update/{key} [put]
func (h *AlertHandler) Update(c *gin.Context) {

	body := new([]models.Label)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	realm := c.Keys["user"].(models.SSOUser).Realm

	update_labels := *body

	//* 檢查 label value 有沒有重複
	for _, label := range update_labels {
		label.RealmName = realm
		check := h.LabelService.CheckLabelKey(label)
		if !check.Success {
			c.JSON(http.StatusBadRequest, check.Msg)
			return
		}
	}

	// Update DB
	res, err := h.LabelService.UpdateLabel(realm, c.Param("key"), *body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Delete Custom Label By Key
// @Description only admin
// @Tags AlertLabel
// @Accept  json
// @Produce  json
// @Param key path string true "key"
// @Success 200 {object} models.Response
// @Security ApiKeyAuth
// @Router /alert-label/delete/{key} [delete]
func (h *AlertHandler) Delete(c *gin.Context) {

	// Delete DB
	res := h.LabelService.DeleteLabelByKey(c.Keys["user"].(models.SSOUser).Realm, c.Param("key"))
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Msg)
}
