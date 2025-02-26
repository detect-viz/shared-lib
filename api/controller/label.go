package controller

import (
	"encoding/csv"
	"strconv"

	"github.com/detect-viz/shared-lib/api/response"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// @Summary 獲取標籤列表
// @Description 取得所有標籤，支援分頁
// @Tags Label
// @Accept json
// @Produce json
// @Param limit query int false "限制返回的數量 (預設 100)"
// @Param offset query int false "偏移量 (預設 0)"
// @Success 200 {array} models.Label "成功回應"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label [get]
func (a *AlertAPI) ListLabels(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	labels, err := a.labelService.List(user.Realm, limit, offset)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, labels)
}

// @Summary 獲取單一標籤
// @Description 根據 Key 獲取特定的標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param key path string true "標籤 Key"
// @Success 200 {object} models.Label "成功回應"
// @Failure 400 {object} response.ErrorResponse "無效的 Key"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/{key} [get]
func (a *AlertAPI) GetLabel(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	key := c.Param("key")

	if key == "" {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	label, err := a.labelService.Get(user.Realm, key)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONSuccess(c, label)
}

// @Summary 創建標籤
// @Description 新增標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param label body models.Label true "標籤內容"
// @Success 201 {object} models.Label "成功創建"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label [post]
func (a *AlertAPI) CreateLabel(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)

	var label models.Label
	if err := c.ShouldBindJSON(&label); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	label.RealmName = user.Realm // 設定標籤所屬的 realm
	createdLabel, err := a.labelService.Create(user.Realm, &label)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	response.JSONCreated(c, createdLabel)
}

// @Summary 更新標籤
// @Description 更新標籤內容，支援局部更新
// @Tags Label
// @Accept json
// @Produce json
// @Param key path string true "標籤 Key"
// @Param label body object true "標籤內容 (只需提供要更新的欄位)"
// @Success 200 {object} models.Label "成功更新"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/{key} [patch]
func (a *AlertAPI) UpdateLabel(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	key := c.Param("key")

	if key == "" {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.JSONError(c, 400, response.ErrInvalidPayload)
		return
	}

	err := a.labelService.Update(user.Realm, key, updates)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	label, _ := a.labelService.Get(user.Realm, key)
	response.JSONSuccess(c, label)
}

// @Summary 刪除標籤
// @Description 根據 Key 刪除標籤
// @Tags Label
// @Accept json
// @Produce json
// @Param key path string true "標籤 Key"
// @Success 200 {object} map[string]string "刪除成功"
// @Failure 400 {object} response.ErrorResponse "無效的 Key"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/{key} [delete]
func (a *AlertAPI) DeleteLabel(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	key := c.Param("key")

	if key == "" {
		response.JSONError(c, 400, response.ErrInvalidID)
		return
	}

	err := a.labelService.Delete(user.Realm, key)
	if err != nil {
		response.JSONError(c, 500, err)
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
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
// @Security ApiKeyAuth
// @Router /alert/label/export [get]
func (a *AlertAPI) ExportCSV(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)

	// 取得標籤列表
	labels, err := a.labelService.List(user.Realm, 1000, 0) // 預設最多 1000 筆
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	// 設定 HTTP Header
	c.Header("Content-Disposition", "attachment; filename=labels.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 寫入 CSV 標題
	writer.Write([]string{"Key", "Value"})

	// 寫入標籤資料
	for _, label := range labels {
		writer.Write([]string{
			label.KeyName,
			label.Value.String(), // JSON 格式轉 string

		})
	}
}

// @Summary 上傳 CSV 匯入標籤
// @Description 批量新增或更新標籤
// @Tags Label
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上傳 CSV 檔案"
// @Success 200 {object} map[string]string "成功匯入"
// @Failure 400 {object} response.ErrorResponse "請求內容無效"
// @Failure 500 {object} response.ErrorResponse "伺服器錯誤"
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

	// 開啟檔案
	src, err := file.Open()
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}
	defer src.Close()

	// 讀取 CSV 內容
	reader := csv.NewReader(src)
	rows, err := reader.ReadAll()
	if err != nil {
		response.JSONError(c, 400, err)
		return
	}

	// 解析標籤
	var labels []models.Label
	for i, row := range rows {
		if i == 0 {
			continue // 跳過標題列
		}

		label := models.Label{
			RealmName: user.Realm,
			KeyName:   row[0],
			Value:     datatypes.JSON(row[1]),
		}
		labels = append(labels, label)
	}

	// 批量寫入 DB
	_, err = a.labelService.BulkCreateOrUpdate(user.Realm, labels)
	if err != nil {
		response.JSONError(c, 500, err)
		return
	}

	response.JSONSuccess(c, gin.H{"message": "CSV 匯入成功"})
}
