package keycloak

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
)

// GetGroupsByRealm 獲取 Realm 中的所有群組
func (c *Client) GetGroupsByRealm(ctx context.Context, realm string) ([]*gocloak.Group, error) {
	groups, err := c.gocloak.GetGroups(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, gocloak.GetGroupsParams{})
	if err != nil {
		return nil, err
	}
	return groups, nil
}
