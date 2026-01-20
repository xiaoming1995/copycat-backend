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

	// 2. 初始化文件日志
	logDir := "logs"
	if err := logger.InitFileLogger(logDir, logger.DEBUG); err != nil {
		log.Printf("初始化文件日志失败: %v，继续使用控制台日志", err)
	} else {
		defer logger.Close()
	}

	log.Printf("Starting %s in %s mode...", cfg.App.Name, cfg.App.Env)

	// 3. 初始化数据库
	db, err := config.InitDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 4. 自动迁移
	if err := config.AutoMigrate(db, &model.User{}, &model.Project{}, &model.UserSettings{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 5. 设置路由
	r := api.SetupRouter(db)

	// 6. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("API Documentation: http://localhost%s/api/v1", addr)
	log.Printf("日志文件: %s", logger.GetLogFilePath(logDir))

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
