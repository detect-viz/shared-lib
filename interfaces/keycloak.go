package interfaces

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
)

// KeycloakClient 定義 Keycloak 客戶端介面
type KeycloakClient interface {
	// 認證相關
	GetAccessTokenByUsernamePassword(ctx context.Context, username, password string) (string, error)

	// 用戶信息
	GetUsers(ctx context.Context, accessToken string) ([]*gocloak.User, error)
	GetUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error)
	GetUserRoles(ctx context.Context, accessToken string) ([]string, error)
	GetUserGroups(ctx context.Context, accessToken string) ([]string, error)

	// 客戶端相關
	GetClientPermissions(ctx context.Context) error
	GetClientPolicies(ctx context.Context) error
	GetResourceByPermissionID(ctx context.Context, permissionID string) error
	GetClientPermissionByPolicyID(ctx context.Context, policyID string) error
	GetResources(ctx context.Context) error
	CreateResource(ctx context.Context, jwt *gocloak.JWT, clientID string) *gocloak.ResourceRepresentation
}
