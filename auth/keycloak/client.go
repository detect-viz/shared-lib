package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/detect-viz/shared-lib/interfaces"

	"github.com/Nerzal/gocloak/v13"
)

// Client 實現 KeycloakClient 介面
type Client struct {
	gocloak      *gocloak.GoCloak
	realm        string
	clientID     string
	clientSecret string
	jwt          *gocloak.JWT
	clientUUID   string
}

// 確保 Client 實現了 KeycloakClient 介面
var _ interfaces.KeycloakClient = (*Client)(nil)

// NewClient 創建新的 Keycloak 客戶端
func NewClient(url, realm, clientID, clientSecret string, insecureSkipVerify bool) (interfaces.KeycloakClient, error) {
	if url == "" || realm == "" || clientID == "" {
		return nil, fmt.Errorf("url, realm and clientID are required")
	}

	gc := gocloak.NewClient(url)
	if insecureSkipVerify {
		gc.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	client := &Client{
		gocloak:      gc,
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	// 初始化時進行服務帳戶登入
	jwt, err := gc.LoginClient(context.Background(), clientID, clientSecret, realm)
	if err != nil {
		return nil, fmt.Errorf("login client failed: %v", err)
	}
	client.jwt = jwt

	// 獲取客戶端UUID
	clients, err := gc.GetClients(context.Background(), jwt.AccessToken, realm, gocloak.GetClientsParams{
		ClientID: &clientID,
	})
	if err != nil {
		return nil, fmt.Errorf("get client ID failed: %v", err)
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("client not found")
	}
	client.clientUUID = *clients[0].ID

	return client, nil
}

// LoginClient 服務帳戶登入
func (c *Client) LoginClient(ctx context.Context) (*gocloak.JWT, error) {
	return c.gocloak.LoginClient(ctx, c.clientID, c.clientSecret, c.realm)
}

// GetClientIDOfClient 獲取客戶端ID
func (c *Client) GetClientIDOfClient(ctx context.Context) (string, error) {
	jwt, err := c.LoginClient(ctx)
	if err != nil {
		return "", err
	}

	clients, err := c.gocloak.GetClients(ctx, jwt.AccessToken, c.realm, gocloak.GetClientsParams{
		ClientID: &c.clientID,
	})
	if err != nil {
		return "", err
	}
	if len(clients) == 0 {
		return "", fmt.Errorf("client not found")
	}

	return *clients[0].ID, nil
}
