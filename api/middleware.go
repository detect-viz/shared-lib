package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/detect-viz/shared-lib/alert"
	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Keycloak [Token 驗證]
func GetUserInfo(keycloakClient *keycloak.Client, alertService *alert.Service, c *gin.Context) {
	logger := alertService.GetLogger()
	sso := keycloakClient.GetKeycloakConfig()
	// 取 token 驗證
	tokens := c.Request.Header["Authorization"]
	if len(tokens) == 0 {
		c.JSON(http.StatusUnauthorized, "authorization token is required")
		return
	}
	token := tokens[0]

	// 取 realm 驗證
	realms := c.Request.Header["Realm"]
	if len(realms) == 0 {
		c.JSON(http.StatusNotFound, "realm is empty")
		return
	}
	realm := realms[0]

	// 設定參數
	ctx := context.Background()
	url := sso.URL
	client := gocloak.NewClient(url)
	restyClient := client.RestyClient()
	restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// 取 UserInfo
	userInfo, err := client.GetUserInfo(ctx, token, realm)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "GetUserInfo"+err.Error())
		return
	}

	user := models.SSOUser{}
	user.IsAdmin = false
	user.ID = *(userInfo.Sub)
	user.Name = *(userInfo.PreferredUsername)
	user.Realm = realm
	user.OrgName = realm

	jwt := keycloakClient.GetJWT()
	gocloakClient := keycloakClient.GetGoCloak()

	// 取 user roles
	roles, err := gocloakClient.GetRoleMappingByUserID(ctx, jwt.AccessToken, realm, *userInfo.Sub)
	if err != nil {
		logger.Error(
			err.Error(),
			zap.String("service", "keycloak"),
		)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var role_list, group_list []string

	// 提取角色名称
	for _, role := range *roles.RealmMappings {
		role_list = append(role_list, *role.Name)
	}

	// 取 user groups
	params := gocloak.GetGroupsParams{
		Full: gocloak.BoolP(true),
	}
	userGroups, err := client.GetUserGroups(ctx, jwt.AccessToken, realm, *userInfo.Sub, params)
	if err != nil {
		logger.Error(
			err.Error(),
			zap.String("service", "keycloak"),
		)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	for _, group := range userGroups {
		group_list = append(group_list, *(group.Name))
		user.RealmGroups = append(user.RealmGroups, *(group.Name))
	}

	//* 按 user roles 判斷有沒有 admin 權限
	user.Roles = role_list
	if slices.Contains(role_list, sso.AdminRole) {
		user.IsAdmin = true
	}

	fmt.Println("===================== 用戶資訊 START =====================")
	fmt.Printf("Realm: 	          %v\n", realm)
	fmt.Printf("OrgID:            %v\n", user.OrgID)
	fmt.Printf("OrgName: 	  %v\n", user.OrgName)
	fmt.Printf("UserID: 	  %v\n", user.ID)
	fmt.Printf("Name: 		  %v\n", user.Name)
	fmt.Printf("Roles: 		  %v\n", strings.Join(role_list, ", "))
	fmt.Printf("IsAdmin: 	  %v\n", user.IsAdmin)
	fmt.Printf("RealmGroups: 	  %v\n", strings.Join(group_list, ", "))
	fmt.Printf("AccessHosts: \n%v\n", strings.Join(user.AccessHosts, ", "))
	fmt.Println("===================== 用戶資訊 END =======================")
	c.Set("user", user)
	c.Next()
}
