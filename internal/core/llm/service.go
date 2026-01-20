package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 提示词文件路径
const (
	PromptsDir         = "prompts"
	AnalyzePromptFile  = "analyze.txt"
	GeneratePromptFile = "generate.txt"
)

// 默认提示词模板，当文件不存在时使用
var defaultAnalyzePrompt = `你是一位爆款文案分析专家。请分析以下小红书/社交媒体文案的爆款逻辑。

标题：
{{title}}

文案内容：
{{content}}

请用 JSON 格式回复分析结果：
{
  "title_analysis": {
    "original": "原始标题",
    "hooks": ["钩子点1", "钩子点2"],
    "techniques": ["使用的技巧"],
    "score": 8
  },
  "emotion": {
    "primary": "主要情绪",
    "intensity": 0.8,
    "tags": ["情绪标签1", "情绪标签2", "情绪标签3"]
  },
  "structure": [
    {"title": "开篇", "description": "分析这部分内容"}
  ],
  "keywords": ["关键词1", "关键词2", "关键词3"],
  "tone": "语气风格",
  "word_count": 文案字数
}

请只回复 JSON，不要其他文字。`

var defaultGeneratePrompt = `你是一位爆款文案创作专家。请根据原文案分析和新主题，创作一篇仿写文案。

原标题：
{{title}}

原文案：
{{content}}

文案分析：
{{analysis}}

新主题：{{topic}}

请创作一篇新文案，保持原文的写作风格、结构和爆款逻辑。文案应包含标题和正文。

输出格式：
【标题】xxx
【正文】
xxx`

// TitleAnalysis 标题分析结果
type TitleAnalysis struct {
	Original   string   `json:"original"`   // 原始标题
	Hooks      []string `json:"hooks"`      // 钩子点
	Techniques []string `json:"techniques"` // 使用的技巧
	Score      int      `json:"score"`      // 评分
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	TitleAnalysis *TitleAnalysis  `json:"title_analysis,omitempty"`
	Emotion       EmotionAnalysis `json:"emotion"`
	Structure     []StructureItem `json:"structure"`
	Keywords      []string        `json:"keywords"`
	Tone          string          `json:"tone"`
	WordCount     int             `json:"word_count"`
}

// EmotionAnalysis 情绪分析
type EmotionAnalysis struct {
	Primary   string   `json:"primary"`   // 主要情绪
	Intensity float64  `json:"intensity"` // 情绪强度 0-1
	Tags      []string `json:"tags"`      // 情绪标签
}

// StructureItem 结构分析项
type StructureItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// loadPrompt 从文件加载提示词，文件不存在则使用默认值
func loadPrompt(filename string, defaultPrompt string) string {
	promptPath := filepath.Join(PromptsDir, filename)

	content, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[Prompt] 提示词文件不存在: %s，使用默认模板", promptPath)
		} else {
			log.Printf("[Prompt] 读取提示词文件失败: %v，使用默认模板", err)
		}
		return defaultPrompt
	}

	log.Printf("[Prompt] 已加载提示词文件: %s (%d 字符)", promptPath, len(content))
	return string(content)
}

// AnalyzeContent 分析爆款内容
func (c *Client) AnalyzeContent(title, content string) (*AnalysisResult, error) {
	log.Printf("[LLM Service] 开始分析内容 (标题: %d 字, 正文: %d 字)", len(title), len(content))

	if title != "" {
		log.Printf("   - 标题: %s", title)
	}
	contentPreview := content
	if len(contentPreview) > 100 {
		contentPreview = contentPreview[:100] + "..."
	}
	log.Printf("   - 内容预览: %s", strings.ReplaceAll(contentPreview, "\n", " "))

	// 从文件加载提示词模板
	promptTemplate := loadPrompt(AnalyzePromptFile, defaultAnalyzePrompt)

	// 替换占位符
	prompt := strings.ReplaceAll(promptTemplate, "{{title}}", title)
	prompt = strings.ReplaceAll(prompt, "{{content}}", content)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	response, err := c.Chat(messages)
	if err != nil {
		log.Printf("[LLM Service] 分析失败: %v", err)
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	// 尝试提取 JSON
	jsonStr := extractJSON(response)
	log.Printf("[LLM Service] 提取的 JSON (长度: %d):", len(jsonStr))
	jsonPreview := jsonStr
	if len(jsonPreview) > 500 {
		jsonPreview = jsonPreview[:500] + "..."
	}
	log.Printf("   %s", strings.ReplaceAll(jsonPreview, "\n", " "))

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[LLM Service] JSON 解析失败: %v", err)
		log.Printf("   原始响应: %s", response)
		return nil, fmt.Errorf("解析分析结果失败: %w, 原始响应: %s", err, response)
	}

	log.Printf("[LLM Service] 分析成功:")
	if result.TitleAnalysis != nil {
		log.Printf("   - 标题评分: %d/10", result.TitleAnalysis.Score)
		log.Printf("   - 标题技巧: %v", result.TitleAnalysis.Techniques)
	}
	log.Printf("   - 主要情绪: %s (强度: %.2f)", result.Emotion.Primary, result.Emotion.Intensity)
	log.Printf("   - 情绪标签: %v", result.Emotion.Tags)
	log.Printf("   - 结构段落: %d 个", len(result.Structure))
	log.Printf("   - 关键词: %v", result.Keywords)
	log.Printf("   - 语气风格: %s", result.Tone)
	log.Printf("   - 字数统计: %d", result.WordCount)

	return &result, nil
}

// GenerateContent 生成仿写文案
func (c *Client) GenerateContent(originalTitle, originalContent string, analysisResult *AnalysisResult, newTopic string) (string, error) {
	log.Printf("[LLM Service] 开始生成仿写文案")
	log.Printf("   - 新主题: %s", newTopic)
	log.Printf("   - 原标题: %s", originalTitle)
	log.Printf("   - 原文长度: %d 字", len(originalContent))

	analysisJSON, _ := json.MarshalIndent(analysisResult, "", "  ")

	// 从文件加载提示词模板
	promptTemplate := loadPrompt(GeneratePromptFile, defaultGeneratePrompt)

	// 替换占位符
	prompt := strings.ReplaceAll(promptTemplate, "{{title}}", originalTitle)
	prompt = strings.ReplaceAll(prompt, "{{content}}", originalContent)
	prompt = strings.ReplaceAll(prompt, "{{analysis}}", string(analysisJSON))
	prompt = strings.ReplaceAll(prompt, "{{topic}}", newTopic)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	response, err := c.Chat(messages)
	if err != nil {
		log.Printf("[LLM Service] 生成失败: %v", err)
		return "", fmt.Errorf("调用 LLM 失败: %w", err)
	}

	result := strings.TrimSpace(response)
	log.Printf("[LLM Service] 生成成功 (长度: %d 字符)", len(result))
	resultPreview := result
	if len(resultPreview) > 200 {
		resultPreview = resultPreview[:200] + "..."
	}
	log.Printf("   - 生成内容预览: %s", strings.ReplaceAll(resultPreview, "\n", " "))

	return result, nil
}

// extractJSON 从文本中提取 JSON
func extractJSON(text string) string {
	// 尝试找到 JSON 的起始和结束位置
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start != -1 && end != -1 && end > start {
		return text[start : end+1]
	}

	return text
}

// ========== 图片分析功能 ==========

const AnalyzeImagePromptFile = "analyze_image.txt"

var defaultAnalyzeImagePrompt = `你是一位视觉内容分析专家。请分析以下图片的视觉特点和创意技巧。

请用 JSON 格式回复分析结果：
{
  "images": [
    {
      "index": 1,
      "composition": "构图分析",
      "technique": "拍摄/设计技巧",
      "highlight": "视觉爆点",
      "color_tone": "色调风格",
      "mood": "表达的情绪/氛围"
    }
  ],
  "overall_style": "整体视觉风格总结",
  "visual_strategy": "视觉策略分析"
}

请只回复 JSON，不要其他文字。`

// ImageAnalysisItem 单张图片分析结果
type ImageAnalysisItem struct {
	Index       int    `json:"index"`
	Composition string `json:"composition"`  // 构图分析
	Technique   string `json:"technique"`    // 拍摄技巧
	Highlight   string `json:"highlight"`    // 视觉爆点
	ColorTone   string `json:"color_tone"`   // 色调风格
	Mood        string `json:"mood"`         // 情绪氛围
	ImagePrompt string `json:"image_prompt"` // AI 图像生成提示词
}

// ImageAnalysisResult 图片分析结果
type ImageAnalysisResult struct {
	Images         []ImageAnalysisItem `json:"images"`
	OverallStyle   string              `json:"overall_style"`   // 整体风格
	VisualStrategy string              `json:"visual_strategy"` // 视觉策略
}

// AnalyzeImages 分析图片内容（多模态）
func (c *Client) AnalyzeImages(imageURLs []string) (*ImageAnalysisResult, error) {
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("没有提供图片")
	}

	log.Printf("[LLM Service] 开始分析图片 (数量: %d)", len(imageURLs))
	for i, url := range imageURLs {
		log.Printf("   - 图片 %d: %s", i+1, url)
	}

	// 从文件加载提示词模板
	promptTemplate := loadPrompt(AnalyzeImagePromptFile, defaultAnalyzeImagePrompt)

	// 调用多模态接口
	response, err := c.ChatWithImages(promptTemplate, imageURLs)
	if err != nil {
		log.Printf("[LLM Service] 图片分析失败: %v", err)
		return nil, fmt.Errorf("调用多模态 LLM 失败: %w", err)
	}

	// 提取 JSON
	jsonStr := extractJSON(response)
	log.Printf("[LLM Service] 图片分析 JSON (长度: %d)", len(jsonStr))

	var result ImageAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[LLM Service] 图片分析 JSON 解析失败: %v", err)
		log.Printf("   原始响应: %s", response)
		return nil, fmt.Errorf("解析图片分析结果失败: %w", err)
	}

	log.Printf("[LLM Service] 图片分析成功:")
	log.Printf("   - 分析图片数量: %d", len(result.Images))
	log.Printf("   - 整体风格: %s", result.OverallStyle)
	log.Printf("   - 视觉策略: %s", result.VisualStrategy)

	return &result, nil
}
