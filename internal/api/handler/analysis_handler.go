package handler

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"copycat/internal/core/llm"
	"copycat/internal/model"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AnalysisHandler 分析理器
type AnalysisHandler struct {
	db           *gorm.DB
	settingsRepo *repository.UserSettingsRepository
	projectRepo  repository.ProjectRepository
}

// NewAnalysisHandler 创建分析理器
func NewAnalysisHandler(db *gorm.DB) *AnalysisHandler {
	return &AnalysisHandler{
		db:           db,
		settingsRepo: repository.NewUserSettingsRepository(db),
		projectRepo:  repository.NewProjectRepository(db),
	}
}

// AnalyzeRequest 分析求
type AnalyzeRequest struct {
	Title       string `json:"title"`                      // 题
	Content     string `json:"content" binding:"required"` // 正容
	ProjectID   string `json:"project_id"`                 // 选项目
	ContentType string `json:"content_type"`               // 内容类型: text/images/video
}

// AnalyzeImagesRequest 图片分析求
type AnalyzeImagesRequest struct {
	Images    []string `json:"images" binding:"required"` // 图片 URL 列表
	ProjectID string   `json:"project_id"`                // 选项目
}

// GenerateRequest 成求
type GenerateRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	NewTopic  string `json:"new_topic" binding:"required"`
}

// Analyze 分析容
// @Summary 分析爆款容
// @Tags Analysis
// @Security BearerAuth
// @Param request body AnalyzeRequest true "分析求"
// @Success 200 {object} response.Response{data=llm.AnalysisResult}
// @Router /analyze [post]
func (h *AnalysisHandler) Analyze(c *gin.Context) {
	log.Printf("[API] 到分析求")

	userID := c.GetInt64("userID")
	if userID == 0 {
		log.Printf("[API] 登录")
		response.Unauthorized(c, "登录")
		return
	}
	log.Printf("   - ID: %d", userID)

	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[API] 数解析: %v", err)
		response.BadRequest(c, "数误: "+err.Error())
		return
	}
	log.Printf("   - 容长: %d 字", len(req.Content))
	log.Printf("   - 项目: %s", req.ProjectID)

	//  LLM 置
	settings, err := h.settingsRepo.GetByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		log.Printf("[API] 置 LLM")
		response.BadRequest(c, "置心设置 LLM API Key")
		return
	}
	if err != nil {
		log.Printf("[API] 置: %v", err)
		response.ServerError(c, "置")
		return
	}

	if settings.LLMApiKey == "" {
		log.Printf("[API] API Key 空")
		response.BadRequest(c, "置心设置 LLM API Key")
		return
	}

	log.Printf("[API] 到置:")
	log.Printf("   - Provider: %s", settings.LLMProvider)
	log.Printf("   - Model: %s", settings.LLMModel)
	log.Printf("   - BaseURL: %s", settings.LLMBaseURL)

	// 创建 LLM 客端
	client := llm.NewClient(llm.Config{
		Provider: settings.LLMProvider,
		ApiKey:   settings.LLMApiKey,
		Model:    settings.LLMModel,
		BaseURL:  settings.LLMBaseURL,
	})

	// 调分析
	log.Printf("[API] 调 LLM 分析...")
	log.Printf("   - 内容类型: %s", req.ContentType)

	var result *llm.AnalysisResult

	// 根据内容类型调用不同的分析方法
	if req.ContentType == "video" {
		log.Printf("[API] 使用视频专属分析方法")
		result, err = client.AnalyzeVideoContent(req.Title, req.Content)
	} else {
		result, err = client.AnalyzeContent(req.Title, req.Content)
	}

	if err != nil {
		log.Printf("[API] 分析: %v", err)
		response.ServerError(c, "分析: "+err.Error())
		return
	}

	log.Printf("[API] 分析成绪: %s", result.Emotion.Primary)

	// 项目更项目分析
	if req.ProjectID != "" {
		analysisJSON, _ := json.Marshal(result)
		projectUUID, _ := uuid.Parse(req.ProjectID)
		project, _ := h.projectRepo.GetByID(context.Background(), projectUUID)
		if project != nil {
			project.AnalysisResult = analysisJSON
			project.Status = model.ProjectStatusAnalyzed
			h.projectRepo.Update(context.Background(), project)
			log.Printf("[API] 更项目分析: %s", req.ProjectID)
		}
	}

	response.Success(c, result)
}

// Generate 成仿写容
// @Summary 成仿写案
// @Tags Analysis
// @Security BearerAuth
// @Param request body GenerateRequest true "成求"
// @Success 200 {object} response.Response{data=string}
// @Router /generate [post]
func (h *AnalysisHandler) Generate(c *gin.Context) {
	log.Printf("[API] 到成求")

	userID := c.GetInt64("userID")
	if userID == 0 {
		log.Printf("[API] 登录")
		response.Unauthorized(c, "登录")
		return
	}
	log.Printf("   - ID: %d", userID)

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[API] 数解析: %v", err)
		response.BadRequest(c, "数误: "+err.Error())
		return
	}
	log.Printf("   - 项目ID: %s", req.ProjectID)
	log.Printf("   - 题: %s", req.NewTopic)

	// 项目信息
	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		log.Printf("[API] 效项目ID: %s", req.ProjectID)
		response.BadRequest(c, "效项目ID")
		return
	}

	project, err := h.projectRepo.GetByID(context.Background(), projectUUID)
	if err != nil {
		log.Printf("[API] 项目存: %v", err)
		response.NotFound(c, "项目存")
		return
	}

	// 证项目所
	if project.UserID != userID {
		log.Printf("[API] 项目")
		response.Forbidden(c, "项目")
		return
	}

	//  LLM 置
	settings, err := h.settingsRepo.GetByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		log.Printf("[API] 置 LLM")
		response.BadRequest(c, "置心设置 LLM API Key")
		return
	}
	if err != nil {
		log.Printf("[API] 置: %v", err)
		response.ServerError(c, "置")
		return
	}

	if settings.LLMApiKey == "" {
		log.Printf("[API] API Key 空")
		response.BadRequest(c, "置心设置 LLM API Key")
		return
	}

	log.Printf("[API] 到置:")
	log.Printf("   - Provider: %s", settings.LLMProvider)
	log.Printf("   - Model: %s", settings.LLMModel)

	// 解析分析
	var analysisResult llm.AnalysisResult
	if project.AnalysisResult != nil {
		if err := json.Unmarshal(project.AnalysisResult, &analysisResult); err != nil {
			log.Printf("[API] 解析分析: %v", err)
			response.BadRequest(c, "项目分析进分析")
			return
		}
		log.Printf("[API] 载项目分析成")
	} else {
		log.Printf("[API] 项目没分析")
		response.BadRequest(c, "项目分析进分析")
		return
	}

	// 创建 LLM 客端
	client := llm.NewClient(llm.Config{
		Provider: settings.LLMProvider,
		ApiKey:   settings.LLMApiKey,
		Model:    settings.LLMModel,
		BaseURL:  settings.LLMBaseURL,
	})

	// 调成
	log.Printf("[API] 调 LLM 成...")
	// 从分析题
	originalTitle := ""
	if analysisResult.TitleAnalysis != nil {
		originalTitle = analysisResult.TitleAnalysis.Original
	}

	// 获取生成条数配置（默认1条）
	generateCount := settings.GenerateCount
	if generateCount <= 0 {
		generateCount = 1
	}
	if generateCount > 10 {
		generateCount = 10 // 最多10条
	}
	log.Printf("   - 生成条数: %d", generateCount)
	log.Printf("   - 内容类型: %s", project.ContentType)

	var generatedContents []string

	// 根据内容类型选择不同的生成方法
	if project.ContentType == "video" {
		// 视频类型：生成视频脚本（包含时间线、分镜头、拍摄建议）
		log.Printf("[API] 使用视频脚本生成方法")
		generatedContents, err = client.GenerateMultipleVideoScripts(originalTitle, project.SourceContent, &analysisResult, req.NewTopic, generateCount)
		if err != nil {
			log.Printf("[API] 视频脚本生成失败: %v", err)
			response.ServerError(c, "生成失败: "+err.Error())
			return
		}
	} else {
		// 图文类型：使用原有的仿写生成方法
		log.Printf("[API] 使用图文仿写生成方法")
		generatedContents, err = client.GenerateMultipleContent(originalTitle, project.SourceContent, &analysisResult, req.NewTopic, generateCount)
		if err != nil {
			log.Printf("[API] 生成失败: %v", err)
			response.ServerError(c, "生成失败: "+err.Error())
			return
		}
	}

	log.Printf("[API] 生成成功，内容条数: %d", len(generatedContents))

	// 更新项目（保存第一条或全部内容）
	project.NewTopic = req.NewTopic
	if len(generatedContents) > 0 {
		project.GeneratedContent = strings.Join(generatedContents, "\n\n===分隔符===\n\n")
	}
	project.Status = model.ProjectStatusCompleted
	h.projectRepo.Update(context.Background(), project)
	log.Printf("[API] 更新项目成功")

	response.Success(c, gin.H{
		"generated_contents": generatedContents,
		"generated_content":  generatedContents[0], // 兼容旧版本
	})
}

// AnalyzeImages 分析图片容
// @Summary 分析图片容模态
// @Tags Analysis
// @Security BearerAuth
// @Param request body AnalyzeImagesRequest true "图片分析求"
// @Success 200 {object} response.Response{data=llm.ImageAnalysisResult}
// @Router /analyze-images [post]
func (h *AnalysisHandler) AnalyzeImages(c *gin.Context) {
	log.Printf("[API] 到图片分析求")

	userID := c.GetInt64("userID")
	if userID == 0 {
		log.Printf("[API] 登录")
		response.Unauthorized(c, "登录")
		return
	}

	var req AnalyzeImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[API] 求数误: %v", err)
		response.BadRequest(c, "求数误: "+err.Error())
		return
	}

	if len(req.Images) == 0 {
		log.Printf("[API] 没供图片")
		response.BadRequest(c, "供少图片")
		return
	}

	log.Printf(" [API] 图片分析求数:")
	log.Printf("   - 图片数: %d", len(req.Images))
	log.Printf("   - 项目 ID: %s", req.ProjectID)

	// LLM 置使图片分析置
	settings, err := h.settingsRepo.GetByUserID(userID)
	if err != nil || settings.ImageLLMApiKey == "" {
		log.Printf("[API] 置图片分析 LLM进图片分析")
		response.BadRequest(c, "置心设置图片分析模 API Key")
		return
	}

	// 检是模态模
	multimodalModels := map[string]bool{
		"gpt-4o":                          true,
		"gpt-4-turbo":                     true,
		"qwen-vl-max":                     true,
		"qwen-vl-plus":                    true,
		"glm-4v":                          true,
		"deepseek-vl":                     true,
		"claude-3-5-sonnet-20240620":      true,
		"claude-3-opus-20240229":          true,
		"kimi-latest":                     true, // Kimi 版视觉
		"moonshot-v1-8k-vision-preview":   true,
		"moonshot-v1-32k-vision-preview":  true,
		"moonshot-v1-128k-vision-preview": true,
	}

	if !multimodalModels[settings.ImageLLMModel] {
		log.Printf("[API] 图片模%s 模态", settings.ImageLLMModel)
		response.BadRequest(c, "图片模图片分析切到模态模垈 gpt-4o, qwen-vl-max, glm-4v")
		return
	}

	// 创建 LLM 客端使图片分析置
	client := llm.NewClient(llm.Config{
		Provider: settings.ImageLLMProvider,
		ApiKey:   settings.ImageLLMApiKey,
		Model:    settings.ImageLLMModel,
		BaseURL:  settings.ImageLLMBaseURL,
	})

	// 调图片分析
	log.Printf("[API] 调 LLM 分析图片...")
	result, err := client.AnalyzeImages(req.Images)
	if err != nil {
		log.Printf("[API] 图片分析: %v", err)
		response.ServerError(c, "图片分析: "+err.Error())
		return
	}

	log.Printf("[API] 图片分析成分析了 %d 图片", len(result.Images))

	// 项目将图片分析合并到项目分析
	if req.ProjectID != "" {
		projectUUID, _ := uuid.Parse(req.ProjectID)
		project, _ := h.projectRepo.GetByID(context.Background(), projectUUID)
		if project != nil && project.AnalysisResult != nil {
			// 解析现分析
			var existingResult map[string]interface{}
			if err := json.Unmarshal(project.AnalysisResult, &existingResult); err == nil {
				// 图片分析
				existingResult["image_analysis"] = result
				updatedJSON, _ := json.Marshal(existingResult)
				project.AnalysisResult = updatedJSON
				h.projectRepo.Update(context.Background(), project)
				log.Printf("[API] 将图片分析存到项目: %s", req.ProjectID)
			}
		}
	}

	response.Success(c, result)
}
