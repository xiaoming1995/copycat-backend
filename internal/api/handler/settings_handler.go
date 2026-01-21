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
	BatchSize int    `json:"batch_size,omitempty"` // 仿写条数（仅用于文案分析）
}

// MultiModalConfigRequest 多模态配置请求
type MultiModalConfigRequest struct {
	ContentAnalysis LLMConfigItem `json:"content_analysis"` // 文案分析
	ImageAnalysis   LLMConfigItem `json:"image_analysis"`   // 图片分析
	VideoAnalysis   LLMConfigItem `json:"video_analysis"`   // 视频分析
}

// MultiModalConfigResponse 多模态配置响应
type MultiModalConfigResponse struct {
	ContentAnalysis LLMConfigItem `json:"content_analysis"`
	ImageAnalysis   LLMConfigItem `json:"image_analysis"`
	VideoAnalysis   LLMConfigItem `json:"video_analysis"`
}

// GetLLMConfig 获取 LLM 配置
// @Summary 获取 LLM 配置（多模态）
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
		// 返回默认配置
		defaultConfig := LLMConfigItem{
			Provider: model.LLMProviderOpenAI,
			ApiKey:   "",
			Model:    "gpt-3.5-turbo",
			BaseURL:  "https://api.openai.com/v1",
		}
		response.Success(c, MultiModalConfigResponse{
			ContentAnalysis: defaultConfig,
			ImageAnalysis:   defaultConfig,
			VideoAnalysis:   defaultConfig,
		})
		return
	}
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}

	// 返回配置（API Key 脱敏显示）
	response.Success(c, MultiModalConfigResponse{
		ContentAnalysis: LLMConfigItem{
			Provider:  settings.LLMProvider,
			ApiKey:    maskApiKey(settings.LLMApiKey),
			Model:     settings.LLMModel,
			BaseURL:   settings.LLMBaseURL,
			BatchSize: settings.GenerateCount,
		},
		ImageAnalysis: LLMConfigItem{
			Provider: settings.ImageLLMProvider,
			ApiKey:   maskApiKey(settings.ImageLLMApiKey),
			Model:    settings.ImageLLMModel,
			BaseURL:  settings.ImageLLMBaseURL,
		},
		VideoAnalysis: LLMConfigItem{
			Provider: settings.VideoLLMProvider,
			ApiKey:   maskApiKey(settings.VideoLLMApiKey),
			Model:    settings.VideoLLMModel,
			BaseURL:  settings.VideoLLMBaseURL,
		},
	})
}

// SaveLLMConfig 保存 LLM 配置
// @Summary 保存 LLM 配置（多模态）
// @Tags Settings
// @Security BearerAuth
// @Param request body MultiModalConfigRequest true "LLM 配置"
// @Success 200 {object} response.Response
// @Router /settings/llm [post]
func (h *SettingsHandler) SaveLLMConfig(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req MultiModalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.ContentAnalysis.Provider == "" {
		req.ContentAnalysis.Provider = model.LLMProviderOpenAI
	}
	if req.ContentAnalysis.Model == "" {
		req.ContentAnalysis.Model = "gpt-3.5-turbo"
	}
	if req.ImageAnalysis.Provider == "" {
		req.ImageAnalysis.Provider = model.LLMProviderOpenAI
	}
	if req.ImageAnalysis.Model == "" {
		req.ImageAnalysis.Model = "gpt-4o"
	}
	if req.VideoAnalysis.Provider == "" {
		req.VideoAnalysis.Provider = model.LLMProviderOpenAI
	}
	if req.VideoAnalysis.Model == "" {
		req.VideoAnalysis.Model = "gpt-4o"
	}

	// 获取现有配置（用于保留未修改的 API Key）
	existing, _ := h.settingsRepo.GetByUserID(userID)

	settings := &model.UserSettings{
		UserID: userID,
		// 文案分析配置
		LLMProvider: req.ContentAnalysis.Provider,
		LLMApiKey:   getApiKeyOrKeepExisting(req.ContentAnalysis.ApiKey, existing, "content"),
		LLMModel:    req.ContentAnalysis.Model,
		LLMBaseURL:  req.ContentAnalysis.BaseURL,
		// 图片分析配置
		ImageLLMProvider: req.ImageAnalysis.Provider,
		ImageLLMApiKey:   getApiKeyOrKeepExisting(req.ImageAnalysis.ApiKey, existing, "image"),
		ImageLLMModel:    req.ImageAnalysis.Model,
		ImageLLMBaseURL:  req.ImageAnalysis.BaseURL,
		// 视频分析配置
		VideoLLMProvider: req.VideoAnalysis.Provider,
		VideoLLMApiKey:   getApiKeyOrKeepExisting(req.VideoAnalysis.ApiKey, existing, "video"),
		VideoLLMModel:    req.VideoAnalysis.Model,
		VideoLLMBaseURL:  req.VideoAnalysis.BaseURL,
		// 生成配置
		GenerateCount: req.ContentAnalysis.BatchSize,
	}

	if err := h.settingsRepo.Upsert(settings); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}

	response.Success(c, gin.H{"message": "配置保存成功"})
}

// maskApiKey 脱敏 API Key，只显示前后几位
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		if len(apiKey) == 0 {
			return ""
		}
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// getApiKeyOrKeepExisting 如果请求中包含脱敏的 key，则保留现有的
func getApiKeyOrKeepExisting(newKey string, existing *model.UserSettings, keyType string) string {
	// 如果请求的 key 包含 **** 说明是脱敏显示的，需要保留原值
	if len(newKey) > 0 && (newKey == "****" || (len(newKey) > 8 && newKey[4:8] == "****")) {
		if existing != nil {
			switch keyType {
			case "content":
				return existing.LLMApiKey
			case "image":
				return existing.ImageLLMApiKey
			case "video":
				return existing.VideoLLMApiKey
			}
		}
		return ""
	}
	return newKey
}
