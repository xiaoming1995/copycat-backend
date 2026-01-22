package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// XHSCrawler 小红书爬虫
type XHSCrawler struct {
	client *http.Client
}

// NewXHSCrawler 创建小红书爬虫实例
func NewXHSCrawler() *XHSCrawler {
	return &XHSCrawler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Crawl 爬取小红书笔记
func (c *XHSCrawler) Crawl(ctx context.Context, url string) (*CrawlResult, error) {
	// 解析笔记 ID
	noteID, err := c.parseNoteID(url)
	if err != nil {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformXiaohongshu,
			Error:    err.Error(),
		}, nil
	}

	// 构建请求
	req, err := c.buildRequest(ctx, url)
	if err != nil {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformXiaohongshu,
			Error:    fmt.Sprintf("failed to build request: %v", err),
		}, nil
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformXiaohongshu,
			Error:    fmt.Sprintf("failed to fetch page: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformXiaohongshu,
			Error:    fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}, nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformXiaohongshu,
			Error:    fmt.Sprintf("failed to read response: %v", err),
		}, nil
	}

	// 解析页面内容
	content, err := c.parseContent(string(body), noteID, url)
	if err != nil {
		return &CrawlResult{
			Success:  false,
			Platform: PlatformXiaohongshu,
			Error:    fmt.Sprintf("failed to parse content: %v", err),
		}, nil
	}

	return &CrawlResult{
		Success:  true,
		Platform: PlatformXiaohongshu,
		Content:  content,
	}, nil
}

// parseNoteID 从 URL 解析笔记 ID
func (c *XHSCrawler) parseNoteID(url string) (string, error) {
	// 支持多种 URL 格式:
	// https://www.xiaohongshu.com/explore/xxxxx
	// https://www.xiaohongshu.com/discovery/item/xxxxx
	// https://xhslink.com/xxxxx (短链接)

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`xiaohongshu\.com/explore/([a-zA-Z0-9]+)`),
		regexp.MustCompile(`xiaohongshu\.com/discovery/item/([a-zA-Z0-9]+)`),
		regexp.MustCompile(`xhslink\.com/([a-zA-Z0-9]+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(url)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("invalid xiaohongshu url: %s", url)
}

// buildRequest 构建 HTTP 请求
func (c *XHSCrawler) buildRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 模拟浏览器请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	return req, nil
}

// parseContent 解析页面内容
func (c *XHSCrawler) parseContent(html, noteID, sourceURL string) (*NoteContent, error) {
	content := &NoteContent{
		NoteID:    noteID,
		SourceURL: sourceURL,
		CrawlTime: time.Now(),
		Images:    []string{},
		Tags:      []string{},
	}

	log.Printf("[XHS Crawler] 解析笔记: %s, HTML长度: %d", noteID, len(html))

	// 方法1: 尝试从页面内嵌的 JSON 数据中提取
	if err := c.extractFromJSON(html, content); err == nil && content.Content != "" {
		log.Printf("[XHS Crawler] JSON提取成功: 标题=%s, 内容长度=%d", content.Title, len(content.Content))
		return content, nil
	} else if err != nil {
		log.Printf("[XHS Crawler] JSON提取失败: %v", err)
	}

	// 方法2: 使用正则表达式提取
	if err := c.extractFromHTML(html, content); err == nil && content.Content != "" {
		log.Printf("[XHS Crawler] HTML提取成功: 标题=%s, 内容长度=%d", content.Title, len(content.Content))
		return content, nil
	} else if err != nil {
		log.Printf("[XHS Crawler] HTML提取失败: %v", err)
	}

	// 如果都失败，返回基础信息
	if content.Title == "" && content.Content == "" {
		log.Printf("[XHS Crawler] 提取失败: 标题和内容都为空")
		// 记录 HTML 的前 500 个字符用于调试
		if len(html) > 500 {
			log.Printf("[XHS Crawler] HTML 开头内容: %s...", html[:500])
		} else {
			log.Printf("[XHS Crawler] HTML 全部内容: %s", html)
		}
		return nil, fmt.Errorf("failed to extract content from page")
	}

	return content, nil
}

// extractFromJSON 从页面内嵌 JSON 提取数据
func (c *XHSCrawler) extractFromJSON(html string, content *NoteContent) error {
	// 小红书页面通常在 script 标签中嵌入 JSON 数据
	// 格式: window.__INITIAL_STATE__ = {...}
	pattern := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*(\{.+?\})\s*;?\s*</script>`)
	matches := pattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		return fmt.Errorf("no initial state found")
	}

	jsonStr := matches[1]
	// 处理 Unicode 转义
	jsonStr = strings.ReplaceAll(jsonStr, "undefined", "null")

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	// 尝试从 note 对象提取数据
	if noteData, ok := c.findNoteData(data); ok {
		c.extractNoteFields(noteData, content)
	}

	return nil
}

// findNoteData 在嵌套数据中查找笔记数据
func (c *XHSCrawler) findNoteData(data map[string]interface{}) (map[string]interface{}, bool) {
	// 常见路径: note.noteDetailMap, note.note
	if note, ok := data["note"].(map[string]interface{}); ok {
		if noteDetail, ok := note["noteDetailMap"].(map[string]interface{}); ok {
			for _, v := range noteDetail {
				if noteMap, ok := v.(map[string]interface{}); ok {
					if noteObj, ok := noteMap["note"].(map[string]interface{}); ok {
						return noteObj, true
					}
				}
			}
		}
		return note, true
	}
	return nil, false
}

// extractNoteFields 提取笔记字段
func (c *XHSCrawler) extractNoteFields(data map[string]interface{}, content *NoteContent) {
	if title, ok := data["title"].(string); ok {
		content.Title = title
	}
	if desc, ok := data["desc"].(string); ok {
		content.Content = desc
	}
	if noteType, ok := data["type"].(string); ok {
		content.Type = noteType
	}

	// 提取图片
	if imageList, ok := data["imageList"].([]interface{}); ok {
		for _, img := range imageList {
			if imgMap, ok := img.(map[string]interface{}); ok {
				if urlDefault, ok := imgMap["urlDefault"].(string); ok {
					content.Images = append(content.Images, urlDefault)
				} else if url, ok := imgMap["url"].(string); ok {
					content.Images = append(content.Images, url)
				}
			}
		}
	}

	// 提取标签
	if tagList, ok := data["tagList"].([]interface{}); ok {
		for _, tag := range tagList {
			if tagMap, ok := tag.(map[string]interface{}); ok {
				if name, ok := tagMap["name"].(string); ok {
					content.Tags = append(content.Tags, name)
				}
			}
		}
	}

	// 提取作者信息
	if user, ok := data["user"].(map[string]interface{}); ok {
		if nickname, ok := user["nickname"].(string); ok {
			content.AuthorName = nickname
		}
		if userId, ok := user["userId"].(string); ok {
			content.AuthorID = userId
		}
		if avatar, ok := user["avatar"].(string); ok {
			content.AuthorAvatar = avatar
		}
	}

	// 提取互动数据
	if interactInfo, ok := data["interactInfo"].(map[string]interface{}); ok {
		if likedCount, ok := interactInfo["likedCount"].(string); ok {
			content.LikeCount = c.parseCount(likedCount)
		}
		if collectedCount, ok := interactInfo["collectedCount"].(string); ok {
			content.CollectCount = c.parseCount(collectedCount)
		}
		if commentCount, ok := interactInfo["commentCount"].(string); ok {
			content.CommentCount = c.parseCount(commentCount)
		}
		if shareCount, ok := interactInfo["shareCount"].(string); ok {
			content.ShareCount = c.parseCount(shareCount)
		}
	}

	// 提取封面图
	if cover, ok := data["cover"].(map[string]interface{}); ok {
		if urlDefault, ok := cover["urlDefault"].(string); ok {
			content.CoverURL = urlDefault
		} else if url, ok := cover["url"].(string); ok {
			content.CoverURL = url
		}
	}

	// 提取视频信息（针对视频类型笔记）
	if content.Type == "video" {
		if video, ok := data["video"].(map[string]interface{}); ok {
			videoInfo := &Video{}
			// 尝试多种可能的视频 URL 字段
			if media, ok := video["media"].(map[string]interface{}); ok {
				if stream, ok := media["stream"].(map[string]interface{}); ok {
					if h264, ok := stream["h264"].([]interface{}); ok && len(h264) > 0 {
						if firstStream, ok := h264[0].(map[string]interface{}); ok {
							if masterUrl, ok := firstStream["masterUrl"].(string); ok {
								videoInfo.URL = masterUrl
							}
						}
					}
				}
			}
			// 备用：直接获取视频 URL
			if videoInfo.URL == "" {
				if url, ok := video["url"].(string); ok {
					videoInfo.URL = url
				} else if firstFrameUrl, ok := video["firstFrameUrl"].(string); ok {
					content.CoverURL = firstFrameUrl
				}
			}
			// 获取视频时长
			if duration, ok := video["duration"].(float64); ok {
				videoInfo.Duration = int(duration)
			}
			// 获取视频尺寸
			if width, ok := video["width"].(float64); ok {
				videoInfo.Width = int(width)
			}
			if height, ok := video["height"].(float64); ok {
				videoInfo.Height = int(height)
			}
			if videoInfo.URL != "" || content.CoverURL != "" {
				content.Video = videoInfo
			}
		}
	}
}

// extractFromHTML 从 HTML 中使用正则提取数据
func (c *XHSCrawler) extractFromHTML(html string, content *NoteContent) error {
	// 提取标题 (og:title meta tag)
	titlePattern := regexp.MustCompile(`<meta\s+(?:property|name)="og:title"\s+content="([^"]+)"`)
	if matches := titlePattern.FindStringSubmatch(html); len(matches) >= 2 {
		content.Title = matches[1]
	}

	// 提取描述 (og:description meta tag)
	descPattern := regexp.MustCompile(`<meta\s+(?:property|name)="og:description"\s+content="([^"]+)"`)
	if matches := descPattern.FindStringSubmatch(html); len(matches) >= 2 {
		content.Content = matches[1]
	}

	// 提取图片 (og:image meta tag)
	imgPattern := regexp.MustCompile(`<meta\s+(?:property|name)="og:image"\s+content="([^"]+)"`)
	if matches := imgPattern.FindStringSubmatch(html); len(matches) >= 2 {
		content.Images = append(content.Images, matches[1])
	}

	return nil
}

// parseCount 解析数量字符串 (如 "1.2万" -> 12000)
func (c *XHSCrawler) parseCount(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	multiplier := 1
	if strings.Contains(s, "万") {
		multiplier = 10000
		s = strings.ReplaceAll(s, "万", "")
	} else if strings.Contains(s, "千") {
		multiplier = 1000
		s = strings.ReplaceAll(s, "千", "")
	}

	var num float64
	fmt.Sscanf(s, "%f", &num)
	return int(num * float64(multiplier))
}

// Platform 返回平台类型
func (c *XHSCrawler) Platform() Platform {
	return PlatformXiaohongshu
}
