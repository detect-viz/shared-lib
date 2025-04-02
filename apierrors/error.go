package apierrors

import "errors"

// API 錯誤定義
type APIError struct {
	Code    int    // HTTP 狀態碼
	Message string // 錯誤訊息
	Err     error  // 內部錯誤
}

// NewAPIError 創建 API 錯誤
func NewAPIError(code int, msg string, err error) *APIError {
	return &APIError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// 實作 `error` 介面
func (e *APIError) Error() string {
	return e.Message
}

// 🔥 定義錯誤常數（提供 `Code`）
var (
	ErrInvalidRealm       = NewAPIError(401, "無效的 Realm", errors.New("invalid realm"))
	ErrInvalidAccessToken = NewAPIError(401, "無效的 Access Token", errors.New("invalid access token"))
	ErrInvalidID          = NewAPIError(400, "無效的 ID", errors.New("invalid id"))
	ErrInvalidKey         = NewAPIError(400, "無效的 Key", errors.New("invalid key"))
	ErrInvalidPayload     = NewAPIError(400, "請求內容無效", errors.New("invalid payload"))
	ErrNotFound           = NewAPIError(404, "資源不存在", errors.New("data not found"))
	ErrUsedByRules        = NewAPIError(409, "仍被使用中", errors.New("used by rules"))
	ErrDuplicateEntry     = NewAPIError(409, "名稱已存在", errors.New("data already exists"))
	ErrInternalError      = NewAPIError(500, "伺服器錯誤", errors.New("internal server error"))
	ErrRecordDeleted      = NewAPIError(404, "資源已軟刪除", errors.New("record deleted"))
)
