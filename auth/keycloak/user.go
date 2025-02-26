package keycloak

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

// GetUsers 獲取用戶列表
func (c *Client) GetUsers(ctx context.Context, accessToken string) ([]*gocloak.User, error) {
	users, err := c.gocloak.GetUsers(ctx, accessToken, c.keycloakConfig.Realm, gocloak.GetUsersParams{})
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserInfo 獲取用戶信息
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	userInfo, err := c.gocloak.GetUserInfo(ctx, accessToken, c.keycloakConfig.Realm)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %v", err)
	}

	if userInfo == nil {
		return nil, fmt.Errorf("user info is nil")
	}

	// 將用戶信息轉換為 map
	result := make(map[string]interface{})

	// 安全地添加用戶信息
	if userInfo.Sub != nil {
		result["id"] = *userInfo.Sub
	}
	if userInfo.PreferredUsername != nil {
		result["username"] = *userInfo.PreferredUsername
	}
	if userInfo.Email != nil {
		result["email"] = *userInfo.Email
	} else {
		result["email"] = ""
	}

	roles, err := c.GetUserRoles(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	result["roles"] = roles

	groups, err := c.GetUserGroups(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	result["groups"] = groups

	return result, nil
}

// GetUserRoles 獲取用戶角色
func (c *Client) GetUserRoles(ctx context.Context, accessToken string) ([]string, error) {
	userInfo, err := c.gocloak.GetUserInfo(ctx, accessToken, c.keycloakConfig.Realm)
	if err != nil {
		return nil, err
	}

	roles, err := c.gocloak.GetRealmRolesByUserID(ctx, accessToken, c.keycloakConfig.Realm, *userInfo.Sub)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return []string{}, nil
	}

	var roleNames []string
	for _, role := range roles {
		if role.Name != nil {
			roleNames = append(roleNames, *role.Name)
		}
	}

	return roleNames, nil
}

// GetUserGroups 獲取用戶群組
func (c *Client) GetUserGroups(ctx context.Context, accessToken string) ([]string, error) {
	userInfo, err := c.gocloak.GetUserInfo(ctx, accessToken, c.keycloakConfig.Realm)
	if err != nil {
		return nil, err
	}

	groups, err := c.gocloak.GetUserGroups(ctx, accessToken, c.keycloakConfig.Realm, *userInfo.Sub, gocloak.GetGroupsParams{})
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []string{}, nil
	}

	var groupNames []string
	for _, group := range groups {
		if group.Name != nil {
			groupNames = append(groupNames, *group.Name)
		}
	}
	return groupNames, nil
}

// CreateUser 創建用戶
func (c *Client) CreateUser(ctx context.Context, user gocloak.User) (string, error) {
	userID, err := c.gocloak.CreateUser(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, user)
	if err != nil {
		return "", fmt.Errorf("create user failed: %w", err)
	}
	return userID, nil
}

// CreateGroup 創建群組
func (c *Client) CreateGroup(ctx context.Context, group gocloak.Group) (string, error) {
	groupID, err := c.gocloak.CreateGroup(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, group)
	if err != nil {
		return "", fmt.Errorf("create group failed: %w", err)
	}
	return groupID, nil
}

// CreateRole 創建角色
func (c *Client) CreateRole(ctx context.Context, role gocloak.Role) error {
	_, err := c.gocloak.CreateRealmRole(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, role)
	if err != nil {
		return fmt.Errorf("create role failed: %w", err)
	}
	return nil
}

// AddUserToRole 將用戶加入角色
func (c *Client) AddUserToRole(ctx context.Context, userID string, roleName string) error {
	role, err := c.gocloak.GetRealmRole(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, roleName)
	if err != nil {
		return fmt.Errorf("get role failed: %w", err)
	}

	err = c.gocloak.AddRealmRoleToUser(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, userID, []gocloak.Role{*role})
	if err != nil {
		return fmt.Errorf("add user to role failed: %w", err)
	}
	return nil
}

// AddUserToGroup 將用戶加入群組
func (c *Client) AddUserToGroup(ctx context.Context, userID string, groupID string) error {
	err := c.gocloak.AddUserToGroup(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, userID, groupID)
	if err != nil {
		return fmt.Errorf("add user to group failed: %w", err)
	}
	return nil
}

// UpdateUser 更新用戶
func (c *Client) UpdateUser(ctx context.Context, userID string, user gocloak.User) error {
	err := c.gocloak.UpdateUser(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, user)
	if err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	return nil
}

// DeleteUser 刪除用戶
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	err := c.gocloak.DeleteUser(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, userID)
	if err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}

// UpdateGroup 更新群組
func (c *Client) UpdateGroup(ctx context.Context, groupID string, group gocloak.Group) error {
	err := c.gocloak.UpdateGroup(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, group)
	if err != nil {
		return fmt.Errorf("update group failed: %w", err)
	}
	return nil
}

// DeleteGroup 刪除群組
func (c *Client) DeleteGroup(ctx context.Context, groupID string) error {
	err := c.gocloak.DeleteGroup(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, groupID)
	if err != nil {
		return fmt.Errorf("delete group failed: %w", err)
	}
	return nil
}

// UpdateRole 更新角色
func (c *Client) UpdateRole(ctx context.Context, roleName string, role gocloak.Role) error {
	err := c.gocloak.UpdateRealmRole(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, roleName, role)
	if err != nil {
		return fmt.Errorf("update role failed: %w", err)
	}
	return nil
}

// DeleteRole 刪除角色
func (c *Client) DeleteRole(ctx context.Context, roleName string) error {
	err := c.gocloak.DeleteRealmRole(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, roleName)
	if err != nil {
		return fmt.Errorf("delete role failed: %w", err)
	}
	return nil
}

// RemoveUserFromRole 將用戶從角色中移除
func (c *Client) RemoveUserFromRole(ctx context.Context, userID string, roleName string) error {
	role, err := c.gocloak.GetRealmRole(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, roleName)
	if err != nil {
		return fmt.Errorf("get role failed: %w", err)
	}

	err = c.gocloak.DeleteRealmRoleFromUser(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, userID, []gocloak.Role{*role})
	if err != nil {
		return fmt.Errorf("remove user from role failed: %w", err)
	}
	return nil
}

// RemoveUserFromGroup 將用戶從群組中移除
func (c *Client) RemoveUserFromGroup(ctx context.Context, userID string, groupID string) error {
	err := c.gocloak.DeleteUserFromGroup(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, userID, groupID)
	if err != nil {
		return fmt.Errorf("remove user from group failed: %w", err)
	}
	return nil
}
