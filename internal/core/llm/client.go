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

// Config LLM 置
type Config struct {
	Provider string // openai, deepseek, anthropic
	ApiKey   string
	Model    string
	BaseURL  string // 选自义 API 
}

// Client LLM 客端
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient 创建 LLM 客端
func NewClient(config Config) *Client {
	log.Printf("[LLM] 创建客端 - Provider: %s, Model: %s, BaseURL: %s, ApiKey: %s",
		config.Provider, config.Model, config.BaseURL, maskApiKey(config.ApiKey))
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 增时时间到 3 分推理模较
		},
	}
}

// Message 消息构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse 聊庈 DeepSeek R1 
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content          string `json:"content"`           // 准容
			ReasoningContent string `json:"reasoning_content"` // DeepSeek R1 推理容
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

// Chat 送聊求
func (c *Client) Chat(messages []Message) (string, error) {
	baseURL := c.getBaseURL()
	endpoint := baseURL + "/chat/completions"

	reqBody := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4000, // 增 token 数
	}

	// DeepSeek Reasoner 模 temperature 数
	if strings.Contains(c.config.Model, "reasoner") {
		reqBody.Temperature = 0
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[LLM] 列求: %v", err)
		return "", fmt.Errorf("列求: %w", err)
	}

	// 志求信息
	log.Printf("[LLM] 送求:")
	log.Printf("   - Endpoint: %s", endpoint)
	log.Printf("   - Model: %s", c.config.Model)
	log.Printf("   - Messages: %d ", len(messages))
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
		log.Printf("[LLM] 创建求: %v", err)
		return "", fmt.Errorf("创建求: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)

	log.Printf("[LLM] 等...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[LLM] 求(耗时 %.2fs): %v", time.Since(startTime).Seconds(), err)
		return "", fmt.Errorf("求: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[LLM] 读: %v", err)
		return "", fmt.Errorf("读: %w", err)
	}

	elapsed := time.Since(startTime)

	// 志信息
	log.Printf("[LLM] 到(耗时 %.2fs):", elapsed.Seconds())
	log.Printf("   - HTTP Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("   - Response Size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LLM] HTTP 误: %s", string(body))
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("[LLM] 解析: %v, : %s", err, string(body))
		return "", fmt.Errorf("解析: %w", err)
	}

	if chatResp.Error != nil {
		log.Printf("[LLM] API 误: %s (type: %s, code: %s)",
			chatResp.Error.Message, chatResp.Error.Type, chatResp.Error.Code)
		return "", fmt.Errorf("API 误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		log.Printf("[LLM] 没回")
		return "", fmt.Errorf("没回")
	}

	// 使 content空则使 reasoning_contentDeepSeek R1
	content := chatResp.Choices[0].Message.Content
	if content == "" && chatResp.Choices[0].Message.ReasoningContent != "" {
		log.Printf("[LLM] 使 reasoning_content (DeepSeek R1 模)")
		content = chatResp.Choices[0].Message.ReasoningContent
	}

	if content == "" {
		log.Printf("[LLM] content reasoning_content 都空")
		log.Printf("   - : %s", string(body))
		return "", fmt.Errorf("LLM 回容空")
	}

	// 志成信息
	log.Printf("[LLM] 调成:")
	if chatResp.Usage != nil {
		log.Printf("   - Token 使: prompt=%d, completion=%d, total=%d",
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
	}
	contentPreview := content
	if len(contentPreview) > 300 {
		contentPreview = contentPreview[:300] + "..."
	}
	log.Printf("   - 容预览: %s", strings.ReplaceAll(contentPreview, "\n", " "))

	return content, nil
}

// getBaseURL API 础 URL
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

// maskApiKey 脱API Key
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		if len(apiKey) == 0 {
			return "(空)"
		}
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// downloadImageAsBase64 载图片并转 Base64 于 Kimi 等 Base64  API
func downloadImageAsBase64(imageURL string) (string, error) {
	log.Printf("[LLM] 载图片: %s", imageURL)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("载图片: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("载图片 HTTP %d", resp.StatusCode)
	}

	// 读图片数
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读图片数: %w", err)
	}

	// 检图片类
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(imageData)
	}

	// 简Content-Type
	mimeType := "image/jpeg" // 默
	if strings.Contains(contentType, "png") {
		mimeType = "image/png"
	} else if strings.Contains(contentType, "gif") {
		mimeType = "image/gif"
	} else if strings.Contains(contentType, "webp") {
		mimeType = "image/webp"
	}

	// 转 Base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 回 Data URL 
	result := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	log.Printf("[LLM] 图片转 Base64 (类: %s, : %d bytes)", mimeType, len(imageData))

	return result, nil
}

// ========== 模态 ==========

// ContentPart 模态消息容部分
type ContentPart struct {
	Type     string    `json:"type"`                // "text" "image_url"
	Text     string    `json:"text,omitempty"`      // 容
	ImageURL *ImageURL `json:"image_url,omitempty"` // 图片 URL
}

// ImageURL 图片 URL 构
type ImageURL struct {
	URL    string `json:"url"`              // 图片 URL
	Detail string `json:"detail,omitempty"` // 细: "low", "high", "auto"
}

// MultimodalMessage 模态消息
type MultimodalMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// MultimodalChatRequest 模态聊求
type MultimodalChatRequest struct {
	Model       string              `json:"model"`
	Messages    []MultimodalMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

// ChatWithImages 送含图片模态求
func (c *Client) ChatWithImages(text string, imageURLs []string) (string, error) {
	baseURL := c.getBaseURL()
	endpoint := baseURL + "/chat/completions"

	// 构建模态消息容
	contentParts := []ContentPart{
		{Type: "text", Text: text},
	}

	for _, url := range imageURLs {
		var imageContent string

		// Kimi/Moonshot 使 Base64 
		if c.config.Provider == "moonshot" {
			base64Data, err := downloadImageAsBase64(url)
			if err != nil {
				log.Printf("[LLM] 载图片: %v过图片", err)
				continue
			}
			// Kimi : data:image/jpeg;base64,<base64_data>
			imageContent = base64Data
		} else {
			// OpenAI 等供商直使 URL
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
		log.Printf("[LLM] 列模态求: %v", err)
		return "", fmt.Errorf("列求: %w", err)
	}

	log.Printf("[LLM] 送模态求:")
	log.Printf("   - Endpoint: %s", endpoint)
	log.Printf("   - Model: %s", c.config.Model)
	log.Printf("   - 图片数: %d", len(imageURLs))
	textPreview := text
	if len(textPreview) > 200 {
		textPreview = textPreview[:200] + "..."
	}
	log.Printf("   - 容: %s", strings.ReplaceAll(textPreview, "\n", " "))

	startTime := time.Now()

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建求: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)

	log.Printf("[LLM] 等模态...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[LLM] 模态求(耗时 %.2fs): %v", time.Since(startTime).Seconds(), err)
		return "", fmt.Errorf("求: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("[LLM] 到模态(耗时 %.2fs):", elapsed.Seconds())
	log.Printf("   - HTTP Status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LLM] HTTP 误: %s", string(body))
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析: %w", err)
	}

	if chatResp.Error != nil {
		log.Printf("[LLM] API 误: %s", chatResp.Error.Message)
		return "", fmt.Errorf("API 误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("没回")
	}

	content := chatResp.Choices[0].Message.Content
	if content == "" && chatResp.Choices[0].Message.ReasoningContent != "" {
		content = chatResp.Choices[0].Message.ReasoningContent
	}

	log.Printf("[LLM] 模态调成:")
	if chatResp.Usage != nil {
		log.Printf("   - Token 使: prompt=%d, completion=%d, total=%d",
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
	}

	return content, nil
}
