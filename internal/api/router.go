package api

import (
	"copycat/internal/api/handler"
	"copycat/internal/api/middleware"
	"copycat/internal/core/agent"
	"copycat/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 设置路由
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	// 初始化服务
	contentService := agent.NewContentService(projectRepo)

	// 初始化处理器
	userHandler := handler.NewUserHandler(userRepo)
	projectHandler := handler.NewProjectHandler(projectRepo)
	crawlerHandler := handler.NewCrawlerHandler(contentService)
	settingsHandler := handler.NewSettingsHandler(db)
	analysisHandler := handler.NewAnalysisHandler(db)
	batchHandler := handler.NewBatchHandler(db, contentService)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 公开路由 (无需认证)
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)

		// 需要认证的路由
		auth := v1.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// 用户相关
			auth.GET("/user/profile", userHandler.GetProfile)
			auth.PUT("/user/profile", userHandler.UpdateProfile)
			auth.PUT("/user/password", userHandler.ChangePassword)

			// 项目相关
			auth.POST("/projects", projectHandler.Create)
			auth.GET("/projects", projectHandler.List)
			auth.GET("/projects/check", projectHandler.GetByURL)       // 检查链接是否已分析（需在 :id 之前）
			auth.DELETE("/projects/batch", projectHandler.BatchDelete) // 批量删除（需在 :id 之前）
			auth.GET("/projects/:id", projectHandler.Get)
			auth.PUT("/projects/:id", projectHandler.Update)
			auth.DELETE("/projects/:id", projectHandler.Delete)

			// 爬虫相关
			auth.POST("/crawl", crawlerHandler.Crawl)

			// 设置相关
			auth.GET("/settings/llm", settingsHandler.GetLLMConfig)
			auth.POST("/settings/api-config", settingsHandler.SaveApiConfig)           // 模块1: API 配置
			auth.POST("/settings/model-config", settingsHandler.SaveModelConfig)       // 模块2: 模型选择
			auth.POST("/settings/generate-config", settingsHandler.SaveGenerateConfig) // 模块3: 生成设置

			// 分析与生成相关
			auth.POST("/analyze", analysisHandler.Analyze)
			auth.POST("/analyze-images", analysisHandler.AnalyzeImages)
			auth.POST("/generate", analysisHandler.Generate)

			// 批量任务相关
			auth.POST("/batch/analyze", batchHandler.CreateBatchAnalyze)
			auth.GET("/batch/:id", batchHandler.GetBatchStatus)
			auth.GET("/batch/list", batchHandler.ListBatchTasks)
		}
	}

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "project": "CopyCat MVP"})
	})

	return r
}
