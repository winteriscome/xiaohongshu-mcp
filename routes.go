package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupRoutes 设置路由配置
func setupRoutes(appServer *AppServer) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加中间件
	router.Use(errorHandlingMiddleware())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", healthHandler)

	// MCP 端点 - 使用官方 SDK 的 Streamable HTTP Handler
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return appServer.mcpServer
		},
		&mcp.StreamableHTTPOptions{
			JSONResponse: true, // 支持 JSON 响应
		},
	)

	// MCP GET端点 - 用于MCP Inspector的初始连接检查
	router.GET("/mcp", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "MCP Server is running",
			"version": "2.0.0",
			"name":    "xiaohongshu-mcp",
		})
	})

	// MCP POST端点 - 实际的MCP通信
	router.POST("/mcp", mcpAuthMiddleware(appServer.authManager), gin.WrapH(mcpHandler))
	router.POST("/mcp/*path", mcpAuthMiddleware(appServer.authManager), gin.WrapH(mcpHandler))

	// API 路由组 - 需要权限验证
	api := router.Group("/api/v1")
	api.Use(authMiddleware(appServer.authManager))
	{
		api.GET("/login/status", appServer.checkLoginStatusHandler)
		api.GET("/login/qrcode", appServer.getLoginQrcodeHandler)
		api.POST("/publish", appServer.publishHandler)
		api.POST("/publish_video", appServer.publishVideoHandler)
		api.GET("/feeds/list", appServer.listFeedsHandler)
		api.GET("/feeds/search", appServer.searchFeedsHandler)
		api.POST("/feeds/detail", appServer.getFeedDetailHandler)
		api.POST("/user/profile", appServer.userProfileHandler)
		api.POST("/feeds/comment", appServer.postCommentHandler)
	}

	return router
}
