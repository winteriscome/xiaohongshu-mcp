# 权限验证功能使用示例

## 1. 基本配置

### 1.1. 创建权限配置文件

创建 `configs/auth.json` 文件：

```json
{
  "enabled": true,
  "api_keys": [
    "default-key",
    "client-1-key",
    "client-2-key"
  ]
}
```

### 1.2. 启动服务

```bash
go run . -port :18060
```

## 2. API调用示例

### 2.1. 无API Key的请求（应该失败）

```bash
# 检查登录状态（拒绝）
curl http://localhost:18060/api/v1/login/status

# 响应：
# {
#   "error": "缺少API Key",
#   "code": "MISSING_API_KEY"
# }
```

### 2.2. 有效API Key的请求（应该成功）

```bash
# 使用有效API Key
API_KEY="default-key"

# 检查登录状态（允许）
curl -H "X-API-Key: $API_KEY" \
  http://localhost:18060/api/v1/login/status

# 发布内容（允许）
curl -X POST -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试","content":"测试内容","images":["https://example.com/image.jpg"]}' \
  http://localhost:18060/api/v1/publish
```

### 2.3. 无效API Key的请求（应该失败）

```bash
# 使用无效API Key
curl -H "X-API-Key: invalid-key" \
  http://localhost:18060/api/v1/login/status

# 响应：
# {
#   "error": "无效的API Key",
#   "code": "INVALID_API_KEY"
# }
```

## 3. MCP工具调用示例

### 3.1. 通过HTTP头传递API Key

```bash
# 方式1: 使用X-API-Key头
curl -H "X-API-Key: default-key" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "check_login_status",
      "arguments": {}
    }
  }' \
  http://localhost:18060/mcp

# 方式2: 使用x-custom-auth-headers头（MCP客户端推荐）
curl -H "x-custom-auth-headers: default-key" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "check_login_status",
      "arguments": {}
    }
  }' \
  http://localhost:18060/mcp
```

## 4. 配置管理

权限验证配置只能通过修改配置文件 `configs/auth.json` 来管理：

### 4.1. 启用/禁用权限验证

```bash
# 编辑配置文件
vim configs/auth.json

# 启用权限验证
{
  "enabled": true,
  "api_keys": [
    "default-key",
    "your-api-key"
  ]
}

# 禁用权限验证
{
  "enabled": false,
  "api_keys": [
    "default-key"
  ]
}
```

### 4.2. 添加/删除API Key

```bash
# 编辑配置文件添加新的API Key
{
  "enabled": true,
  "api_keys": [
    "default-key",
    "client-1-key",
    "client-2-key"
  ]
}

# 删除API Key只需从列表中移除即可
{
  "enabled": true,
  "api_keys": [
    "default-key",
    "client-1-key"
  ]
}
```

## 5. 错误处理

### 5.1. 常见错误响应

```json
// 缺少API Key
{
  "error": "缺少API Key",
  "code": "MISSING_API_KEY"
}

// 无效的API Key
{
  "error": "无效的API Key",
  "code": "INVALID_API_KEY"
}
```

## 6. 安全建议

1. **生产环境配置**:
   - 启用权限验证 (`enabled: true`)
   - 使用强随机API Key
   - 定期轮换API Key

2. **监控和日志**:
   - 监控权限验证失败的请求
   - 记录API Key使用情况
   - 设置异常访问告警

3. **API Key管理**:
   - 为每个客户端分配唯一的API Key
   - 定期审查和清理不使用的API Key
   - 使用强随机字符串作为API Key