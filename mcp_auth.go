package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// mcpRequest MCP 协议请求结构（简化版，仅用于判断请求类型）
type mcpRequest struct {
	Method string `json:"method"`
}

// mcpAuthMiddleware MCP 权限验证中间件
// 连接建立（initialize）时不验证，工具调用（tools/call）时需要验证
func mcpAuthMiddleware(appServer *AppServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果权限验证未启用，直接放行
		if !IsAuthEnabled() {
			c.Next()
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logrus.Errorf("读取 MCP 请求体失败: %v", err)
			c.Next()
			return
		}

		// 恢复请求体，以便后续处理
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 解析请求，判断是否是工具调用
		var req mcpRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// 无法解析，可能是非 JSON 请求，放行
			logrus.Debugf("无法解析 MCP 请求: %v", err)
			c.Next()
			return
		}

		// 判断请求类型
		// initialize: 初始化连接，不需要验证
		// tools/call: 调用工具，需要验证
		// tools/list: 列出工具，不需要验证
		if req.Method == "tools/call" {
			// 工具调用需要验证
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				logrus.Warnf("MCP 工具调用缺少 Authorization header, ip: %s", c.ClientIP())
				respondError(c, http.StatusUnauthorized, "MISSING_AUTH_HEADER",
					"MCP 工具调用需要提供 Authorization header", "请提供 Bearer token")
				c.Abort()
				return
			}

			// 验证 Bearer token 格式
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				logrus.Warnf("MCP 工具调用 Authorization header 格式无效, ip: %s", c.ClientIP())
				respondError(c, http.StatusUnauthorized, "INVALID_AUTH_HEADER",
					"无效的 Authorization header 格式", "格式应为: Bearer <token>")
				c.Abort()
				return
			}

			token := parts[1]

			// 验证 token
			if !ValidateAPIKey(token) {
				logrus.Warnf("MCP 工具调用使用无效的 API Key, ip: %s", c.ClientIP())
				respondError(c, http.StatusUnauthorized, "INVALID_API_KEY",
					"无效的 API Key", "请提供有效的 API Key")
				c.Abort()
				return
			}

			logrus.Debugf("MCP 工具调用权限验证通过, method: %s, ip: %s", req.Method, c.ClientIP())
		} else {
			// 其他请求（initialize, tools/list 等）不需要验证
			logrus.Debugf("MCP %s 请求不需要验证, ip: %s", req.Method, c.ClientIP())
		}

		c.Next()
	}
}
