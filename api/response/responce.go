package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 通用 API 回應結構
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONResponse - 返回 JSON 格式的 API 回應
func JSONResponse(c *gin.Context, status int, data interface{}, message string) {
	resp := Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
	c.JSON(status, resp)
}

// JSONSuccess - 成功回應
func JSONSuccess(c *gin.Context, data interface{}) {
	JSONResponse(c, http.StatusOK, data, "success")
}

// JSONCreated - 創建成功
func JSONCreated(c *gin.Context, data interface{}) {
	JSONResponse(c, http.StatusCreated, data, "created")
}

// JSONError - 錯誤回應
func JSONError(c *gin.Context, status int, err error) {
	JSONResponse(c, status, nil, err.Error())
}
