package crawler

import (
	"context"
	"fmt"
	"strings"
)

// Crawler 爬虫接口
type Crawler interface {
	Crawl(ctx context.Context, url string) (*CrawlResult, error)
	Platform() Platform
}

// CrawlerManager 爬虫管理器
type CrawlerManager struct {
	crawlers map[Platform]Crawler
}

// NewCrawlerManager 创建爬虫管理器
func NewCrawlerManager() *CrawlerManager {
	manager := &CrawlerManager{
		crawlers: make(map[Platform]Crawler),
	}

	// 注册小红书爬虫
	xhsCrawler := NewXHSCrawler()
	manager.Register(xhsCrawler)

	return manager
}

// Register 注册爬虫
func (m *CrawlerManager) Register(crawler Crawler) {
	m.crawlers[crawler.Platform()] = crawler
}

// Crawl 根据 URL 自动选择爬虫进行爬取
func (m *CrawlerManager) Crawl(ctx context.Context, url string) (*CrawlResult, error) {
	platform := m.detectPlatform(url)
	if platform == PlatformUnknown {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformUnknown,
			Error:    "unsupported platform",
		}, nil
	}

	crawler, ok := m.crawlers[platform]
	if !ok {
		return &CrawlResult{
			Success:  false,
			Platform: platform,
			Error:    fmt.Sprintf("no crawler registered for platform: %s", platform),
		}, nil
	}

	return crawler.Crawl(ctx, url)
}

// detectPlatform 根据 URL 检测平台
func (m *CrawlerManager) detectPlatform(url string) Platform {
	url = strings.ToLower(url)

	if strings.Contains(url, "xiaohongshu.com") || strings.Contains(url, "xhslink.com") {
		return PlatformXiaohongshu
	}

	if strings.Contains(url, "mp.weixin.qq.com") {
		return PlatformWechat
	}

	return PlatformUnknown
}

// GetCrawler 获取指定平台的爬虫
func (m *CrawlerManager) GetCrawler(platform Platform) (Crawler, bool) {
	crawler, ok := m.crawlers[platform]
	return crawler, ok
}
