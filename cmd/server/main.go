package main

import (
	"fmt"
	"log"

	"copycat/config"
	"copycat/internal/api"
	"copycat/internal/model"
	"copycat/pkg/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化文件日志（API 日志）
	apiLogDir := "logs/api"
	if err := logger.InitFileLogger(apiLogDir, logger.DEBUG); err != nil {
		log.Printf("初始化 API 日志失败: %v，继续使用控制台日志", err)
	} else {
		defer logger.Close()
	}

	// 3. 初始化 LLM 日志
	llmLogDir := "logs/llm"
	if err := logger.InitLLMLogger(llmLogDir, logger.DEBUG); err != nil {
		log.Printf("初始化 LLM 日志失败: %v，继续使用控制台日志", err)
	}

	log.Printf("Starting %s in %s mode...", cfg.App.Name, cfg.App.Env)

	// 3. 初始化数据库
	db, err := config.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 4. 自动迁移（注意顺序：BatchTask 需要在 Project 之前，因为 Project 有外键引用 BatchTask）
	if err := config.AutoMigrate(db, &model.User{}, &model.UserSettings{}, &model.BatchTask{}, &model.Project{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 5. 设置路由
	r := api.SetupRouter(db)

	// 6. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("API Documentation: http://localhost%s/api/v1", addr)
	log.Printf("API 日志文件: %s", logger.GetLogFilePath(apiLogDir))

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
