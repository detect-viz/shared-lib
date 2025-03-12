package keycloak

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/detect-viz/shared-lib/models"
)

// KeycloakClient 定義 Keycloak 客戶端介面
type KeycloakClient interface {
	GetKeycloakConfig() models.KeycloakConfig
	GetJWT() *gocloak.JWT
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

	// 新增的方法
	CreateUser(ctx context.Context, user gocloak.User) (string, error)
	CreateGroup(ctx context.Context, group gocloak.Group) (string, error)
	CreateRole(ctx context.Context, role gocloak.Role) error
	AddUserToRole(ctx context.Context, userID string, roleName string) error
	AddUserToGroup(ctx context.Context, userID string, groupID string) error

	// User CRUD
	UpdateUser(ctx context.Context, userID string, user gocloak.User) error
	DeleteUser(ctx context.Context, userID string) error

	// Group CRUD
	UpdateGroup(ctx context.Context, groupID string, group gocloak.Group) error
	DeleteGroup(ctx context.Context, groupID string) error

	// Role CRUD
	UpdateRole(ctx context.Context, roleName string, role gocloak.Role) error
	DeleteRole(ctx context.Context, roleName string) error

	// 關聯操作
	RemoveUserFromRole(ctx context.Context, userID string, roleName string) error
	RemoveUserFromGroup(ctx context.Context, userID string, groupID string) error

	// 獲取 Realm 的屬性
	GetRealmAttribute(ctx context.Context, attrKey string) (string, error)
}
