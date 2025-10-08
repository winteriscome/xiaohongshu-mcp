package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, x-custom-auth-headers, mcp-protocol-version, mcp-client-info, mcp-server-info")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// errorHandlingMiddleware 错误处理中间件
func errorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logrus.Errorf("服务器内部错误: %v, path: %s", recovered, c.Request.URL.Path)

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR",
			"服务器内部错误", recovered)
	})
}

// authMiddleware 权限验证中间件
func authMiddleware(authManager *AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果权限验证未启用，直接通过
		if !authManager.IsEnabled() {
			c.Next()
			return
		}

		// 获取API Key
		apiKey := getAPIKeyFromRequest(c)
		if apiKey == "" {
			logrus.Warnf("缺少API Key: %s", c.Request.URL.Path)
			respondError(c, http.StatusUnauthorized, "MISSING_API_KEY", "缺少API Key", nil)
			c.Abort()
			return
		}

		// 验证API Key
		valid := authManager.ValidateAPIKey(apiKey)
		if !valid {
			logrus.Warnf("无效的API Key: %s", c.Request.URL.Path)
			respondError(c, http.StatusUnauthorized, "INVALID_API_KEY", "无效的API Key", nil)
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		authCtx := &AuthContext{
			APIKey:  apiKey,
			IsValid: true,
		}
		c.Set("auth", authCtx)

		logrus.Debugf("权限验证通过: API Key=%s, Path=%s", apiKey, c.Request.URL.Path)
		c.Next()
	}
}

// mcpAuthMiddleware MCP权限验证中间件
func mcpAuthMiddleware(authManager *AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果权限验证未启用，直接通过
		if !authManager.IsEnabled() {
			c.Next()
			return
		}

		// 对于GET请求，允许通过（MCP Inspector的初始连接检查）
		if c.Request.Method == "GET" {
			logrus.Debugf("MCP GET请求，跳过权限验证")
			c.Next()
			return
		}

		// 对于POST请求，进行权限验证
		apiKey := getAPIKeyFromRequest(c)
		if apiKey == "" {
			logrus.Warnf("MCP POST请求缺少API Key")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
			c.Abort()
			return
		}

		// 验证API Key
		valid := authManager.ValidateAPIKey(apiKey)
		if !valid {
			logrus.Warnf("MCP POST请求无效的API Key")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		authCtx := &AuthContext{
			APIKey:  apiKey,
			IsValid: true,
		}
		c.Set("auth", authCtx)

		logrus.Debugf("MCP权限验证通过: API Key=%s", apiKey)
		c.Next()
	}
}

// getAPIKeyFromRequest 从请求中获取API Key
func getAPIKeyFromRequest(c *gin.Context) string {
	// 优先从Header中获取
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// 从x-custom-auth-headers头获取（MCP客户端使用）
	customAuth := c.GetHeader("x-custom-auth-headers")
	if customAuth != "" {
		return customAuth
	}

	// 从Authorization头获取
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// 从查询参数获取
	apiKey = c.Query("api_key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}
