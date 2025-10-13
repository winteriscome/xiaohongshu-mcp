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
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

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

// authMiddleware Token-based 权限验证中间件
// 从 Authorization header 中提取 Bearer token 并验证
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果权限验证未启用，直接放行
		if !IsAuthEnabled() {
			c.Next()
			return
		}

		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logrus.Warnf("缺少 Authorization header, path: %s, ip: %s", c.Request.URL.Path, c.ClientIP())
			respondError(c, http.StatusUnauthorized, "MISSING_AUTH_HEADER",
				"缺少 Authorization header", "请提供 Bearer token")
			c.Abort()
			return
		}

		// 验证 Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logrus.Warnf("无效的 Authorization header 格式, path: %s, ip: %s", c.Request.URL.Path, c.ClientIP())
			respondError(c, http.StatusUnauthorized, "INVALID_AUTH_HEADER",
				"无效的 Authorization header 格式", "格式应为: Bearer <token>")
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 token
		if !ValidateAPIKey(token) {
			logrus.Warnf("无效的 API Key, path: %s, ip: %s", c.Request.URL.Path, c.ClientIP())
			respondError(c, http.StatusUnauthorized, "INVALID_API_KEY",
				"无效的 API Key", "请提供有效的 API Key")
			c.Abort()
			return
		}

		// 验证通过，记录日志
		logrus.Debugf("API Key 验证通过, path: %s, ip: %s", c.Request.URL.Path, c.ClientIP())
		c.Next()
	}
}
