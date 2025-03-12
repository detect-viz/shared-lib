package controller

import (
	"encoding/csv"
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
)

// @Summary 獲取標籤列表
// @Description 取得所有標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param cursor query int false "created_at"
// @Param limit query int false "最大筆數"
// @Success 200 {object} response.Response "成功回應"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label [get]
func (a *AlertAPI) ListLabels(c *gin.Context) {
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

	Labels, nextCursor, err := a.labelService.List(user.Realm, cursor, limit)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONResponse(c, 200, gin.H{
		"Labels":      Labels,
		"next_cursor": nextCursor,
	}, "success")
}

// @Summary 獲取單一標籤
// @Description 根據 ID 獲取特定的標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param id path int true "標籤 ID"
// @Success 200 {object} models.LabelDTO "成功回應"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/{id} [get]
func (a *AlertAPI) GetLabel(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	Label, err := a.labelService.Get(id)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}
	response.JSONSuccess(c, Label)
}

// @Summary 創建標籤
// @Description 新增一條標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param Label body models.LabelDTO true "標籤內容"
// @Success 201 {object} models.LabelDTO "成功創建"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label [post]
func (a *AlertAPI) CreateLabel(c *gin.Context) {

	user := c.Keys["user"].(models.SSOUser)
	var input models.LabelDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}
	if input.Key == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidKey)
		return
	}
	newLabel, err := a.labelService.Create(user.Realm, &input)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONCreated(c, newLabel)
}

// @Summary 更新標籤
// @Description 根據 ID 更新標籤內容
// @Tags Label
// @Accept json
// @Produce json
// @Param id path int true "標籤 ID"
// @Param Label body models.LabelDTO true "標籤內容"
// @Success 200 {object} models.LabelDTO "成功更新"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/{id} [put]
func (a *AlertAPI) UpdateLabel(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}
	var label models.LabelDTO
	if err := c.ShouldBindJSON(&label); err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidPayload)
		return
	}

	// 確保 key 來自 path
	label.ID = id

	updatedLabel, err := a.labelService.Update(user.Realm, &label)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			response.JSONError(c, apiErr.Code, apiErr)
		} else {
			response.JSONError(c, 500, apierrors.ErrInternalError)
		}
		return
	}

	response.JSONSuccess(c, updatedLabel)
}

// @Summary 刪除標籤
// @Description 根據 ID 刪除標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param id path int true "標籤 ID"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.Response "無效的 ID"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/{id} [delete]
func (a *AlertAPI) DeleteLabel(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.JSONError(c, 400, apierrors.ErrInvalidID)
		return
	}

	err = a.labelService.Delete(id)
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

// @Summary 下載標籤 CSV
// @Description 匯出標籤列表為 CSV
// @Tags Label
// @Accept json
// @Produce text/csv
// @Success 200 {string} string "成功下載 CSV"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/export [post]
func (a *AlertAPI) ExportCSV(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	// 設定 Response Headers
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=labels.csv")
	c.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")

	// 創建 CSV Writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	labels, err := a.labelService.ExportCSV(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	for _, label := range labels {
		if err := writer.Write(label); err != nil {
			response.JSONError(c, 500, err)
			return
		}
	}
}

// @Summary 上傳 CSV 匯入標籤
// @Description 批量新增或更新標籤
// @Tags Label
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上傳 CSV 檔案"
// @Success 200 {object} map[string]string "成功匯入"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/import [post]
func (a *AlertAPI) ImportCSV(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)

	// 取得上傳的檔案
	file, err := c.FormFile("file")
	if err != nil {
		response.JSONError(c, 400, err)
		return
	}
	// 異步處理
	go a.labelService.ImportCSV(user.Realm, file)

	response.JSONSuccess(c, "CSV is being processed")
}

// @Summary 更新標籤 key
// @Description 更新標籤 key
// @Tags Label
// @Accept json
// @Produce json
// @Param key path int true "標籤 key"
// @Param Label body models.LabelDTO true "標籤內容"
// @Success 200 {object} models.LabelDTO "成功更新"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/key-options [get]
func (a *AlertAPI) GetKeyOptions(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	updatedLabel, err := a.labelService.GetKeyOptions(user.Realm)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, updatedLabel)
}

// @Summary 更新標籤 key
// @Description 更新標籤 key
// @Tags Label
// @Accept json
// @Produce json
// @Param name path string true "舊標籤 key"
// @Param new_name query string true "新標籤 key"
// @Success 200 {object} models.LabelDTO "成功更新"
// @Failure 400 {object} response.Response "請求內容無效"
// @Failure 500 {object} response.Response "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/key/{name} [put]
func (a *AlertAPI) UpdateLabelKeyName(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	oldName := c.Param("name")
	newName := c.Query("new_name")
	if newName == "" {
		response.JSONError(c, 400, apierrors.ErrInvalidKey)
		return
	}
	updatedLabel, err := a.labelService.UpdateKeyName(user.Realm, oldName, newName)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, updatedLabel)
}
