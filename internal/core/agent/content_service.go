package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"copycat/internal/core/crawler"
	"copycat/internal/repository"

	"gorm.io/datatypes"
)

// ContentService 内容服务
type ContentService struct {
	crawlerManager *crawler.CrawlerManager
	projectRepo    repository.ProjectRepository
}

// NewContentService 创建内容服务
func NewContentService(projectRepo repository.ProjectRepository) *ContentService {
	return &ContentService{
		crawlerManager: crawler.NewCrawlerManager(),
		projectRepo:    projectRepo,
	}
}

// CrawlAndSave 爬取内容并保存到项目
func (s *ContentService) CrawlAndSave(ctx context.Context, projectID interface{}, url string) (*crawler.CrawlResult, error) {
	// 爬取内容
	result, err := s.crawlerManager.Crawl(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("crawl failed: %w", err)
	}

	if !result.Success {
		return result, nil
	}

	// 如果项目 ID 有效，更新项目内容
	if projectID != nil {
		if err := s.updateProject(ctx, projectID, result); err != nil {
			return nil, fmt.Errorf("failed to update project: %w", err)
		}
	}

	return result, nil
}

// CrawlOnly 仅爬取内容，不保存
func (s *ContentService) CrawlOnly(ctx context.Context, url string) (*crawler.CrawlResult, error) {
	return s.crawlerManager.Crawl(ctx, url)
}

// updateProject 更新项目内容
func (s *ContentService) updateProject(ctx context.Context, projectID interface{}, result *crawler.CrawlResult) error {
	// 这里需要根据实际的 projectID 类型进行处理
	// 由于 Project 使用 UUID，这里简化处理
	if result.Content == nil {
		return nil
	}

	// 将爬取结果转换为 JSON 存储到 analysis_result 字段
	// 这里先将原始内容存储，后续 LLM 分析后会更新
	contentJSON, err := json.Marshal(map[string]interface{}{
		"crawled_content": result.Content,
		"platform":        result.Platform,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	// 构建更新数据
	_ = datatypes.JSON(contentJSON)
	// 实际更新逻辑需要根据 projectID 类型实现
	// 这里仅作为示例

	return nil
}

// CrawlRequest 爬取请求
type CrawlRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// CrawlResponse 爬取响应
type CrawlResponse struct {
	Success  bool                 `json:"success"`
	Platform string               `json:"platform"`
	Content  *crawler.NoteContent `json:"content,omitempty"`
	Error    string               `json:"error,omitempty"`
}

// ToCrawlResponse 将爬取结果转换为响应格式
func ToCrawlResponse(r *crawler.CrawlResult) *CrawlResponse {
	resp := &CrawlResponse{
		Success:  r.Success,
		Platform: string(r.Platform),
		Error:    r.Error,
	}
	if r.Content != nil {
		resp.Content = r.Content
	}
	return resp
}
