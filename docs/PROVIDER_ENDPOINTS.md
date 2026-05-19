# Provider API Endpoints

本文档描述了从原项目迁移到pomclaw的两个提供商相关的新API端点。

## 新增端点

### 1. GET /v1/providers/{id}/models

**功能**: 列出指定提供商可用的模型列表

**请求**:
```bash
GET /v1/providers/01KRX7TA2BEV8JJ6NQ8RTAVREF/models
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
{
  "models": [
    {
      "id": "claude-opus-4",
      "name": "Claude Opus 4"
    },
    {
      "id": "claude-sonnet-4",
      "name": "Claude Sonnet 4"
    },
    {
      "id": "claude-haiku-3",
      "name": "Claude Haiku 3"
    }
  ]
}
```

**支持的提供商类型**:
- `claude_cli` - 返回: sonnet, opus, haiku
- `acp` - 返回: claude, codex, gemini
- `chatgpt_oauth` - 返回: gpt-5.4, gpt-5.3-codex 等
- `bailian` - 返回: qwen3.6-plus, kimi-k2.5 等
- `dashscope` - 返回: qwen3.6-plus, wan2.6-image 等
- `minimax` - 返回: MiniMax-Text-01, MiniMax-M2.5 等
- `anthropic_native` - 从 Anthropic API 获取
- `gemini_native` - 从 Google Gemini API 获取
- OpenAI 兼容 - 从 `/v1/models` 端点获取

**错误处理**:
- 提供商不存在: 返回 404
- 无效的用户ID: 返回错误
- API 连接失败: 返回空列表

---

### 2. POST /v1/providers/{id}/verify

**功能**: 验证提供商和模型组合是否有效

**请求**:
```bash
POST /v1/providers/01KRX7TA2BEV8JJ6NQ8RTAVREF/verify
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "model": "claude-opus-4"
}
```

**成功响应**:
```json
{
  "valid": true
}
```

**失败响应**:
```json
{
  "valid": false,
  "error": "Invalid model. Use: sonnet, opus, or haiku"
}
```

**验证逻辑**:
- **claude_cli**: 检查模型是否在 (sonnet, opus, haiku) 中
- **acp**: 检查模型是否在 (claude, codex, gemini) 中
- **非聊天模型** (图像/视频生成): 直接返回有效，但注明不可通过聊天验证
- **其他提供商**: 验证API密钥存在并尝试连接

**超时**: 30秒

---

## 迁移细节

### 新增文件

1. **Handler 层**:
   - `internal/handler/listprovidermodelshandler.go` - 列表模型处理器
   - `internal/handler/verifyproviderhandler.go` - 验证处理器

2. **Logic 层**:
   - `internal/logic/provider_models.go` - 包含模型列表逻辑和API调用
   - `internal/logic/provider_verify.go` - 包含验证逻辑

3. **路由**:
   - `internal/handler/routes.go` - 新增两个路由注册

### 支持的提供商模型数据库

支持的模型信息包括：

- **Claude CLI**: sonnet, opus, haiku
- **ACP**: claude, codex, gemini
- **ChatGPT OAuth**: gpt-5.4, gpt-5.3-codex, gpt-5.2-codex, gpt-5.1, 等
- **Bailian**: qwen3.6-plus, qwen3.5-plus, kimi-k2.5, GLM-5, MiniMax-M2.5
- **DashScope**: qwen3.6-plus, qwen3.5-plus, qwen3.5-flash, wan2.6-image, 等
- **MiniMax**: MiniMax-Text-01, MiniMax-M2.5, MiniMax-M2.7, image-01, 等

### API 调用支持

支持从以下API获取模型列表：

- **Anthropic**: `/v1/models` 端点
- **Google Gemini**: `/v1beta/models` 端点
- **OpenAI 兼容**: `/v1/models` 端点

---

## 使用示例

### 获取特定提供商的模型列表

```bash
# 创建或获取 JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' | jq -r '.access_token')

# 列出模型
curl -X GET http://localhost:8080/v1/providers/provider-id-here/models \
  -H "Authorization: Bearer $TOKEN"
```

### 验证提供商

```bash
curl -X POST http://localhost:8080/v1/providers/provider-id-here/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-opus-4"}'
```

---

## 原项目对应位置

从以下原项目文件迁移：
- `/Users/zhengjm/product/jourmey/goclaw/internal/http/provider_models.go`
- `/Users/zhengjm/product/jourmey/goclaw/internal/http/provider_models_fetch.go`
- `/Users/zhengjm/product/jourmey/goclaw/internal/http/provider_models_catalog.go`
- `/Users/zhengjm/product/jourmey/goclaw/internal/http/provider_verify.go`

