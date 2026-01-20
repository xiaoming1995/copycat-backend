# CopyCat Backend

CopyCat 是一个基于 Go 语言开发的后端项目，结合了爬虫、大模型 (LLM) 和智能代理 (Agent) 技术。

## 📁 目录结构

```text
copycat-backend/
├── cmd/                         # 【入口】项目启动入口
│   └── server/
│       └── main.go              # 服务启动主文件 (Gin 框架启动点)
├── config/                      # 【配置】存放静态配置文件
│   └── config.yaml              # 基础配置文件 (如端口号, 环境等)
├── internal/                    # 【核心】存放私有业务逻辑，外部项目无法引用
│   ├── api/                     # 接口层
│   │   └── v1/
│   │       ├── handler/         # 控制器层 (处理具体请求逻辑)
│   │       └── request/         # 请求结构定义 (参数验证/绑定)
│   ├── core/                    # 核心业务组件
│   │   ├── agent/               # 智能代理相关逻辑
│   │   ├── crawler/             # 爬虫相关逻辑
│   │   └── llm/                 # 大模型集成逻辑
│   ├── model/                   # 数据模型 (数据库结构体定义)
│   └── repository/              # 数据持久层 (数据库 CRUD 操作)
├── pkg/                         # 【包】存放可被外部引用的公共工具包
│   ├── logger/                  # 日志包
│   └── response/                # 统一返回结构定义
├── docs/                        # 【文档】项目相关文档
│   └── context/
│       └── tech_stack.md        # 技术栈说明
├── scripts/                     # 【脚本】运维或辅助脚本目录
├── go.mod                       # Go 模块依赖定义
├── go.sum                       # 依赖库版本锁定文件
└── .gitignore                   # Git 忽略文件配置
```

## 🛠 技术栈

- **后端框架**: Gin (Go)
- **数据库**: PostgreSQL (推荐) / GORM (ORM)
- **配置管理**: YAML / Viper
- **核心功能**: 网络爬虫, 大模型接入, 任务代理 (AI Agent)

## 🚀 快速开始

### 环境依赖
- Go 1.21+

### 运行服务
```bash
go run cmd/server/main.go
```

默认请求地址：`http://localhost:8088/ping`
