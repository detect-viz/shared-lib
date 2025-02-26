package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Nerzal/gocloak/v13"
)

// GetClientPermission 獲取用戶權限
func (c *Client) GetClientPermissions(ctx context.Context) error {

	permissions, err := c.gocloak.GetPermissions(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, c.keycloakConfig.ClientUUID, gocloak.GetPermissionParams{})
	if err != nil {
		return err
	}
	fmt.Println(permissions)
	return nil
}

// GetClientPermissionByPolicyID 獲取用戶權限
func (c *Client) GetClientPermissionByPolicyID(ctx context.Context, policyID string) error {

	permissions, err := c.gocloak.GetDependentPermissions(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, c.keycloakConfig.ClientUUID, policyID)
	if err != nil {
		return err
	}
	fmt.Println(permissions)
	return nil
}

// GetClientPolicies 獲取用戶權限
func (c *Client) GetClientPolicies(ctx context.Context) error {

	policies, err := c.gocloak.GetPolicies(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, c.keycloakConfig.ClientUUID, gocloak.GetPolicyParams{})
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(policies, "", "\t")
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
	//fmt.Println(policies)
	return nil
}

func (c *Client) CreateResource(ctx context.Context, jwt *gocloak.JWT, clientID string) *gocloak.ResourceRepresentation {
	c.gocloak.CreateResource(ctx, jwt.AccessToken, c.keycloakConfig.Realm, clientID, gocloak.ResourceRepresentation{
		ID: &clientID,
	})
	return &gocloak.ResourceRepresentation{
		ID: &clientID,
	}
}

func (c *Client) GetResources(ctx context.Context) error {
	resources, err := c.gocloak.GetResources(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, c.keycloakConfig.ClientUUID, gocloak.GetResourceParams{
		Owner: &c.keycloakConfig.ClientUUID,
	})
	if err != nil {
		return err
	}
	fmt.Println(resources)
	return nil
}

func (c *Client) GetResourceByPermissionID(ctx context.Context, permissionID string) error {

	resources, err := c.gocloak.GetPermissionResources(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, c.keycloakConfig.ClientUUID, permissionID)
	if err != nil {
		return err
	}
	fmt.Println(resources)
	return nil
}

// Login 使用用戶名密碼登入
func (c *Client) GetAccessTokenByUsernamePassword(ctx context.Context, username, password string) (string, error) {
	// 使用密碼方式登入
	token, err := c.gocloak.Login(ctx,
		c.keycloakConfig.ClientID,
		c.keycloakConfig.ClientSecret,
		c.keycloakConfig.Realm,
		username,
		password,
	)
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	return token.AccessToken, nil
}
