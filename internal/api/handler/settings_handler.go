package handler

import (
	"copycat/internal/model"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SettingsHandler 设置处理器
type SettingsHandler struct {
	settingsRepo *repository.UserSettingsRepository
}

// NewSettingsHandler 创建设置处理器
func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: repository.NewUserSettingsRepository(db),
	}
}

// LLMConfigItem 单个 LLM 配置
type LLMConfigItem struct {
	Provider  string `json:"provider"`
	ApiKey    string `json:"api_key"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	BatchSize int    `json:"batch_size,omitempty"`
}

// ProviderApiKeys 各提供商 API Key
type ProviderApiKeys struct {
	OpenAI    string `json:"openai"`
	DeepSeek  string `json:"deepseek"`
	Moonshot  string `json:"moonshot"`
	Qwen      string `json:"qwen"`
	Hunyuan   string `json:"hunyuan"`
	Doubao    string `json:"doubao"`
	Zhipu     string `json:"zhipu"`
	Anthropic string `json:"anthropic"`
}

// MultiModalConfigResponse 多模态配置响应
type MultiModalConfigResponse struct {
	ContentAnalysis LLMConfigItem   `json:"content_analysis"`
	ImageAnalysis   LLMConfigItem   `json:"image_analysis"`
	VideoAnalysis   LLMConfigItem   `json:"video_analysis"`
	ProviderKeys    ProviderApiKeys `json:"provider_keys"`
	GenerateCount   int             `json:"generate_count"`
	DefaultTaskType string          `json:"default_task_type"`
}

// === 请求结构体 ===

// SaveApiConfigRequest 保存 API 配置请求（模块1：服务商 + API Key + Base URL）
type SaveApiConfigRequest struct {
	ContentAnalysis struct {
		Provider string `json:"provider"`
		BaseURL  string `json:"base_url"`
	} `json:"content_analysis"`
	ImageAnalysis struct {
		Provider string `json:"provider"`
		BaseURL  string `json:"base_url"`
	} `json:"image_analysis"`
	VideoAnalysis struct {
		Provider string `json:"provider"`
		BaseURL  string `json:"base_url"`
	} `json:"video_analysis"`
	ProviderKeys ProviderApiKeys `json:"provider_keys"`
}

// SaveModelConfigRequest 保存模型选择请求（模块2：各分析类型的模型和服务商）
type SaveModelConfigRequest struct {
	ContentModel    string `json:"content_model"`
	ContentProvider string `json:"content_provider"`
	ImageModel      string `json:"image_model"`
	ImageProvider   string `json:"image_provider"`
	VideoModel      string `json:"video_model"`
	VideoProvider   string `json:"video_provider"`
}

// SaveGenerateConfigRequest 保存生成设置请求（模块3：仿写条数等）
type SaveGenerateConfigRequest struct {
	GenerateCount int `json:"generate_count"`
}

// SaveTaskTypeRequest 保存任务类型请求
type SaveTaskTypeRequest struct {
	TaskType string `json:"task_type"`
}

// GetLLMConfig 获取 LLM 配置
// @Summary 获取 LLM 配置
// @Tags Settings
// @Security BearerAuth
// @Success 200 {object} response.Response{data=MultiModalConfigResponse}
// @Router /settings/llm [get]
func (h *SettingsHandler) GetLLMConfig(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	settings, err := h.settingsRepo.GetByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		defaultConfig := LLMConfigItem{
			Provider: model.LLMProviderOpenAI,
			ApiKey:   "",
			Model:    "gpt-3.5-turbo",
			BaseURL:  "https://api.openai.com/v1",
		}
		response.Success(c, MultiModalConfigResponse{
			ContentAnalysis: defaultConfig,
			ImageAnalysis:   LLMConfigItem{Provider: model.LLMProviderOpenAI, Model: "gpt-4o", BaseURL: "https://api.openai.com/v1"},
			VideoAnalysis:   LLMConfigItem{Provider: model.LLMProviderOpenAI, Model: "gpt-4o", BaseURL: "https://api.openai.com/v1"},
			ProviderKeys:    ProviderApiKeys{},
			GenerateCount:   1,
		})
		return
	}
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}

	response.Success(c, MultiModalConfigResponse{
		ContentAnalysis: LLMConfigItem{
			Provider: settings.LLMProvider,
			ApiKey:   maskApiKey(getProviderApiKey(settings, settings.LLMProvider)),
			Model:    settings.LLMModel,
			BaseURL:  settings.LLMBaseURL,
		},
		ImageAnalysis: LLMConfigItem{
			Provider: settings.ImageLLMProvider,
			ApiKey:   maskApiKey(getProviderApiKey(settings, settings.ImageLLMProvider)),
			Model:    settings.ImageLLMModel,
			BaseURL:  settings.ImageLLMBaseURL,
		},
		VideoAnalysis: LLMConfigItem{
			Provider: settings.VideoLLMProvider,
			ApiKey:   maskApiKey(getProviderApiKey(settings, settings.VideoLLMProvider)),
			Model:    settings.VideoLLMModel,
			BaseURL:  settings.VideoLLMBaseURL,
		},
		ProviderKeys: ProviderApiKeys{
			OpenAI:    maskApiKey(settings.OpenAIApiKey),
			DeepSeek:  maskApiKey(settings.DeepSeekApiKey),
			Moonshot:  maskApiKey(settings.MoonshotApiKey),
			Qwen:      maskApiKey(settings.QwenApiKey),
			Hunyuan:   maskApiKey(settings.HunyuanApiKey),
			Doubao:    maskApiKey(settings.DoubaoApiKey),
			Zhipu:     maskApiKey(settings.ZhipuApiKey),
			Anthropic: maskApiKey(settings.AnthropicApiKey),
		},
		GenerateCount:   settings.GenerateCount,
		DefaultTaskType: settings.DefaultTaskType,
	})
}

// SaveApiConfig 保存 API 配置（模块1：服务商 + API Key + Base URL）
// @Summary 保存 API 配置
// @Tags Settings
// @Security BearerAuth
// @Param request body SaveApiConfigRequest true "API 配置"
// @Success 200 {object} response.Response
// @Router /settings/api-config [post]
func (h *SettingsHandler) SaveApiConfig(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req SaveApiConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取现有配置
	existing, _ := h.settingsRepo.GetByUserID(userID)
	if existing == nil {
		existing = &model.UserSettings{UserID: userID}
	}

	// 更新服务商和 Base URL
	if req.ContentAnalysis.Provider != "" {
		existing.LLMProvider = req.ContentAnalysis.Provider
	}
	if req.ContentAnalysis.BaseURL != "" {
		existing.LLMBaseURL = req.ContentAnalysis.BaseURL
	}
	if req.ImageAnalysis.Provider != "" {
		existing.ImageLLMProvider = req.ImageAnalysis.Provider
	}
	if req.ImageAnalysis.BaseURL != "" {
		existing.ImageLLMBaseURL = req.ImageAnalysis.BaseURL
	}
	if req.VideoAnalysis.Provider != "" {
		existing.VideoLLMProvider = req.VideoAnalysis.Provider
	}
	if req.VideoAnalysis.BaseURL != "" {
		existing.VideoLLMBaseURL = req.VideoAnalysis.BaseURL
	}

	// 更新各提供商 API Key
	existing.OpenAIApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.OpenAI, existing, "openai")
	existing.DeepSeekApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.DeepSeek, existing, "deepseek")
	existing.MoonshotApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.Moonshot, existing, "moonshot")
	existing.QwenApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.Qwen, existing, "qwen")
	existing.HunyuanApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.Hunyuan, existing, "hunyuan")
	existing.DoubaoApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.Doubao, existing, "doubao")
	existing.ZhipuApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.Zhipu, existing, "zhipu")
	existing.AnthropicApiKey = getApiKeyOrKeepExisting(req.ProviderKeys.Anthropic, existing, "anthropic")

	// 同步到旧字段
	existing.LLMApiKey = getProviderApiKey(existing, existing.LLMProvider)
	existing.ImageLLMApiKey = getProviderApiKey(existing, existing.ImageLLMProvider)
	existing.VideoLLMApiKey = getProviderApiKey(existing, existing.VideoLLMProvider)

	if err := h.settingsRepo.Upsert(existing); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}

	response.Success(c, gin.H{"message": "API 配置保存成功"})
}

// SaveModelConfig 保存模型选择（模块2：各分析类型的模型）
// @Summary 保存模型选择
// @Tags Settings
// @Security BearerAuth
// @Param request body SaveModelConfigRequest true "模型配置"
// @Success 200 {object} response.Response
// @Router /settings/model-config [post]
func (h *SettingsHandler) SaveModelConfig(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req SaveModelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取现有配置
	existing, _ := h.settingsRepo.GetByUserID(userID)
	if existing == nil {
		existing = &model.UserSettings{UserID: userID}
	}

	// 更新模型和服务商
	if req.ContentModel != "" {
		existing.LLMModel = req.ContentModel
	}
	if req.ContentProvider != "" {
		existing.LLMProvider = req.ContentProvider
		existing.LLMBaseURL = getProviderBaseURL(req.ContentProvider)
		// 同步 API Key
		existing.LLMApiKey = getProviderApiKey(existing, req.ContentProvider)
	}
	if req.ImageModel != "" {
		existing.ImageLLMModel = req.ImageModel
	}
	if req.ImageProvider != "" {
		existing.ImageLLMProvider = req.ImageProvider
		existing.ImageLLMBaseURL = getProviderBaseURL(req.ImageProvider)
		existing.ImageLLMApiKey = getProviderApiKey(existing, req.ImageProvider)
	}
	if req.VideoModel != "" {
		existing.VideoLLMModel = req.VideoModel
	}
	if req.VideoProvider != "" {
		existing.VideoLLMProvider = req.VideoProvider
		existing.VideoLLMBaseURL = getProviderBaseURL(req.VideoProvider)
		existing.VideoLLMApiKey = getProviderApiKey(existing, req.VideoProvider)
	}

	if err := h.settingsRepo.Upsert(existing); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}

	response.Success(c, gin.H{"message": "模型配置保存成功"})
}

// SaveGenerateConfig 保存生成设置（模块3：仿写条数等）
// @Summary 保存生成设置
// @Tags Settings
// @Security BearerAuth
// @Param request body SaveGenerateConfigRequest true "生成设置"
// @Success 200 {object} response.Response
// @Router /settings/generate-config [post]
func (h *SettingsHandler) SaveGenerateConfig(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req SaveGenerateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取现有配置
	existing, _ := h.settingsRepo.GetByUserID(userID)
	if existing == nil {
		existing = &model.UserSettings{UserID: userID}
	}

	// 更新生成配置
	if req.GenerateCount > 0 {
		if req.GenerateCount > 10 {
			req.GenerateCount = 10
		}
		existing.GenerateCount = req.GenerateCount
	}

	if err := h.settingsRepo.Upsert(existing); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}

	response.Success(c, gin.H{"message": "生成设置保存成功"})
}

// SaveTaskType 保存任务类型偏好
func (h *SettingsHandler) SaveTaskType(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req SaveTaskTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取现有配置
	existing, _ := h.settingsRepo.GetByUserID(userID)
	if existing == nil {
		existing = &model.UserSettings{UserID: userID}
	}

	// 更新
	if req.TaskType != "" {
		existing.DefaultTaskType = req.TaskType
	}

	if err := h.settingsRepo.Upsert(existing); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}

	response.Success(c, gin.H{"message": "任务偏好保存成功"})
}

// maskApiKey 脱敏 API Key
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		if len(apiKey) == 0 {
			return ""
		}
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// getApiKeyOrKeepExisting 保留现有 key
func getApiKeyOrKeepExisting(newKey string, existing *model.UserSettings, provider string) string {
	if len(newKey) > 0 && (newKey == "****" || (len(newKey) > 8 && newKey[4:8] == "****")) {
		if existing != nil {
			return getProviderApiKeyByName(existing, provider)
		}
		return ""
	}
	return newKey
}

// getProviderApiKey 根据提供商获取 API Key
func getProviderApiKey(settings *model.UserSettings, provider string) string {
	return getProviderApiKeyByName(settings, provider)
}

func getProviderApiKeyByName(settings *model.UserSettings, provider string) string {
	if settings == nil {
		return ""
	}
	switch provider {
	case model.LLMProviderOpenAI:
		return settings.OpenAIApiKey
	case model.LLMProviderDeepSeek:
		return settings.DeepSeekApiKey
	case model.LLMProviderMoonshot:
		return settings.MoonshotApiKey
	case model.LLMProviderQwen:
		return settings.QwenApiKey
	case model.LLMProviderHunyuan:
		return settings.HunyuanApiKey
	case model.LLMProviderDoubao:
		return settings.DoubaoApiKey
	case model.LLMProviderZhipu:
		return settings.ZhipuApiKey
	case model.LLMProviderAnthropic:
		return settings.AnthropicApiKey
	default:
		return ""
	}
}

// getProviderBaseURL 根据提供商获取 API Base URL
func getProviderBaseURL(provider string) string {
	switch provider {
	case model.LLMProviderOpenAI:
		return "https://api.openai.com/v1"
	case model.LLMProviderDeepSeek:
		return "https://api.deepseek.com"
	case model.LLMProviderMoonshot:
		return "https://api.moonshot.cn/v1"
	case model.LLMProviderQwen:
		return "https://dashscope.aliyuncs.com/compatible-mode/v1"
	case model.LLMProviderHunyuan:
		return "https://api.hunyuan.cloud.tencent.com/v1"
	case model.LLMProviderDoubao:
		return "https://ark.cn-beijing.volces.com/api/v3"
	case model.LLMProviderZhipu:
		return "https://open.bigmodel.cn/api/paas/v4"
	case model.LLMProviderAnthropic:
		return "https://api.anthropic.com/v1"
	default:
		return "https://api.openai.com/v1"
	}
}
