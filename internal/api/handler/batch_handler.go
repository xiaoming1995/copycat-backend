package handler

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"copycat/internal/core/agent"
	"copycat/internal/core/llm"
	"copycat/internal/model"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BatchHandler 批量任务处理器
type BatchHandler struct {
	db             *gorm.DB
	batchTaskRepo  *repository.BatchTaskRepository
	projectRepo    repository.ProjectRepository
	settingsRepo   *repository.UserSettingsRepository
	contentService *agent.ContentService
}

// NewBatchHandler 创建批量任务处理器
func NewBatchHandler(db *gorm.DB, contentService *agent.ContentService) *BatchHandler {
	return &BatchHandler{
		db:             db,
		batchTaskRepo:  repository.NewBatchTaskRepository(db),
		projectRepo:    repository.NewProjectRepository(db),
		settingsRepo:   repository.NewUserSettingsRepository(db),
		contentService: contentService,
	}
}

// BatchAnalyzeRequest 批量分析请求
type BatchAnalyzeRequest struct {
	URLs []string `json:"urls" binding:"required,min=1,max=10"`
}

// BatchAnalyzeResponse 批量分析响应
type BatchAnalyzeResponse struct {
	BatchID    string `json:"batch_id"`
	TotalCount int    `json:"total_count"`
	Status     string `json:"status"`
}

// BatchTaskStatusResponse 批量任务状态响应
type BatchTaskStatusResponse struct {
	BatchID      string                 `json:"batch_id"`
	TotalCount   int                    `json:"total_count"`
	SuccessCount int                    `json:"success_count"`
	FailedCount  int                    `json:"failed_count"`
	Status       string                 `json:"status"`
	Projects     []BatchProjectResponse `json:"projects,omitempty"`
}

// BatchProjectResponse 批量任务中的项目响应
type BatchProjectResponse struct {
	ID           string `json:"id"`
	SourceURL    string `json:"source_url"`
	Status       string `json:"status"`
	Title        string `json:"title,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CreateBatchAnalyze 创建批量分析任务
func (h *BatchHandler) CreateBatchAnalyze(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req BatchAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 检查用户 LLM 配置
	settings, err := h.settingsRepo.GetByUserID(userID)
	if err != nil || settings.LLMApiKey == "" {
		response.BadRequest(c, "请先在配置中心设置 LLM API Key")
		return
	}

	// 去重
	uniqueURLs := make([]string, 0)
	urlSet := make(map[string]bool)
	for _, url := range req.URLs {
		if url != "" && !urlSet[url] {
			urlSet[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}

	if len(uniqueURLs) == 0 {
		response.BadRequest(c, "请提供有效的链接")
		return
	}

	log.Printf("[Batch] 创建批量任务 - 用户ID: %d, 链接数: %d", userID, len(uniqueURLs))

	// 创建批量任务
	batchTask := &model.BatchTask{
		UserID:     userID,
		TotalCount: len(uniqueURLs),
		Status:     model.BatchTaskStatusProcessing,
	}

	if err := h.batchTaskRepo.Create(batchTask); err != nil {
		log.Printf("[Batch] 创建批量任务失败: %v", err)
		response.ServerError(c, "创建任务失败")
		return
	}

	log.Printf("[Batch] 批量任务已创建 - BatchID: %s", batchTask.ID.String())

	// 异步处理每个链接（传入 LLM 配置）
	go h.processBatchTask(batchTask.ID, userID, uniqueURLs, settings)

	response.Success(c, BatchAnalyzeResponse{
		BatchID:    batchTask.ID.String(),
		TotalCount: len(uniqueURLs),
		Status:     model.BatchTaskStatusProcessing,
	})
}

// processBatchTask 异步处理批量任务
func (h *BatchHandler) processBatchTask(batchID uuid.UUID, userID int64, urls []string, settings *model.UserSettings) {
	log.Printf("[Batch] 开始处理批量任务 - BatchID: %s, 链接数: %d", batchID.String(), len(urls))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 2) // 限制并发数为2（LLM 调用比较耗时）

	for _, url := range urls {
		wg.Add(1)
		go func(targetURL string) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			h.processSingleURL(batchID, userID, targetURL, settings)
		}(url)
	}

	wg.Wait()

	// 更新任务状态为完成
	task, err := h.batchTaskRepo.FindByID(batchID)
	if err != nil {
		log.Printf("[Batch] 查询任务失败: %v", err)
		return
	}

	if task.FailedCount == task.TotalCount {
		task.Status = model.BatchTaskStatusFailed
	} else {
		task.Status = model.BatchTaskStatusCompleted
	}

	if err := h.batchTaskRepo.Update(task); err != nil {
		log.Printf("[Batch] 更新任务状态失败: %v", err)
	}

	log.Printf("[Batch] 批量任务完成 - BatchID: %s, 成功: %d, 失败: %d",
		batchID.String(), task.SuccessCount, task.FailedCount)
}

// processSingleURL 处理单个URL（包含爬取、文本分析和图片分析）
func (h *BatchHandler) processSingleURL(batchID uuid.UUID, userID int64, url string, settings *model.UserSettings) {
	log.Printf("[Batch] 处理链接: %s", url)

	ctx := context.Background()

	// 1. 创建项目记录
	project := &model.Project{
		UserID:        userID,
		BatchTaskID:   &batchID,
		SourceURL:     url,
		SourceContent: "正在爬取内容...",
		Status:        model.ProjectStatusDraft,
	}

	if err := h.projectRepo.Create(ctx, project); err != nil {
		log.Printf("[Batch] 创建项目失败: %v", err)
		h.batchTaskRepo.IncrementFailedCount(batchID)
		return
	}

	// 2. 爬取内容
	result, err := h.contentService.CrawlOnly(ctx, url)
	if err != nil || !result.Success {
		log.Printf("[Batch] 爬取失败: %s - %v", url, err)
		project.Status = "failed"
		project.SourceContent = "爬取失败"
		h.projectRepo.Update(ctx, project)
		h.batchTaskRepo.IncrementFailedCount(batchID)
		return
	}

	// 3. 保存爬取结果（包括图片URL，以JSON格式保存供前端读取）
	var title, content string
	var images []string
	if result.Content != nil {
		title = result.Content.Title
		content = result.Content.Content
		images = result.Content.Images

		// 将内容和图片以 JSON 格式保存，供前端读取
		sourceData := map[string]interface{}{
			"title":   title,
			"content": content,
			"images":  images,
		}
		sourceJSON, _ := json.Marshal(sourceData)
		project.SourceContent = string(sourceJSON)
	}

	// 4. 调用文本 LLM 分析
	log.Printf("[Batch] 开始文本分析: %s", url)
	textClient := llm.NewClient(llm.Config{
		Provider: settings.LLMProvider,
		ApiKey:   settings.LLMApiKey,
		Model:    settings.LLMModel,
		BaseURL:  settings.LLMBaseURL,
	})

	analysisResult, err := textClient.AnalyzeContent(title, content)
	if err != nil {
		log.Printf("[Batch] 文本分析失败: %s - %v", url, err)
		project.Status = model.ProjectStatusDraft
		h.projectRepo.Update(ctx, project)
		h.batchTaskRepo.IncrementFailedCount(batchID)
		return
	}
	log.Printf("[Batch] 文本分析成功: %s (情绪: %s)", url, analysisResult.Emotion.Primary)

	// 5. 图片分析（如果有图片且配置了图片 LLM）
	var imageAnalysisResult *llm.ImageAnalysisResult
	if len(images) > 0 && settings.ImageLLMApiKey != "" {
		log.Printf("[Batch] 开始图片分析: %s (图片数: %d)", url, len(images))

		imageClient := llm.NewClient(llm.Config{
			Provider: settings.ImageLLMProvider,
			ApiKey:   settings.ImageLLMApiKey,
			Model:    settings.ImageLLMModel,
			BaseURL:  settings.ImageLLMBaseURL,
		})

		imageAnalysisResult, err = imageClient.AnalyzeImages(images)
		if err != nil {
			log.Printf("[Batch] 图片分析失败（继续保存文本分析）: %s - %v", url, err)
			// 图片分析失败不影响整体，继续保存文本分析结果
		} else {
			log.Printf("[Batch] 图片分析成功: %s (分析图片数: %d)", url, len(imageAnalysisResult.Images))
		}
	} else if len(images) > 0 {
		log.Printf("[Batch] 跳过图片分析（未配置图片 LLM）: %s", url)
	}

	// 6. 合并分析结果
	finalResult := map[string]interface{}{
		"emotion":        analysisResult.Emotion,
		"structure":      analysisResult.Structure,
		"keywords":       analysisResult.Keywords,
		"title_analysis": analysisResult.TitleAnalysis,
		"tone":           analysisResult.Tone,
		"word_count":     analysisResult.WordCount,
	}
	if imageAnalysisResult != nil {
		finalResult["image_analysis"] = imageAnalysisResult
	}

	analysisJSON, _ := json.Marshal(finalResult)
	project.AnalysisResult = analysisJSON
	project.Status = model.ProjectStatusAnalyzed

	if err := h.projectRepo.Update(ctx, project); err != nil {
		log.Printf("[Batch] 更新项目失败: %v", err)
		h.batchTaskRepo.IncrementFailedCount(batchID)
		return
	}

	h.batchTaskRepo.IncrementSuccessCount(batchID)
	log.Printf("[Batch] 链接处理完成: %s", url)
}

// GetBatchStatus 获取批量任务状态
func (h *BatchHandler) GetBatchStatus(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		response.BadRequest(c, "无效的批次ID")
		return
	}

	task, err := h.batchTaskRepo.FindByIDWithProjects(batchID)
	if err != nil {
		response.NotFound(c, "批次任务不存在")
		return
	}

	// 检查权限
	if task.UserID != userID {
		response.Forbidden(c, "无权查看此任务")
		return
	}

	// 构建项目列表
	projects := make([]BatchProjectResponse, 0)
	for _, p := range task.Projects {
		projects = append(projects, BatchProjectResponse{
			ID:        p.ID.String(),
			SourceURL: p.SourceURL,
			Status:    p.Status,
		})
	}

	response.Success(c, BatchTaskStatusResponse{
		BatchID:      task.ID.String(),
		TotalCount:   task.TotalCount,
		SuccessCount: task.SuccessCount,
		FailedCount:  task.FailedCount,
		Status:       task.Status,
		Projects:     projects,
	})
}

// ListBatchTasks 获取批量任务列表
func (h *BatchHandler) ListBatchTasks(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 50 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	tasks, total, err := h.batchTaskRepo.FindByUserID(userID, pageSize, offset)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}

	// 构建响应
	list := make([]BatchTaskStatusResponse, 0)
	for _, t := range tasks {
		list = append(list, BatchTaskStatusResponse{
			BatchID:      t.ID.String(),
			TotalCount:   t.TotalCount,
			SuccessCount: t.SuccessCount,
			FailedCount:  t.FailedCount,
			Status:       t.Status,
		})
	}

	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
