package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"copycat/pkg/logger"
)

// 提示词文件路径
const (
	PromptsDir              = "prompts"
	AnalyzePromptFile       = "analyze.txt"
	GeneratePromptFile      = "generate.txt"
	GenerateVideoPromptFile = "generate_video.txt"
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
	Original    string   `json:"original"`               // 原始标题
	Hooks       []string `json:"hooks"`                  // 钩子点
	Techniques  []string `json:"techniques"`             // 使用的技巧
	Score       float64  `json:"score"`                  // 评分（支持小数）
	ScoreReason string   `json:"score_reason,omitempty"` // 评分理由
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	TitleAnalysis *TitleAnalysis  `json:"title_analysis,omitempty"`
	Emotion       EmotionAnalysis `json:"emotion"`
	Structure     []StructureItem `json:"structure"`
	Keywords      []string        `json:"keywords"`
	Tone          string          `json:"tone"`
	WordCount     int             `json:"word_count"`
	// 视频分析专属字段（支持新旧两种命名）
	Hook            *HookAnalysis            `json:"hook,omitempty"`
	HookStrategy    *HookAnalysis            `json:"hook_strategy,omitempty"` // 新命名
	GoldenQuotes    []string                 `json:"golden_quotes,omitempty"`
	Narrative       *NarrativeAnalysis       `json:"narrative,omitempty"`
	NarrativeLogic  *NarrativeLogicAnalysis  `json:"narrative_logic,omitempty"` // 新命名
	PPP             *PPPAnalysis             `json:"ppp,omitempty"`
	PPPModel        *PPPAnalysis             `json:"ppp_model,omitempty"` // 新命名
	Persona         *PersonaAnalysis         `json:"persona,omitempty"`
	ViralLogic      *ViralLogicAnalysis      `json:"viral_logic,omitempty"`
	ViralMechanics  *ViralMechanicsAnalysis  `json:"viral_mechanics,omitempty"` // 新命名
	Visual          *VisualAnalysis          `json:"visual,omitempty"`
	VisualDirection *VisualDirectionAnalysis `json:"visual_direction,omitempty"` // 新命名
	Audio           *AudioAnalysis           `json:"audio,omitempty"`
	AudioAtmosphere *AudioAtmosphereAnalysis `json:"audio_atmosphere,omitempty"` // 新命名
	TagsAndSEO      *TagsAndSEOAnalysis      `json:"tags_&_seo,omitempty"`       // 新字段
}

// HookAnalysis 开头钩子分析
type HookAnalysis struct {
	Type          string `json:"type"`          // 悬念式/冲突式/利益式/情绪式/反转式/提问式
	Description   string `json:"description"`   // 具体描述
	Duration      string `json:"duration"`      // 时长
	Effectiveness int    `json:"effectiveness"` // 有效性评分
}

// NarrativeAnalysis 叙事分析
type NarrativeAnalysis struct {
	Structure  string   `json:"structure"`  // 叙事结构
	Pacing     string   `json:"pacing"`     // 节奏
	Techniques []string `json:"techniques"` // 叙事技巧
}

// PPPAnalysis 人货场分析
type PPPAnalysis struct {
	People  string `json:"people"`  // 人物
	Place   string `json:"place"`   // 场景
	Product string `json:"product"` // 产品
}

// PersonaAnalysis 人设分析
type PersonaAnalysis struct {
	Type          string   `json:"type"`           // 人设类型
	Traits        []string `json:"traits"`         // 人设特点
	TrustBuilding string   `json:"trust_building"` // 信任建立方式
}

// ViralLogicAnalysis 爆款逻辑分析
type ViralLogicAnalysis struct {
	Core               string   `json:"core"`                // 核心逻辑
	Triggers           []string `json:"triggers"`            // 情绪触发点
	ReplicableElements []string `json:"replicable_elements"` // 可复用元素
}

// VisualAnalysis 视觉分析
type VisualAnalysis struct {
	Scenes         []string         `json:"scenes"`          // 场景描述
	Composition    string           `json:"composition"`     // 画面构图
	CameraMovement interface{}      `json:"camera_movement"` // 运镜手法（支持 string 或 []string）
	Editing        *EditingAnalysis `json:"editing"`         // 剪辑分析
	ColorTone      string           `json:"color_tone"`      // 色调
	Lighting       string           `json:"lighting"`        // 光线
}

// GetCameraMovementList 将 CameraMovement 统一转换为 []string
func (v *VisualAnalysis) GetCameraMovementList() []string {
	if v.CameraMovement == nil {
		return nil
	}

	// 如果是字符串，返回单元素数组
	if s, ok := v.CameraMovement.(string); ok {
		if s == "" {
			return nil
		}
		return []string{s}
	}

	// 如果是数组
	if arr, ok := v.CameraMovement.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return nil
}

// EditingAnalysis 剪辑分析
type EditingAnalysis struct {
	Style       string   `json:"style"`       // 剪辑风格
	Techniques  []string `json:"techniques"`  // 剪辑技巧
	Transitions []string `json:"transitions"` // 转场效果
}

// AudioAnalysis 音频分析
type AudioAnalysis struct {
	BGMStyle     string   `json:"bgm_style"`     // BGM风格
	BGMMatch     string   `json:"bgm_match"`     // BGM与内容匹配度
	VoiceStyle   string   `json:"voice_style"`   // 人声风格
	SoundEffects []string `json:"sound_effects"` // 音效
}

// NarrativeLogicAnalysis 新版叙事逻辑分析
type NarrativeLogicAnalysis struct {
	StructureType string   `json:"structure_type"` // 叙事结构
	Pacing        string   `json:"pacing"`         // 节奏
	GoldenQuotes  []string `json:"golden_quotes"`  // 金句
}

// ViralMechanicsAnalysis 新版爆款逻辑分析
type ViralMechanicsAnalysis struct {
	CoreLogic          string   `json:"core_logic"`          // 核心逻辑
	EmotionalTriggers  []string `json:"emotional_triggers"`  // 情绪触发点
	ReplicableElements []string `json:"replicable_elements"` // 可复用元素
}

// VisualDirectionAnalysis 新版视觉方向分析
type VisualDirectionAnalysis struct {
	SuggestedScenes          []string `json:"suggested_scenes"`           // 建议场景
	CompositionVibe          string   `json:"composition_vibe"`           // 构图风格
	CameraMovementSuggestion string   `json:"camera_movement_suggestion"` // 运镜建议
	EditingStyle             string   `json:"editing_style"`              // 剪辑风格
}

// AudioAtmosphereAnalysis 新版音频氛围分析
type AudioAtmosphereAnalysis struct {
	BGMStyle     string   `json:"bgm_style"`     // BGM风格
	VoiceTone    string   `json:"voice_tone"`    // 人声风格
	SoundEffects []string `json:"sound_effects"` // 音效
}

// TagsAndSEOAnalysis 标签和SEO分析
type TagsAndSEOAnalysis struct {
	Keywords         []string    `json:"keywords"`          // 关键词
	EmotionIntensity float64     `json:"emotion_intensity"` // 情绪强度
	WordCount        interface{} `json:"word_count"`        // 字数（支持 string 或 int）
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
			logger.LLMInfo("[Prompt] 提示词文件不存在: %s，使用默认模板", promptPath)
		} else {
			logger.LLMWarn("[Prompt] 读取提示词文件失败: %v，使用默认模板", err)
		}
		return defaultPrompt
	}

	logger.LLMInfo("[Prompt] 已加载提示词文件: %s (%d 字符)", promptPath, len(content))
	return string(content)
}

// AnalyzeContent 分析爆款内容
func (c *Client) AnalyzeContent(title, content string) (*AnalysisResult, error) {
	logger.LLMInfo("[LLM Service] 开始分析内容 (标题: %d 字, 正文: %d 字)", len(title), len(content))

	if title != "" {
		logger.LLMInfo("   - 标题: %s", title)
	}
	contentPreview := content
	if len(contentPreview) > 100 {
		contentPreview = contentPreview[:100] + "..."
	}
	logger.LLMInfo("   - 内容预览: %s", strings.ReplaceAll(contentPreview, "\n", " "))

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
		logger.LLMInfo("[LLM Service] 分析失败: %v", err)
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	// 尝试提取 JSON
	jsonStr := extractJSON(response)
	logger.LLMInfo("[LLM Service] 提取的 JSON (长度: %d):", len(jsonStr))
	jsonPreview := jsonStr
	if len(jsonPreview) > 500 {
		jsonPreview = jsonPreview[:500] + "..."
	}
	logger.LLMInfo("   %s", strings.ReplaceAll(jsonPreview, "\n", " "))

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.LLMInfo("[LLM Service] JSON 解析失败: %v", err)
		logger.LLMInfo("   原始响应: %s", response)
		return nil, fmt.Errorf("解析分析结果失败: %w, 原始响应: %s", err, response)
	}

	logger.LLMInfo("[LLM Service] 分析成功:")
	if result.TitleAnalysis != nil {
		logger.LLMInfo("   - 标题评分: %.1f/10", result.TitleAnalysis.Score)
		logger.LLMInfo("   - 标题技巧: %v", result.TitleAnalysis.Techniques)
	}
	logger.LLMInfo("   - 主要情绪: %s (强度: %.2f)", result.Emotion.Primary, result.Emotion.Intensity)
	logger.LLMInfo("   - 情绪标签: %v", result.Emotion.Tags)
	logger.LLMInfo("   - 结构段落: %d 个", len(result.Structure))
	logger.LLMInfo("   - 关键词: %v", result.Keywords)
	logger.LLMInfo("   - 语气风格: %s", result.Tone)
	logger.LLMInfo("   - 字数统计: %d", result.WordCount)

	return &result, nil
}

// AnalyzeVideoContent 视频内容分析（包含开头钩子、金句、叙事、人货场、人设、爆款逻辑）
func (c *Client) AnalyzeVideoContent(title, content string) (*AnalysisResult, error) {
	logger.LLMInfo("[LLM Service] 开始视频内容分析 (标题: %d 字, 正文: %d 字)", len(title), len(content))

	if title != "" {
		logger.LLMInfo("   - 标题: %s", title)
	}
	contentPreview := content
	if len(contentPreview) > 100 {
		contentPreview = contentPreview[:100] + "..."
	}
	logger.LLMInfo("   - 内容预览: %s", strings.ReplaceAll(contentPreview, "\n", " "))

	// 从文件加载视频分析提示词模板
	promptTemplate := loadPrompt("analyze_video.txt", defaultAnalyzePrompt)

	// 替换占位符
	prompt := strings.ReplaceAll(promptTemplate, "{{title}}", title)
	prompt = strings.ReplaceAll(prompt, "{{content}}", content)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	response, err := c.Chat(messages)
	if err != nil {
		logger.LLMInfo("[LLM Service] 视频分析失败: %v", err)
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	// 尝试提取 JSON
	jsonStr := extractJSON(response)
	logger.LLMInfo("[LLM Service] 提取的 JSON (长度: %d)", len(jsonStr))

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.LLMInfo("[LLM Service] JSON 解析失败: %v", err)
		logger.LLMInfo("   原始响应: %s", response)
		return nil, fmt.Errorf("解析分析结果失败: %w, 原始响应: %s", err, response)
	}

	logger.LLMInfo("[LLM Service] 视频分析成功:")
	if result.TitleAnalysis != nil {
		logger.LLMInfo("   - 标题评分: %.1f/10", result.TitleAnalysis.Score)
	}
	if result.Hook != nil {
		logger.LLMInfo("   - 开头钩子: %s (%s)", result.Hook.Type, result.Hook.Duration)
	}
	if len(result.GoldenQuotes) > 0 {
		logger.LLMInfo("   - 金句数量: %d", len(result.GoldenQuotes))
	}
	if result.Narrative != nil {
		logger.LLMInfo("   - 叙事结构: %s", result.Narrative.Structure)
	}
	if result.PPP != nil {
		logger.LLMInfo("   - 人货场: 人物=%s", result.PPP.People)
	}
	if result.Persona != nil {
		logger.LLMInfo("   - 人设类型: %s", result.Persona.Type)
	}
	if result.ViralLogic != nil {
		logger.LLMInfo("   - 爆款核心: %s", result.ViralLogic.Core)
	}

	return &result, nil
}

// GenerateContent 生成仿写文案
func (c *Client) GenerateContent(originalTitle, originalContent string, analysisResult *AnalysisResult, newTopic string) (string, error) {
	logger.LLMInfo("[LLM Service] 开始生成仿写文案")
	logger.LLMInfo("   - 新主题: %s", newTopic)
	logger.LLMInfo("   - 原标题: %s", originalTitle)
	logger.LLMInfo("   - 原文长度: %d 字", len(originalContent))

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
		logger.LLMInfo("[LLM Service] 生成失败: %v", err)
		return "", fmt.Errorf("调用 LLM 失败: %w", err)
	}

	result := strings.TrimSpace(response)
	logger.LLMInfo("[LLM Service] 生成成功 (长度: %d 字符)", len(result))
	resultPreview := result
	if len(resultPreview) > 200 {
		resultPreview = resultPreview[:200] + "..."
	}
	logger.LLMInfo("   - 生成内容预览: %s", strings.ReplaceAll(resultPreview, "\n", " "))

	return result, nil
}

// GenerateVideoScript 生成视频脚本仿写（包含时间线、分镜头、拍摄建议）
func (c *Client) GenerateVideoScript(originalTitle, originalContent string, analysisResult *AnalysisResult, newTopic string) (string, error) {
	logger.LLMInfo("[LLM Service] 开始生成视频脚本仿写")
	logger.LLMInfo("   - 新主题: %s", newTopic)
	logger.LLMInfo("   - 原标题: %s", originalTitle)
	logger.LLMInfo("   - 原文长度: %d 字", len(originalContent))

	analysisJSON, _ := json.MarshalIndent(analysisResult, "", "  ")

	// 从文件加载视频脚本生成提示词模板
	promptTemplate := loadPrompt(GenerateVideoPromptFile, defaultGeneratePrompt)

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
		logger.LLMInfo("[LLM Service] 视频脚本生成失败: %v", err)
		return "", fmt.Errorf("调用 LLM 失败: %w", err)
	}

	result := strings.TrimSpace(response)
	logger.LLMInfo("[LLM Service] 视频脚本生成成功 (长度: %d 字符)", len(result))
	resultPreview := result
	if len(resultPreview) > 300 {
		resultPreview = resultPreview[:300] + "..."
	}
	logger.LLMInfo("   - 生成脚本预览: %s", strings.ReplaceAll(resultPreview, "\n", " "))

	return result, nil
}

// GenerateMultipleVideoScripts 生成多条视频脚本仿写（逐条生成，避免超时）
func (c *Client) GenerateMultipleVideoScripts(originalTitle, originalContent string, analysisResult *AnalysisResult, newTopic string, count int) ([]string, error) {
	if count <= 0 {
		count = 1
	}

	logger.LLMInfo("[LLM Service] 开始生成多条视频脚本 (目标条数: %d)", count)
	logger.LLMInfo("   - 新主题: %s", newTopic)
	logger.LLMInfo("   - 原标题: %s", originalTitle)

	results := make([]string, 0, count)

	// 逐条生成，避免单次请求输出过多导致超时
	for i := 0; i < count; i++ {
		logger.LLMInfo("[LLM Service] 正在生成第 %d/%d 条视频脚本...", i+1, count)

		content, err := c.GenerateVideoScript(originalTitle, originalContent, analysisResult, newTopic)
		if err != nil {
			logger.LLMInfo("[LLM Service] 生成第 %d 条失败: %v", i+1, err)
			// 如果已经有结果了，返回已有的；否则返回错误
			if len(results) > 0 {
				logger.LLMInfo("[LLM Service] 返回已生成的 %d 条结果", len(results))
				return results, nil
			}
			return nil, err
		}
		results = append(results, content)
		logger.LLMInfo("[LLM Service] 成功生成第 %d/%d 条", i+1, count)
	}

	logger.LLMInfo("[LLM Service] 多条视频脚本生成完成 (实际条数: %d)", len(results))
	return results, nil
}

// GenerateMultipleContent 生成多条仿写文案
func (c *Client) GenerateMultipleContent(originalTitle, originalContent string, analysisResult *AnalysisResult, newTopic string, count int) ([]string, error) {
	if count <= 1 {
		// 单条生成走原逻辑
		content, err := c.GenerateContent(originalTitle, originalContent, analysisResult, newTopic)
		if err != nil {
			return nil, err
		}
		return []string{content}, nil
	}

	logger.LLMInfo("[LLM Service] 开始生成多条仿写文案 (条数: %d)", count)
	logger.LLMInfo("   - 新主题: %s", newTopic)
	logger.LLMInfo("   - 原标题: %s", originalTitle)

	analysisJSON, _ := json.MarshalIndent(analysisResult, "", "  ")

	// 从文件加载提示词模板
	promptTemplate := loadPrompt(GeneratePromptFile, defaultGeneratePrompt)

	// 替换占位符
	prompt := strings.ReplaceAll(promptTemplate, "{{title}}", originalTitle)
	prompt = strings.ReplaceAll(prompt, "{{content}}", originalContent)
	prompt = strings.ReplaceAll(prompt, "{{analysis}}", string(analysisJSON))
	prompt = strings.ReplaceAll(prompt, "{{topic}}", newTopic)

	// 添加多条生成的指令
	multiPrompt := fmt.Sprintf(`%s

【重要】请生成 %d 条不同风格的仿写文案，每条之间用以下分隔符分开：

===SEPARATOR===

确保每条文案都有独特的表达方式和切入角度，但都要围绕"%s"这个主题。`, prompt, count, newTopic)

	messages := []Message{
		{Role: "user", Content: multiPrompt},
	}

	response, err := c.Chat(messages)
	if err != nil {
		logger.LLMInfo("[LLM Service] 多条生成失败: %v", err)
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	// 分割多条内容
	parts := strings.Split(response, "===SEPARATOR===")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			results = append(results, trimmed)
		}
	}

	logger.LLMInfo("[LLM Service] 多条生成成功 (实际条数: %d)", len(results))
	return results, nil
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

	logger.LLMInfo("[LLM Service] 开始分析图片 (数量: %d)", len(imageURLs))
	for i, url := range imageURLs {
		logger.LLMInfo("   - 图片 %d: %s", i+1, url)
	}

	// 从文件加载提示词模板
	promptTemplate := loadPrompt(AnalyzeImagePromptFile, defaultAnalyzeImagePrompt)

	// 调用多模态接口
	response, err := c.ChatWithImages(promptTemplate, imageURLs)
	if err != nil {
		logger.LLMInfo("[LLM Service] 图片分析失败: %v", err)
		return nil, fmt.Errorf("调用多模态 LLM 失败: %w", err)
	}

	// 提取 JSON
	jsonStr := extractJSON(response)
	logger.LLMInfo("[LLM Service] 图片分析 JSON (长度: %d)", len(jsonStr))

	var result ImageAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.LLMInfo("[LLM Service] 图片分析 JSON 解析失败: %v", err)
		logger.LLMInfo("   原始响应: %s", response)
		return nil, fmt.Errorf("解析图片分析结果失败: %w", err)
	}

	logger.LLMInfo("[LLM Service] 图片分析成功:")
	logger.LLMInfo("   - 分析图片数量: %d", len(result.Images))
	logger.LLMInfo("   - 整体风格: %s", result.OverallStyle)
	logger.LLMInfo("   - 视觉策略: %s", result.VisualStrategy)

	return &result, nil
}
