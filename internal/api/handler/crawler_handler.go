package handler

import (
	"copycat/internal/core/agent"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
)

// CrawlerHandler 爬虫处理器
type CrawlerHandler struct {
	contentService *agent.ContentService
}

// NewCrawlerHandler 创建爬虫处理器
func NewCrawlerHandler(contentService *agent.ContentService) *CrawlerHandler {
	return &CrawlerHandler{contentService: contentService}
}

// CrawlRequest 爬取请求
type CrawlRequest struct {
	URL string `json:"url" binding:"required"`
}

// Crawl 爬取内容
// @Summary 爬取小红书/公众号内容
// @Tags Crawler
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CrawlRequest true "爬取请求"
// @Success 200 {object} response.Response
// @Router /api/v1/crawl [post]
func (h *CrawlerHandler) Crawl(c *gin.Context) {
	var req CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.contentService.CrawlOnly(c.Request.Context(), req.URL)
	if err != nil {
		response.ServerError(c, "crawl failed: "+err.Error())
		return
	}

	if !result.Success {
		response.BadRequest(c, result.Error)
		return
	}

	response.Success(c, agent.ToCrawlResponse(result))
}
