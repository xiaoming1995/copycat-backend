package llm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Config LLM 配置
type Config struct {
	Provider string // openai, deepseek, anthropic
	ApiKey   string
	Model    string
	BaseURL  string // 可选自定义 API 地址
}

// Client LLM 客户端
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient 创建 LLM 客户端
func NewClient(config Config) *Client {
	log.Printf("[LLM] 创建客户端 - Provider: %s, Model: %s, BaseURL: %s, ApiKey: %s",
		config.Provider, config.Model, config.BaseURL, maskApiKey(config.ApiKey))
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 增加超时时间到 3 分钟，推理模型较慢
		},
	}
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse 聊天响应，支持 DeepSeek R1 格式
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content          string `json:"content"`           // 标准内容
			ReasoningContent string `json:"reasoning_content"` // DeepSeek R1 推理内容
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// Chat 发送聊天请求
func (c *Client) Chat(messages []Message) (string, error) {
	baseURL := c.getBaseURL()
	endpoint := baseURL + "/chat/completions"

	reqBody := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4000, // 增加 token 数
	}

	// DeepSeek Reasoner 模型不支持 temperature 参数
	if strings.Contains(c.config.Model, "reasoner") {
		reqBody.Temperature = 0
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[LLM] 序列化请求失败: %v", err)
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 日志请求信息
	log.Printf("[LLM] 发送请求:")
	log.Printf("   - Endpoint: %s", endpoint)
	log.Printf("   - Model: %s", c.config.Model)
	log.Printf("   - Messages: %d 条", len(messages))
	for i, msg := range messages {
		contentPreview := msg.Content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		log.Printf("   - [%d] %s: %s", i+1, msg.Role, contentPreview)
	}

	startTime := time.Now()

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[LLM] 创建请求失败: %v", err)
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)

	log.Printf("[LLM] 等待响应...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[LLM] 请求失败 (耗时 %.2fs): %v", time.Since(startTime).Seconds(), err)
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[LLM] 读取响应失败: %v", err)
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	elapsed := time.Since(startTime)

	// 日志响应信息
	log.Printf("[LLM] 收到响应 (耗时 %.2fs):", elapsed.Seconds())
	log.Printf("   - HTTP Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("   - Response Size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LLM] HTTP 错误响应: %s", string(body))
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("[LLM] 解析响应失败: %v, 原始响应: %s", err, string(body))
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if chatResp.Error != nil {
		log.Printf("[LLM] API 错误: %s (type: %s, code: %s)",
			chatResp.Error.Message, chatResp.Error.Type, chatResp.Error.Code)
		return "", fmt.Errorf("API 错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		log.Printf("[LLM] 没有返回结果")
		return "", fmt.Errorf("没有返回结果")
	}

	// 优先使用 content，如果为空则使用 reasoning_content（DeepSeek R1）
	content := chatResp.Choices[0].Message.Content
	if content == "" && chatResp.Choices[0].Message.ReasoningContent != "" {
		log.Printf("[LLM] 使用 reasoning_content (DeepSeek R1 模式)")
		content = chatResp.Choices[0].Message.ReasoningContent
	}

	if content == "" {
		log.Printf("[LLM] content 和 reasoning_content 都为空")
		log.Printf("   - 原始响应: %s", string(body))
		return "", fmt.Errorf("LLM 返回内容为空")
	}

	// 日志成功信息
	log.Printf("[LLM] 调用成功:")
	if chatResp.Usage != nil {
		log.Printf("   - Token 使用: prompt=%d, completion=%d, total=%d",
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
	}
	contentPreview := content
	if len(contentPreview) > 300 {
		contentPreview = contentPreview[:300] + "..."
	}
	log.Printf("   - 内容预览: %s", strings.ReplaceAll(contentPreview, "\n", " "))

	return content, nil
}

// getBaseURL 获取 API 基础 URL
func (c *Client) getBaseURL() string {
	if c.config.BaseURL != "" {
		return c.config.BaseURL
	}

	switch c.config.Provider {
	case "deepseek":
		return "https://api.deepseek.com/v1"
	case "anthropic":
		return "https://api.anthropic.com/v1"
	default: // openai
		return "https://api.openai.com/v1"
	}
}

// maskApiKey 脱敏 API Key
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		if len(apiKey) == 0 {
			return "(空)"
		}
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// downloadImageAsBase64 下载图片并转换为 Base64，用于 Kimi 等需要 Base64 格式的 API
func downloadImageAsBase64(imageURL string) (string, error) {
	log.Printf("[LLM] 下载图片: %s", imageURL)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败 HTTP %d", resp.StatusCode)
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %w", err)
	}

	// 检测图片类型
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(imageData)
	}

	// 简化 Content-Type
	mimeType := "image/jpeg" // 默认
	if strings.Contains(contentType, "png") {
		mimeType = "image/png"
	} else if strings.Contains(contentType, "gif") {
		mimeType = "image/gif"
	} else if strings.Contains(contentType, "webp") {
		mimeType = "image/webp"
	}

	// 转换为 Base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 返回 Data URL 格式
	result := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	log.Printf("[LLM] 图片转换为 Base64 成功 (类型: %s, 大小: %d bytes)", mimeType, len(imageData))

	return result, nil
}

// ========== 多模态支持 ==========

// ContentPart 多模态消息内容部分
type ContentPart struct {
	Type     string    `json:"type"`                // "text" 或 "image_url"
	Text     string    `json:"text,omitempty"`      // 文本内容
	ImageURL *ImageURL `json:"image_url,omitempty"` // 图片 URL
}

// ImageURL 图片 URL 结构
type ImageURL struct {
	URL    string `json:"url"`              // 图片 URL
	Detail string `json:"detail,omitempty"` // 详细程度: "low", "high", "auto"
}

// MultimodalMessage 多模态消息
type MultimodalMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// MultimodalChatRequest 多模态聊天请求
type MultimodalChatRequest struct {
	Model       string              `json:"model"`
	Messages    []MultimodalMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

// ChatWithImages 发送含图片的多模态请求
func (c *Client) ChatWithImages(text string, imageURLs []string) (string, error) {
	baseURL := c.getBaseURL()
	endpoint := baseURL + "/chat/completions"

	// 构建多模态消息内容
	contentParts := []ContentPart{
		{Type: "text", Text: text},
	}

	for _, url := range imageURLs {
		var imageContent string

		// Kimi/Moonshot 需要使用 Base64 格式
		if c.config.Provider == "moonshot" {
			base64Data, err := downloadImageAsBase64(url)
			if err != nil {
				log.Printf("[LLM] 下载图片失败: %v，跳过该图片", err)
				continue
			}
			// Kimi 格式: data:image/jpeg;base64,<base64_data>
			imageContent = base64Data
		} else {
			// OpenAI 等供应商直接使用 URL
			imageContent = url
		}

		contentParts = append(contentParts, ContentPart{
			Type: "image_url",
			ImageURL: &ImageURL{
				URL:    imageContent,
				Detail: "auto",
			},
		})
	}

	reqBody := MultimodalChatRequest{
		Model: c.config.Model,
		Messages: []MultimodalMessage{
			{Role: "user", Content: contentParts},
		},
		Temperature: 0.7,
		MaxTokens:   4000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[LLM] 序列化多模态请求失败: %v", err)
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	log.Printf("[LLM] 发送多模态请求:")
	log.Printf("   - Endpoint: %s", endpoint)
	log.Printf("   - Model: %s", c.config.Model)
	log.Printf("   - 图片数量: %d", len(imageURLs))
	textPreview := text
	if len(textPreview) > 200 {
		textPreview = textPreview[:200] + "..."
	}
	log.Printf("   - 文本内容: %s", strings.ReplaceAll(textPreview, "\n", " "))

	startTime := time.Now()

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)

	log.Printf("[LLM] 等待多模态响应...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[LLM] 多模态请求失败 (耗时 %.2fs): %v", time.Since(startTime).Seconds(), err)
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("[LLM] 收到多模态响应 (耗时 %.2fs):", elapsed.Seconds())
	log.Printf("   - HTTP Status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LLM] HTTP 错误响应: %s", string(body))
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if chatResp.Error != nil {
		log.Printf("[LLM] API 错误: %s", chatResp.Error.Message)
		return "", fmt.Errorf("API 错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("没有返回结果")
	}

	content := chatResp.Choices[0].Message.Content
	if content == "" && chatResp.Choices[0].Message.ReasoningContent != "" {
		content = chatResp.Choices[0].Message.ReasoningContent
	}

	log.Printf("[LLM] 多模态调用成功:")
	if chatResp.Usage != nil {
		log.Printf("   - Token 使用: prompt=%d, completion=%d, total=%d",
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
	}

	return content, nil
}
