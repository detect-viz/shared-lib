package keycloak

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

// GetUsers 獲取用戶列表
func (c *Client) GetUsers(ctx context.Context, accessToken string) ([]*gocloak.User, error) {
	users, err := c.gocloak.GetUsers(ctx, accessToken, c.realm, gocloak.GetUsersParams{})
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

	userInfo, err := c.gocloak.GetUserInfo(ctx, accessToken, c.realm)
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
	userInfo, err := c.gocloak.GetUserInfo(ctx, accessToken, c.realm)
	if err != nil {
		return nil, err
	}

	roles, err := c.gocloak.GetRealmRolesByUserID(ctx, accessToken, c.realm, *userInfo.Sub)
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
	userInfo, err := c.gocloak.GetUserInfo(ctx, accessToken, c.realm)
	if err != nil {
		return nil, err
	}

	groups, err := c.gocloak.GetUserGroups(ctx, accessToken, c.realm, *userInfo.Sub, gocloak.GetGroupsParams{})
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
