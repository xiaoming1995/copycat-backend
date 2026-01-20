package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 件径
const (
	PromptsDir         = "prompts"
	AnalyzePromptFile  = "analyze.txt"
	GeneratePromptFile = "generate.txt"
)

// 默诈件存时使稉
var defaultAnalyzePrompt = `是爆款案分析家。分析红/媒案爆款逻辑。

题
{{title}}

案容
{{content}}

JSON 回分析
{
  "title_analysis": {
    "original": "题",
    "hooks": ["点1", "点2"],
    "techniques": ["使技"],
    "score": 8
  },
  "emotion": {
    "primary": "绪",
    "intensity": 0.8,
    "tags": ["绪签1", "绪签2", "绪签3"]
  },
  "structure": [
    {"title": "篇", "description": "分析这部分容"}
  ],
  "keywords": ["1", "2", "3"],
  "tone": "语风",
  "word_count": 案字数
}

回 JSON字。`

var defaultGeneratePrompt = `是爆款案创家。案分析题创篇仿写案。

题
{{title}}

案
{{content}}

案分析
{{analysis}}

题{{topic}}

创篇案写风构逻辑案含题正。

出
【题】xxx
【正】
xxx`

// TitleAnalysis 题分析
type TitleAnalysis struct {
	Original   string   `json:"original"`   // 题
	Hooks      []string `json:"hooks"`      // 点
	Techniques []string `json:"techniques"` // 使技
	Score      int      `json:"score"`      // 评分
}

// AnalysisResult 分析
type AnalysisResult struct {
	TitleAnalysis *TitleAnalysis  `json:"title_analysis,omitempty"`
	Emotion       EmotionAnalysis `json:"emotion"`
	Structure     []StructureItem `json:"structure"`
	Keywords      []string        `json:"keywords"`
	Tone          string          `json:"tone"`
	WordCount     int             `json:"word_count"`
}

// EmotionAnalysis 绪分析
type EmotionAnalysis struct {
	Primary   string   `json:"primary"`   // 绪
	Intensity float64  `json:"intensity"` // 绪0-1
	Tags      []string `json:"tags"`      // 绪签
}

// StructureItem 构分析项
type StructureItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// loadPrompt 从件载件存则使默
func loadPrompt(filename string, defaultPrompt string) string {
	promptPath := filepath.Join(PromptsDir, filename)

	content, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[Prompt] 件存: %s使默模", promptPath)
		} else {
			log.Printf("[Prompt] 读件: %v使默模", err)
		}
		return defaultPrompt
	}

	log.Printf("[Prompt] 载件: %s (%d 字)", promptPath, len(content))
	return string(content)
}

// AnalyzeContent 分析爆款容
func (c *Client) AnalyzeContent(title, content string) (*AnalysisResult, error) {
	log.Printf(" [LLM Service] 分析容 (题: %d 字, 正: %d 字)", len(title), len(content))

	if title != "" {
		log.Printf("   - 题: %s", title)
	}
	contentPreview := content
	if len(contentPreview) > 100 {
		contentPreview = contentPreview[:100] + "..."
	}
	log.Printf("   - 容预览: %s", strings.ReplaceAll(contentPreview, "\n", " "))

	// 从件载模
	promptTemplate := loadPrompt(AnalyzePromptFile, defaultAnalyzePrompt)

	// 替
	prompt := strings.ReplaceAll(promptTemplate, "{{title}}", title)
	prompt = strings.ReplaceAll(prompt, "{{content}}", content)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	response, err := c.Chat(messages)
	if err != nil {
		log.Printf("[LLM Service] 分析: %v", err)
		return nil, fmt.Errorf("调 LLM : %w", err)
	}

	// 试JSON
	jsonStr := extractJSON(response)
	log.Printf("[LLM Service]  JSON (长: %d):", len(jsonStr))
	jsonPreview := jsonStr
	if len(jsonPreview) > 500 {
		jsonPreview = jsonPreview[:500] + "..."
	}
	log.Printf("   %s", strings.ReplaceAll(jsonPreview, "\n", " "))

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[LLM Service] JSON 解析: %v", err)
		log.Printf("   : %s", response)
		return nil, fmt.Errorf("解析分析: %w, : %s", err, response)
	}

	log.Printf("[LLM Service] 分析成:")
	if result.TitleAnalysis != nil {
		log.Printf("   - 题评分: %d/10", result.TitleAnalysis.Score)
		log.Printf("   - 题技: %v", result.TitleAnalysis.Techniques)
	}
	log.Printf("   - 绪: %s (: %.2f)", result.Emotion.Primary, result.Emotion.Intensity)
	log.Printf("   - 绪签: %v", result.Emotion.Tags)
	log.Printf("   - 构段落: %d ", len(result.Structure))
	log.Printf("   - : %v", result.Keywords)
	log.Printf("   - 语风: %s", result.Tone)
	log.Printf("   - 字数计: %d", result.WordCount)

	return &result, nil
}

// GenerateContent 成仿写案
func (c *Client) GenerateContent(originalTitle, originalContent string, analysisResult *AnalysisResult, newTopic string) (string, error) {
	log.Printf("[LLM Service] 成仿写案")
	log.Printf("   - 题: %s", newTopic)
	log.Printf("   - 题: %s", originalTitle)
	log.Printf("   - 长: %d 字", len(originalContent))

	analysisJSON, _ := json.MarshalIndent(analysisResult, "", "  ")

	// 从件载模
	promptTemplate := loadPrompt(GeneratePromptFile, defaultGeneratePrompt)

	// 替
	prompt := strings.ReplaceAll(promptTemplate, "{{title}}", originalTitle)
	prompt = strings.ReplaceAll(prompt, "{{content}}", originalContent)
	prompt = strings.ReplaceAll(prompt, "{{analysis}}", string(analysisJSON))
	prompt = strings.ReplaceAll(prompt, "{{topic}}", newTopic)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	response, err := c.Chat(messages)
	if err != nil {
		log.Printf("[LLM Service] 成: %v", err)
		return "", fmt.Errorf("调 LLM : %w", err)
	}

	result := strings.TrimSpace(response)
	log.Printf("[LLM Service] 成成(长: %d 字)", len(result))
	resultPreview := result
	if len(resultPreview) > 200 {
		resultPreview = resultPreview[:200] + "..."
	}
	log.Printf("   - 成容预览: %s", strings.ReplaceAll(resultPreview, "\n", " "))

	return result, nil
}

// extractJSON 从JSON
func extractJSON(text string) string {
	// 试找到 JSON 置
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start != -1 && end != -1 && end > start {
		return text[start : end+1]
	}

	return text
}

// ========== 图片分析 ==========

const AnalyzeImagePromptFile = "analyze_image.txt"

var defaultAnalyzeImagePrompt = `是视觉容分析家。分析图片视觉特点创。

JSON 回分析
{
  "images": [
    {
      "index": 1,
      "composition": "构图分析",
      "technique": "摄/设计",
      "highlight": "视觉爆点",
      "color_tone": "色调风",
      "mood": "达绪/氛围"
    }
  ],
  "overall_style": "整视觉风总",
  "visual_strategy": "视觉分析"
}

回 JSON字。`

// ImageAnalysisItem 图片分析
type ImageAnalysisItem struct {
	Index       int    `json:"index"`
	Composition string `json:"composition"`  // 构图分析
	Technique   string `json:"technique"`    // 摄
	Highlight   string `json:"highlight"`    // 视觉爆点
	ColorTone   string `json:"color_tone"`   // 色调风
	Mood        string `json:"mood"`         // 绪氛围
	ImagePrompt string `json:"image_prompt"` // AI图成
}

// ImageAnalysisResult 图片分析
type ImageAnalysisResult struct {
	Images         []ImageAnalysisItem `json:"images"`
	OverallStyle   string              `json:"overall_style"`   // 整风
	VisualStrategy string              `json:"visual_strategy"` // 视觉
}

// AnalyzeImages 分析图片容模态
func (c *Client) AnalyzeImages(imageURLs []string) (*ImageAnalysisResult, error) {
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("没供图片")
	}

	log.Printf("[LLM Service] 分析图片 (数: %d)", len(imageURLs))
	for i, url := range imageURLs {
		log.Printf("   - 图片 %d: %s", i+1, url)
	}

	// 从件载模
	promptTemplate := loadPrompt(AnalyzeImagePromptFile, defaultAnalyzeImagePrompt)

	// 调模态
	response, err := c.ChatWithImages(promptTemplate, imageURLs)
	if err != nil {
		log.Printf("[LLM Service] 图片分析: %v", err)
		return nil, fmt.Errorf("调模态 LLM : %w", err)
	}

	// JSON
	jsonStr := extractJSON(response)
	log.Printf("[LLM Service] 图片分析 JSON (长: %d)", len(jsonStr))

	var result ImageAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[LLM Service] 图片分析 JSON 解析: %v", err)
		log.Printf("   : %s", response)
		return nil, fmt.Errorf("解析图片分析: %w", err)
	}

	log.Printf("[LLM Service] 图片分析成:")
	log.Printf("   - 分析图片数: %d", len(result.Images))
	log.Printf("   - 整风: %s", result.OverallStyle)
	log.Printf("   - 视觉: %s", result.VisualStrategy)

	return &result, nil
}
