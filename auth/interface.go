package auth

import (
	"context"
)

// SSOClient 定義 SSO 客戶端介面
type SSOClient interface {
	GetAdminAcecessToken(ctx context.Context) (string, error)
}
