package api

import (
	"net/http"
	"shared-lib/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Create AlertMute
// @Tags AlertMute
// @Accept json
// @Produce json
// @Param mute body models.AlertMute true "mute"
// @Success 200 {object} models.AlertMute
// @Router /alert-mute/create [post]
func (h *AlertHandler) CreateAlertMute(c *gin.Context) {
	var mute models.AlertMute
	if err := c.ShouldBindJSON(&mute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 檢查時間範圍
	start := time.Unix(int64(mute.StartTime), 0)
	end := time.Unix(int64(mute.EndTime), 0)
	now := time.Now()
	if now.After(start) && now.Before(end) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time range exceeded"})
		return
	}

	// 檢查名稱
	if ok, msg := h.MuteService.CheckName(mute); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// 設置 realm
	mute.RealmName = c.MustGet("user").(models.SSOUser).Realm

	// 創建
	if err := h.MuteService.Create(&mute); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mute)
}

// @Summary Get All Alert Mute
// @Tags AlertMute
// @Produce json
// @Success 200 {array} models.AlertMute
// @Router /alert-mute/get-all [get]
func (h *AlertHandler) GetAllAlertMute(c *gin.Context) {
	realm := c.MustGet("user").(models.SSOUser).Realm
	mutes, err := h.MuteService.List(realm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mutes)
}

// @Summary Update Alert Mute
// @Tags AlertMute
// @Accept json
// @Produce json
// @Param mute body models.AlertMute true "mute"
// @Success 200 {object} models.AlertMute
// @Router /alert-mute/update [put]
func (h *AlertHandler) UpdateAlertMute(c *gin.Context) {
	var mute models.AlertMute
	if err := c.ShouldBindJSON(&mute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 檢查名稱
	if ok, msg := h.MuteService.CheckName(mute); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// 設置 realm
	mute.RealmName = c.MustGet("user").(models.SSOUser).Realm

	// 更新
	if err := h.MuteService.Update(&mute); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mute)
}

// @Summary Delete Alert Mute
// @Tags AlertMute
// @Param id path int true "mute id"
// @Success 200 {object} models.Response
// @Router /alert-mute/delete/{id} [delete]
func (h *AlertHandler) DeleteAlertMuteByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.MuteService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
