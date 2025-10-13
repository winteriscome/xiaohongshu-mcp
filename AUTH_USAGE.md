# API 权限验证使用说明

## 快速开始

### 1. 配置权限验证

复制配置文件并编辑：

```bash
cp auth.yaml.example auth.yaml
```

编辑 `auth.yaml`：

```yaml
# 是否启用权限验证
enabled: true

# API Key 列表
api_keys:
  - "your-secret-api-key-1"
  - "your-secret-api-key-2"
```

### 2. 启动服务

```bash
./xiaohongshu-mcp -port :18060
```

### 3. 使用 API

**使用 curl 调用 API**：

```bash
curl -X GET http://localhost:18060/api/v1/login/status \
  -H "Authorization: Bearer your-secret-api-key-1"
```

## MCP Inspector 使用

**重要提示**：启用权限验证后，MCP 工具调用需要提供 API Key：

```bash
# MCP Inspector 需要在 HTTP header 中提供 Authorization
npx @modelcontextprotocol/inspector http://localhost:18060/mcp \
  -H "Authorization: Bearer your-secret-api-key-1"
```

**权限验证说明**：
- ✅ **MCP 连接建立**（initialize）- 不需要验证
- ✅ **工具列表**（tools/list）- 不需要验证
- ⚠️ **工具调用**（tools/call）- 需要验证 API Key

如果在 MCP Inspector 中调用工具时看到权限错误，请确保：
1. 已创建并配置 `auth.yaml` 文件
2. 在 HTTP header 中提供了正确的 `Authorization: Bearer <your-api-key>`

## Claude Code 配置

如果需要通过 Claude Code 使用 HTTP API（非 MCP），可以在配置中添加 headers：

```json
{
  "xiaohongshu-api": {
    "url": "http://localhost:18060/mcp",
    "headers": {
      "Authorization": "Bearer your-secret-api-key-1"
    }
  }
}
```

添加 mcp 服务器
```shell
claude mcp add --transport http xiaohongshu-mcp http://localhost:18060/mcp --header "Authorization: Bearer your-secret-api-key-1" 
```

删除 mcp 服务器
```shell
claude mcp remove xiaohongshu-mcp
```

## 端点说明

### 不需要验证的端点

- `/health` - 健康检查
- `/mcp` (initialize) - MCP 连接建立
- `/mcp` (tools/list) - 列出可用工具

### 需要验证的端点（启用认证时）

- `/mcp` (tools/call) - MCP 工具调用 ⚠️ **需要 API Key**
- `/api/v1/*` - 所有 HTTP API 端点
  - `GET /api/v1/login/status` - 检查登录状态
  - `GET /api/v1/login/qrcode` - 获取登录二维码
  - `POST /api/v1/publish` - 发布内容
  - `POST /api/v1/publish_video` - 发布视频
  - `GET /api/v1/feeds/list` - 获取 Feed 列表
  - `GET /api/v1/feeds/search` - 搜索 Feeds
  - `POST /api/v1/feeds/detail` - 获取 Feed 详情
  - `POST /api/v1/user/profile` - 获取用户主页
  - `POST /api/v1/feeds/comment` - 发表评论

## 禁用权限验证

### 方法 1：删除配置文件

```bash
rm auth.yaml
# 重启服务
```

### 方法 2：修改配置文件

编辑 `auth.yaml`，设置 `enabled: false`：

```yaml
enabled: false
api_keys:
  - "your-secret-api-key-1"
```

## 错误处理

### 401 Unauthorized: MISSING_AUTH_HEADER

**原因**：请求未包含 Authorization header

**解决**：添加 `Authorization: Bearer <api-key>` header

```bash
curl -X GET http://localhost:18060/api/v1/login/status \
  -H "Authorization: Bearer your-api-key"
```

### 401 Unauthorized: INVALID_AUTH_HEADER

**原因**：Authorization header 格式不正确

**解决**：确保格式为 `Bearer <token>`，注意 "Bearer" 和 token 之间有一个空格

### 401 Unauthorized: INVALID_API_KEY

**原因**：提供的 API Key 无效

**解决**：检查 API Key 是否与 `auth.yaml` 中配置的一致

## 生成强 API Key

```bash
# 使用 openssl 生成 32 字符的随机 API Key
openssl rand -base64 32

# 使用 uuidgen 生成 UUID 作为 API Key
uuidgen
```

## 安全建议

1. **不要将 auth.yaml 提交到版本控制**
   - 该文件已添加到 `.gitignore`

2. **使用强密码**
   - 建议至少 32 字符的随机字符串

3. **定期轮换 Key**
   - 支持配置多个 Key，方便平滑过渡

4. **生产环境**
   - 始终使用 HTTPS
   - 配合防火墙规则使用
