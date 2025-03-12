package keycloak

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

// 獲取 Realm 中的所有群組
func (c *Client) GetGroupsByRealm(ctx context.Context, realm string) ([]*gocloak.Group, error) {
	groups, err := c.gocloak.GetGroups(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm, gocloak.GetGroupsParams{})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// 獲取 Realm 的屬性
func (c *Client) GetRealmAttribute(ctx context.Context, attrKey string) (string, error) {
	realmRep, err := c.gocloak.GetRealm(ctx, c.jwt.AccessToken, c.keycloakConfig.Realm)
	if err != nil {
		return "", err
	}

	// 確保 Attributes 存在
	if realmRep.Attributes == nil {
		return "", fmt.Errorf("no attributes found in realm")
	}

	// 先嘗試直接獲取 key
	if val, ok := (*realmRep.Attributes)[attrKey]; ok {
		return val, nil
	}

	// 檢查是否有 "acr.loa.map"，並解析 JSON
	if nestedStr, ok := (*realmRep.Attributes)["acr.loa.map"]; ok {
		var nestedMap map[string]string
		if err := json.Unmarshal([]byte(nestedStr), &nestedMap); err == nil {
			if nestedVal, ok := nestedMap[attrKey]; ok {
				return nestedVal, nil
			}
		}
	}

	// 先嘗試直接讀取 smtp_config
	smtpJSON, ok := (*realmRep.Attributes)["smtp_config"]
	if !ok {
		return "", fmt.Errorf("smtp_config attribute not found in realm")
	}

	// 解析 JSON
	var smtpConfig map[string]string
	if err := json.Unmarshal([]byte(smtpJSON), &smtpConfig); err != nil {
		return "", fmt.Errorf("failed to parse smtp_config: %w", err)
	}

	// 解碼 Base64 密碼（如果有的話）
	if encodedPass, exists := smtpConfig["password"]; exists {
		decodedPass, err := base64.StdEncoding.DecodeString(encodedPass)
		if err == nil {
			smtpConfig["password"] = string(decodedPass)
		}
	}

	return "", fmt.Errorf("attribute %s not found", attrKey)
}
