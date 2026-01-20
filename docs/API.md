# CopyCat API Documentation

> **Base URL**: `http://localhost:8088/api/v1`

---

## 目录

- [认证说明](#认证说明)
- [通用响应格式](#通用响应格式)
- [用户模块](#用户模块)
- [项目模块](#项目模块)
- [爬虫模块](#爬虫模块)
- [设置模块](#设置模块)
- [分析模块](#分析模块)
- [健康检查](#健康检查)
- [错误码说明](#错误码说明)

---

## 认证说明

除了 `注册` 和 `登录` 接口外，所有接口都需要在请求头中携带 JWT Token：

```
Authorization: Bearer <your_token>
```

Token 有效期：72 小时

---

## 通用响应格式

所有接口返回统一的 JSON 格式：

```json
{
  "code": 0,       // 状态码，0 表示成功
  "msg": "success", // 状态消息
  "data": {}       // 响应数据（可选）
}
```

---

## 用户模块

### 注册

创建新用户账号。

**请求**

```
POST /api/v1/register
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | ✅ | 用户邮箱 |
| password | string | ✅ | 密码 (最少6位) |
| nickname | string | ❌ | 用户昵称 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "123456",
    "nickname": "用户昵称"
  }'
```

**响应示例**

```json
{
  "code": 0,
  "msg": "user registered successfully"
}
```

---

### 登录

用户登录获取 JWT Token。

**请求**

```
POST /api/v1/login
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | ✅ | 用户邮箱 |
| password | string | ✅ | 密码 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "123456"
  }'
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | JWT 认证令牌 |
| user | object | 用户信息对象 |
| user.id | int | 用户唯一 ID |
| user.email | string | 用户邮箱 |
| user.nickname | string | 用户昵称 |
| user.created_at | string | 账号创建时间 |
| user.updated_at | string | 账号更新时间 |

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "nickname": "用户昵称",
      "created_at": "2026-01-18T17:00:00+08:00",
      "updated_at": "2026-01-18T17:00:00+08:00"
    }
  }
}
```

---

### 获取用户信息

获取当前登录用户的信息。

**请求**

```
GET /api/v1/user/profile
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 用户唯一 ID |
| email | string | 用户邮箱 |
| nickname | string | 用户昵称 |
| created_at | string | 账号创建时间 (ISO 8601 格式) |
| updated_at | string | 账号更新时间 (ISO 8601 格式) |

**请求示例**

```bash
curl http://localhost:8088/api/v1/user/profile \
  -H "Authorization: Bearer <token>"
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "用户昵称",
    "created_at": "2026-01-18T17:00:00+08:00",
    "updated_at": "2026-01-18T17:00:00+08:00"
  }
}
```

---

## 项目模块

### 项目字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 项目唯一 ID |
| user_id | int | 所属用户 ID |
| source_url | string | 原始文案来源 URL (小红书/公众号链接) |
| source_content | string | 原始文案内容 |
| analysis_result | object | LLM 分析结果 (情绪/结构/关键词) |
| new_topic | string | 用户输入的新主题 |
| generated_content | string | AI 生成的仿写文案 |
| status | string | 项目状态：draft(草稿) / analyzed(已分析) / completed(已完成) |
| created_at | string | 项目创建时间 |
| updated_at | string | 项目更新时间 |

---

### 创建项目

创建新的文案创作项目。

**请求**

```
POST /api/v1/projects
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source_url | string | ❌ | 原始文案来源 URL |
| source_content | string | ✅ | 原始文案内容 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "source_url": "https://www.xiaohongshu.com/explore/xxx",
    "source_content": "这是一段爆款文案..."
  }'
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": 1,
    "source_url": "https://www.xiaohongshu.com/explore/xxx",
    "source_content": "这是一段爆款文案...",
    "status": "draft",
    "created_at": "2026-01-18T17:00:00+08:00",
    "updated_at": "2026-01-18T17:00:00+08:00"
  }
}
```

---

### 获取项目列表

获取当前用户的项目列表（分页）。

**请求**

```
GET /api/v1/projects
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | ❌ | 页码 (默认 1) |
| page_size | int | ❌ | 每页数量 (默认 10, 最大 100) |

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | array | 项目列表 |
| total | int | 项目总数 |
| page | int | 当前页码 |
| page_size | int | 每页数量 |

**请求示例**

```bash
curl "http://localhost:8088/api/v1/projects?page=1&page_size=10" \
  -H "Authorization: Bearer <token>"
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": 1,
        "source_content": "这是一段爆款文案...",
        "status": "draft",
        "created_at": "2026-01-18T17:00:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

---

### 获取项目详情

获取单个项目的详细信息。

**请求**

```
GET /api/v1/projects/:id
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | UUID | ✅ | 项目 ID (路径参数) |

**请求示例**

```bash
curl http://localhost:8088/api/v1/projects/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <token>"
```

---

### 更新项目

更新项目信息。

**请求**

```
PUT /api/v1/projects/:id
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source_url | string | ❌ | 原始文案来源 URL |
| source_content | string | ❌ | 原始文案内容 |
| new_topic | string | ❌ | 新主题 (用于仿写) |
| generated_content | string | ❌ | 生成的仿写内容 |
| status | string | ❌ | 项目状态 (draft/analyzed/completed) |

**请求示例**

```bash
curl -X PUT http://localhost:8088/api/v1/projects/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "new_topic": "新品推广",
    "status": "analyzed"
  }'
```

---

### 删除项目

删除指定项目。

**请求**

```
DELETE /api/v1/projects/:id
```

**请求示例**

```bash
curl -X DELETE http://localhost:8088/api/v1/projects/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <token>"
```

**响应示例**

```json
{
  "code": 0,
  "msg": "project deleted"
}
```

---

## 爬虫模块

### 爬取内容

爬取小红书笔记内容。

**请求**

```
POST /api/v1/crawl
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| url | string | ✅ | 小红书笔记 URL |

**支持的 URL 格式**

- `https://www.xiaohongshu.com/explore/xxx`
- `https://www.xiaohongshu.com/discovery/item/xxx`
- `https://xhslink.com/xxx`

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 是否爬取成功 |
| platform | string | 平台标识 (xiaohongshu) |
| content | object | 笔记内容对象 |
| error | string | 错误信息 (失败时) |

**笔记内容 (content) 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| note_id | string | 笔记唯一 ID |
| title | string | 笔记标题 |
| content | string | 笔记正文内容 |
| type | string | 笔记类型：normal(图文) / video(视频) |
| cover_url | string | 封面图 URL |
| author_id | string | 作者 ID |
| author_name | string | 作者昵称 |
| author_avatar | string | 作者头像 URL |
| images | array | 图片 URL 列表 |
| video | object | 视频信息 (视频笔记时存在) |
| tags | array | 标签列表 |
| like_count | int | 点赞数 |
| comment_count | int | 评论数 |
| collect_count | int | 收藏数 |
| share_count | int | 分享数 |
| publish_time | string | 发布时间 |
| crawl_time | string | 爬取时间 |
| source_url | string | 原始链接 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/crawl \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "url": "https://www.xiaohongshu.com/explore/xxx"
  }'
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true,
    "platform": "xiaohongshu",
    "content": {
      "note_id": "xxx",
      "title": "笔记标题",
      "content": "笔记正文内容...",
      "type": "normal",
      "author_name": "作者昵称",
      "images": [
        "https://example.com/image1.jpg",
        "https://example.com/image2.jpg"
      ],
      "tags": ["标签1", "标签2"],
      "like_count": 1234,
      "collect_count": 567,
      "comment_count": 89,
      "crawl_time": "2026-01-18T17:00:00+08:00",
      "source_url": "https://www.xiaohongshu.com/explore/xxx"
    }
  }
}
```

---

## 设置模块

### 获取 LLM 配置

获取用户的 LLM 配置。

**请求**

```
GET /api/v1/settings/llm
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "provider": "openai",
    "api_key": "sk-x****xxxx",
    "model": "gpt-3.5-turbo",
    "base_url": ""
  }
}
```

---

### 保存 LLM 配置

保存用户的 LLM 配置。

**请求**

```
POST /api/v1/settings/llm
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | ❌ | 服务商 (openai/deepseek/anthropic) |
| api_key | string | ✅ | API 密钥 |
| model | string | ❌ | 模型名称 |
| base_url | string | ❌ | 自定义 API 地址 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/settings/llm \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "provider": "openai",
    "api_key": "sk-xxxxxx",
    "model": "gpt-4",
    "base_url": ""
  }'
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "配置保存成功"
  }
}
```

---

## 分析模块

### 分析爆款内容

使用 LLM 分析文案的情绪、结构和关键词。

**请求**

```
POST /api/v1/analyze
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | ✅ | 待分析的文案内容 |
| project_id | string | ❌ | 关联项目ID，分析结果将保存到项目 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/analyze \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "content": "这款面霜真的太好用了！...",
    "project_id": "uuid-xxx"
  }'
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "emotion": {
      "primary": "惊喜",
      "intensity": 0.85,
      "tags": ["种草", "推荐", "好用"]
    },
    "structure": [
      {"title": "开篇引入", "description": "直接表达惊喜感受"},
      {"title": "痛点挖掘", "description": "提出皮肤问题"},
      {"title": "解决方案", "description": "介绍产品效果"},
      {"title": "行动号召", "description": "引导购买"}
    ],
    "keywords": ["面霜", "好用", "推荐"],
    "tone": "casual",
    "word_count": 350
  }
}
```

---

### 生成仿写文案

基于分析结果和新主题生成仿写文案。

**请求**

```
POST /api/v1/generate
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| project_id | string | ✅ | 项目ID（需先完成分析） |
| new_topic | string | ✅ | 新主题描述 |

**请求示例**

```bash
curl -X POST http://localhost:8088/api/v1/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "project_id": "uuid-xxx",
    "new_topic": "护眼仪使用体验"
  }'
```

**响应示例**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "generated_content": "这款护眼仪真的太惊艳了！..."
  }
}
```

---

## 健康检查

### Ping

检查服务是否正常运行。

**请求**

```
GET /ping
```

> ⚠️ 注意：此接口不需要 `/api/v1` 前缀，也不需要认证

**请求示例**

```bash
curl http://localhost:8088/ping
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| message | string | 响应消息 (pong) |
| project | string | 项目名称 |

**响应示例**

```json
{
  "message": "pong",
  "project": "CopyCat MVP"
}
```

---

## 错误码说明

| Code | HTTP Status | 说明 |
|------|-------------|------|
| 0 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未认证或 Token 无效/过期 |
| 403 | 403 | 无权限访问该资源 |
| 404 | 404 | 资源不存在 |
| 500 | 500 | 服务器内部错误 |

**错误响应示例**

```json
{
  "code": 401,
  "msg": "invalid or expired token"
}
```

---

## 快速开始

```bash
# 1. 注册用户
curl -X POST http://localhost:8088/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456","nickname":"Test"}'

# 2. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8088/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

echo "Token: $TOKEN"

# 3. 使用 Token 调用其他接口
curl http://localhost:8088/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"

# 4. 爬取小红书内容
curl -X POST http://localhost:8088/api/v1/crawl \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url":"https://www.xiaohongshu.com/explore/xxx"}'
```
