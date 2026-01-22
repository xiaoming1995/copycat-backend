package handler

import (
	"errors"

	"copycat/internal/model"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectRepo repository.ProjectRepository
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectRepo repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo}
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	SourceURL     string `json:"source_url"`
	SourceContent string `json:"source_content" binding:"required"`
	ContentType   string `json:"content_type"` // text/video/images
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	SourceURL        string `json:"source_url"`
	SourceContent    string `json:"source_content"`
	ContentType      string `json:"content_type"`
	NewTopic         string `json:"new_topic"`
	GeneratedContent string `json:"generated_content"`
	Status           string `json:"status"`
}

// ListProjectsRequest 列表请求
type ListProjectsRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// Create 创建项目
// @Summary 创建项目
// @Tags Project
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "项目信息"
// @Success 200 {object} response.Response{data=model.Project}
// @Router /api/v1/projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")

	// 设置默认 content_type
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text"
	}

	project := &model.Project{
		UserID:        userID.(int64),
		SourceURL:     req.SourceURL,
		SourceContent: req.SourceContent,
		ContentType:   contentType,
		Status:        model.ProjectStatusDraft,
	}

	if err := h.projectRepo.Create(c.Request.Context(), project); err != nil {
		response.ServerError(c, "failed to create project")
		return
	}

	response.Success(c, project)
}

// List 获取项目列表
// @Summary 获取项目列表
// @Tags Project
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	var req ListProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	userID, _ := c.Get("userID")

	projects, total, err := h.projectRepo.GetByUserID(c.Request.Context(), userID.(int64), req.Page, req.PageSize)
	if err != nil {
		response.ServerError(c, "failed to get projects")
		return
	}

	response.SuccessWithPage(c, projects, total, req.Page, req.PageSize)
}

// Get 获取项目详情
// @Summary 获取项目详情
// @Tags Project
// @Security BearerAuth
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} response.Response{data=model.Project}
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.ServerError(c, "failed to get project")
		return
	}

	// 验证项目所有权
	userID, _ := c.Get("userID")
	if project.UserID != userID.(int64) {
		response.Forbidden(c, "access denied")
		return
	}

	response.Success(c, project)
}

// Update 更新项目
// @Summary 更新项目
// @Tags Project
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param request body UpdateProjectRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.Project}
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.ServerError(c, "failed to get project")
		return
	}

	// 验证项目所有权
	userID, _ := c.Get("userID")
	if project.UserID != userID.(int64) {
		response.Forbidden(c, "access denied")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 更新字段
	if req.SourceURL != "" {
		project.SourceURL = req.SourceURL
	}
	if req.SourceContent != "" {
		project.SourceContent = req.SourceContent
	}
	if req.NewTopic != "" {
		project.NewTopic = req.NewTopic
	}
	if req.GeneratedContent != "" {
		project.GeneratedContent = req.GeneratedContent
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	if err := h.projectRepo.Update(c.Request.Context(), project); err != nil {
		response.ServerError(c, "failed to update project")
		return
	}

	response.Success(c, project)
}

// Delete 删除项目
// @Summary 删除项目
// @Tags Project
// @Security BearerAuth
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} response.Response
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.ServerError(c, "failed to get project")
		return
	}

	// 验证项目所有权
	userID, _ := c.Get("userID")
	if project.UserID != userID.(int64) {
		response.Forbidden(c, "access denied")
		return
	}

	if err := h.projectRepo.Delete(c.Request.Context(), id); err != nil {
		response.ServerError(c, "failed to delete project")
		return
	}

	response.SuccessWithMessage(c, "project deleted", nil)
}

// GetByURL 根据 URL 查询已有的项目（检查是否已分析）
// @Summary 根据 URL 查询项目
// @Tags Project
// @Security BearerAuth
// @Produce json
// @Param url query string true "来源URL"
// @Success 200 {object} response.Response{data=model.Project}
// @Router /api/v1/projects/check [get]
func (h *ProjectHandler) GetByURL(c *gin.Context) {
	sourceURL := c.Query("url")
	if sourceURL == "" {
		response.BadRequest(c, "url is required")
		return
	}

	userID, _ := c.Get("userID")

	project, err := h.projectRepo.GetBySourceURL(c.Request.Context(), userID.(int64), sourceURL)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有找到已分析的项目，返回 null
			response.Success(c, nil)
			return
		}
		response.ServerError(c, "failed to get project")
		return
	}

	response.Success(c, project)
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// BatchDelete 批量删除项目
// @Summary 批量删除项目
// @Tags Project
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body BatchDeleteRequest true "批量删除请求"
// @Success 200 {object} response.Response
// @Router /api/v1/projects/batch [delete]
func (h *ProjectHandler) BatchDelete(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "user not found")
		return
	}

	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "ids is required")
		return
	}

	// 批量删除
	deletedCount := 0
	for _, idStr := range req.IDs {
		projectID, err := uuid.Parse(idStr)
		if err != nil {
			continue // 跳过无效的 ID
		}

		// 验证项目属于当前用户
		project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
		if err != nil {
			continue
		}
		if project.UserID != userID.(int64) {
			continue
		}

		// 删除
		if err := h.projectRepo.Delete(c.Request.Context(), projectID); err == nil {
			deletedCount++
		}
	}

	response.Success(c, gin.H{
		"deleted_count": deletedCount,
		"message":       "批量删除成功",
	})
}
