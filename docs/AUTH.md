# 权限验证功能说明

## 概述

MCP服务现在支持基于API Key的简单权限验证功能，确保只有授权的客户端才能调用相关服务。该功能特别适用于将MCP服务部署在公网环境中的场景。

## 功能特性

- **API Key验证**: 基于API Key的身份验证
- **简单配置**: 只需配置API Key列表，无需复杂的权限分级
- **配置热更新**: 支持运行时更新权限配置
- **向后兼容**: 默认关闭，不影响现有使用

## 配置说明

### 配置文件位置
权限配置文件位于：`./configs/auth.json`

### 配置示例
```json
{
  "enabled": false,
  "api_keys": [
    "default-key",
    "your-api-key-1",
    "your-api-key-2"
  ]
}
```

### 配置字段说明

- `enabled`: 是否启用权限验证（默认false）
- `api_keys`: 有效的API Key列表

## API Key使用方式

### HTTP API调用
```bash
# 方式1: 使用X-API-Key头
curl -H "X-API-Key: your-api-key" http://localhost:18060/api/v1/login/status

# 方式2: 使用Authorization头
curl -H "Authorization: Bearer your-api-key" http://localhost:18060/api/v1/login/status

# 方式3: 使用查询参数
curl "http://localhost:18060/api/v1/login/status?api_key=your-api-key"
```

### MCP工具调用
MCP工具调用时，API Key需要通过HTTP头传递。支持以下方式：

1. **X-API-Key头**：
```json
{
  "meta": {
    "headers": {
      "X-API-Key": "your-api-key"
    }
  }
}
```

2. **x-custom-auth-headers头**（MCP客户端推荐）：
```json
{
  "meta": {
    "headers": {
      "x-custom-auth-headers": "your-api-key"
    }
  }
}
```

3. **Authorization头**：
```json
{
  "meta": {
    "headers": {
      "Authorization": "Bearer your-api-key"
    }
  }
}
```

### MCP Inspector客户端支持
服务器已完全支持MCP Inspector v0.17.0客户端，包括：
- ✅ 支持 `mcp-protocol-version` 头
- ✅ 支持 `mcp-client-info` 和 `mcp-server-info` 头
- ✅ 支持Streamable HTTP传输类型
- ✅ 完整的CORS跨域支持
- ✅ 使用官方MCP SDK，完全兼容MCP协议标准
- ✅ 支持GET请求（MCP Inspector初始连接检查）
- ✅ 支持POST请求（实际MCP通信）

**端点说明**：
- `GET /mcp`：返回服务器状态信息，用于MCP Inspector的初始连接检查
- `POST /mcp`：实际的MCP通信，使用官方SDK处理

**权限验证说明**：
- GET请求：无需权限验证（用于MCP Inspector的初始连接检查）
- POST请求：需要API Key验证（用于实际的MCP工具调用）

**注意**：MCP Inspector客户端需要发送标准JSON-RPC 2.0格式的请求，包含 `jsonrpc: "2.0"` 字段。

## 配置管理

权限验证配置只能通过修改配置文件 `configs/auth.json` 来管理，不支持通过API动态修改。

## 安全建议

1. **启用权限验证**: 在生产环境中务必启用权限验证
2. **使用强API Key**: 生成足够复杂和随机的API Key
3. **定期轮换**: 定期更换API Key
4. **监控日志**: 定期检查权限验证失败的日志

## 默认配置

服务启动时会自动创建默认配置文件，包含：
- 默认API Key: `default-key`
- 权限验证默认关闭

## 故障排除

### 权限验证失败
1. 检查API Key是否正确
2. 确认权限验证已启用
3. 查看服务器日志获取详细错误信息

### 配置不生效
1. 确认配置文件格式正确
2. 检查文件权限
3. 重启服务使配置生效

## 使用示例

### 场景1: 启用权限验证
```json
{
  "enabled": true,
  "api_keys": [
    "client-1-key",
    "client-2-key"
  ]
}
```

### 场景2: 添加新的API Key
```bash
curl -X POST -H "X-API-Key: admin-key" \
  -H "Content-Type: application/json" \
  -d '{"api_key":"new-client-key"}' \
  http://localhost:18060/api/v1/auth/keys
```

### 场景3: 禁用权限验证
```bash
curl -X PUT -H "X-API-Key: admin-key" \
  -H "Content-Type: application/json" \
  -d '{"enabled":false}' \
  http://localhost:18060/api/v1/auth/config
```