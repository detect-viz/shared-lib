package response

import "errors"

var (
	ErrInvalidID      = errors.New("無效的 ID")
	ErrInvalidPayload = errors.New("無效的請求內容")
	ErrInternalError  = errors.New("伺服器錯誤")
)
