package api

import (
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/detect-viz/shared-lib/alert"
	"github.com/detect-viz/shared-lib/api/controller"
	"github.com/detect-viz/shared-lib/auth/keycloak"

	_ "github.com/detect-viz/shared-lib/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// 設定 API 路由
func RegisterRoutes(alertService *alert.Service, keycloak keycloak.Client) *gin.Engine {
	if alertService == nil {
		panic("🚨 alertService 未初始化！請檢查 Wire 依賴注入")
	}

	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Use(CorsConfig(), GinLogger(alertService), GinRecovery(false, alertService))

	// 調用 controller 內的方法來註冊所有 API
	controller.RegisterV1Routes(router, alertService, keycloak)

	return router
}

// 接收gin框架默認的日誌
func GinLogger(alertService *alert.Service) gin.HandlerFunc {
	logger := alertService.GetLogger()
	return func(c *gin.Context) {
		start := time.Now()
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		logger.Info(c.Request.URL.Path,
			zap.String("service", "gin"),
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("duration", cost),
			// zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}

// GinRecovery
func GinRecovery(stack bool, alertService *alert.Service) gin.HandlerFunc {
	logger := alertService.GetLogger()
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				//httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.String("service", "gin"),
						zap.Any("error", err),
						//zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("recovery_from_panic",
						zap.String("service", "gin"),
						zap.Any("error", err),
						//zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("recovery_from_panic",
						zap.String("service", "gin"),
						zap.Any("error", err),
						//zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
func CorsConfig() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"PUT", "POST", "GET", "DELETE"},
		AllowHeaders:     []string{"Content-type", "Access-Control-Allow-Origin", "Authorization", "Refresh-token", "realm"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	})
}
