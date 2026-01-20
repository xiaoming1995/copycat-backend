package crawler

import "time"

// Platform 平台类型
type Platform string

const (
	PlatformXiaohongshu Platform = "xiaohongshu" // 小红书
	PlatformWechat      Platform = "wechat"      // 微信公众号
	PlatformUnknown     Platform = "unknown"     // 未知平台
)

// CrawlResult 爬取结果
type CrawlResult struct {
	Success  bool         `json:"success"`
	Platform Platform     `json:"platform"`
	Content  *NoteContent `json:"content,omitempty"`
	Error    string       `json:"error,omitempty"`
}

// NoteContent 笔记内容
type NoteContent struct {
	// 基础信息
	NoteID   string `json:"note_id"`   // 笔记 ID
	Title    string `json:"title"`     // 标题
	Content  string `json:"content"`   // 正文内容
	Type     string `json:"type"`      // 类型: normal, video
	CoverURL string `json:"cover_url"` // 封面图 URL

	// 作者信息
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`

	// 媒体资源
	Images []string `json:"images"` // 图片列表
	Video  *Video   `json:"video,omitempty"`

	// 标签
	Tags []string `json:"tags"`

	// 统计数据
	LikeCount    int `json:"like_count"`
	CommentCount int `json:"comment_count"`
	CollectCount int `json:"collect_count"`
	ShareCount   int `json:"share_count"`

	// 时间信息
	PublishTime time.Time `json:"publish_time"`
	CrawlTime   time.Time `json:"crawl_time"`

	// 原始 URL
	SourceURL string `json:"source_url"`
}

// Video 视频信息
type Video struct {
	URL      string `json:"url"`
	Duration int    `json:"duration"` // 秒
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}
