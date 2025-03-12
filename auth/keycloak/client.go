package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	"github.com/detect-viz/shared-lib/models"
	"github.com/google/wire"
)

var KeycloakSet = wire.NewSet(NewClient)

// Client 實現 KeycloakClient 介面
type Client struct {
	gocloak        *gocloak.GoCloak
	keycloakConfig *models.KeycloakConfig
	jwt            *gocloak.JWT
}

// 確保 Client 實現了 KeycloakClient 介面
var _ KeycloakClient = (*Client)(nil)

// NewClient 創建新的 Keycloak 客戶端
func NewClient(keycloakConfig *models.KeycloakConfig) (KeycloakClient, error) {
	insecureSkipVerify := true
	if keycloakConfig.URL == "" || keycloakConfig.Realm == "" || keycloakConfig.ClientID == "" {
		return nil, fmt.Errorf("url, realm and clientID are required")
	}

	gc := gocloak.NewClient(keycloakConfig.URL)
	if insecureSkipVerify {
		gc.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	client := &Client{
		gocloak:        gc,
		keycloakConfig: keycloakConfig,
	}

	if keycloakConfig.User != "" && keycloakConfig.Password != "" {
		jwt, err := gc.LoginAdmin(context.Background(), keycloakConfig.User, keycloakConfig.Password, keycloakConfig.Realm)
		if err != nil {
			return nil, fmt.Errorf("login admin failed: %v", err)
		}
		client.jwt = jwt
	} else {
		// 初始化時進行服務帳戶登入
		jwt, err := gc.LoginClient(context.Background(), keycloakConfig.ClientID, keycloakConfig.ClientSecret, keycloakConfig.Realm)
		if err != nil {
			return nil, fmt.Errorf("login client failed: %v", err)
		}
		client.jwt = jwt
	}

	// 獲取客戶端UUID
	clients, err := gc.GetClients(context.Background(), client.jwt.AccessToken, keycloakConfig.Realm, gocloak.GetClientsParams{
		ClientID: &keycloakConfig.ClientID,
	})
	if err != nil {
		return nil, fmt.Errorf("get client ID failed: %v", err)
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("client not found")
	}
	client.keycloakConfig.ClientUUID = *clients[0].ID

	return client, nil
}

// LoginClient 服務帳戶登入
func (c *Client) LoginClient(ctx context.Context) (*gocloak.JWT, error) {
	return c.gocloak.LoginClient(ctx, c.keycloakConfig.ClientID, c.keycloakConfig.ClientSecret, c.keycloakConfig.Realm)
}

// GetClientIDOfClient 獲取客戶端ID
func (c *Client) GetClientIDOfClient(ctx context.Context) (string, error) {
	jwt, err := c.LoginClient(ctx)
	if err != nil {
		return "", err
	}

	clients, err := c.gocloak.GetClients(ctx, jwt.AccessToken, c.keycloakConfig.Realm, gocloak.GetClientsParams{
		ClientID: &c.keycloakConfig.ClientID,
	})
	if err != nil {
		return "", err
	}
	if len(clients) == 0 {
		return "", fmt.Errorf("client not found")
	}

	return *clients[0].ID, nil
}

func (c *Client) GetKeycloakConfig() models.KeycloakConfig {
	return *c.keycloakConfig
}

func (c *Client) GetJWT() *gocloak.JWT {
	return c.jwt
}

func (c *Client) GetGoCloak() *gocloak.GoCloak {
	return c.gocloak
}
