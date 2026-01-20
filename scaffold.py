import os
import subprocess
import sys

# === 配置区 ===
PROJECT_NAME = "copycat-backend"
GO_MODULE = "copycat"  # go.mod 的名字

# 目录结构映射 (Key: 路径, Value: 初始文件内容 or None)
# None 表示只创建目录
STRUCTURE = {
    # 1. 核心目录
    "cmd/server": None,
    "config": None,
    "internal/api/v1/handler": None,
    "internal/api/v1/request": None,
    "internal/core/agent": None,
    "internal/core/crawler": None,
    "internal/core/llm": None,
    "internal/model": None,
    "internal/repository": None,
    "pkg/response": None,
    "pkg/logger": None,
    "docs/context": None, # 给 AI 看的文档目录
    
    # 2. 根目录文件
    "config/config.yaml": "server:\n  port: 8080\napp:\n  env: dev",
    "scripts": None,
}

# === 核心文件模版 ===

# main.go 模版 (最简 Gin 启动)
CONTENT_MAIN_GO = """package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"project": "CopyCat MVP",
		})
	})

	r.Run(":8080")
}
"""

# go.mod (会被 go mod init 覆盖，但作为占位)
CONTENT_GITIGNORE = """# Binaries
/server
/dist

# Config
config/config.prod.yaml
.env

# IDE
.idea/
.vscode/
.DS_Store
"""

# Agent Context 文档 (自动生成我们刚才商定的文档)
CONTENT_DOC_STACK = """# Tech Stack
- Go 1.21+
- Gin, GORM, Viper
- PostgreSQL
- Vue 3, Tailwind
"""

def create_structure(base_path):
    print(f"🚀 Initializing Project: {PROJECT_NAME}...")
    
    # 1. 创建目录和文件
    for path, content in STRUCTURE.items():
        full_path = os.path.join(base_path, path)
        
        # 如果是文件 (Key 包含扩展名或 content 不为空)
        if content is not None or "." in os.path.basename(path):
            os.makedirs(os.path.dirname(full_path), exist_ok=True)
            with open(full_path, "w", encoding="utf-8") as f:
                f.write(content if content else "")
            print(f"  [File] Created: {path}")
        else:
            os.makedirs(full_path, exist_ok=True)
            print(f"  [Dir]  Created: {path}")

    # 2. 写入 main.go
    with open(os.path.join(base_path, "cmd/server/main.go"), "w") as f:
        f.write(CONTENT_MAIN_GO)
        
    # 3. 写入 .gitignore
    with open(os.path.join(base_path, ".gitignore"), "w") as f:
        f.write(CONTENT_GITIGNORE)

    # 4. 写入 AI Context 文档
    with open(os.path.join(base_path, "docs/context/tech_stack.md"), "w") as f:
        f.write(CONTENT_DOC_STACK)

    print("✅ Structure created.")

def init_go_mod(base_path):
    print("📦 Initializing Go Module...")
    try:
        subprocess.run(["go", "mod", "init", GO_MODULE], check=True, cwd=base_path)
        subprocess.run(["go", "get", "github.com/gin-gonic/gin"], check=True, cwd=base_path)
        subprocess.run(["go", "mod", "tidy"], check=True, cwd=base_path)
        print("✅ Go dependencies installed.")
    except Exception as e:
        print(f"⚠️ Warning: Go init failed (do you have Go installed?): {e}")

if __name__ == "__main__":
    current_dir = os.getcwd()
    create_structure(current_dir)
    init_go_mod(current_dir)
    print(f"\n🎉 Project {PROJECT_NAME} initialized successfully!")
    print("👉 Next Step: Run 'go run cmd/server/main.go'")