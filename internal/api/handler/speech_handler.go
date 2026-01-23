package handler

import (
	"copycat/internal/core/tts"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SpeechHandler 语音合成处理器
type SpeechHandler struct {
	settingsRepo *repository.UserSettingsRepository
}

// NewSpeechHandler 创建语音合成处理器
func NewSpeechHandler(db *gorm.DB) *SpeechHandler {
	return &SpeechHandler{
		settingsRepo: repository.NewUserSettingsRepository(db),
	}
}

// GenerateSpeechRequest 语音合成请求
type GenerateSpeechRequest struct {
	Text  string `json:"text" binding:"required"`  // 要合成的文本
	Voice string `json:"voice" binding:"required"` // 音色
	Model string `json:"model"`                    // 模型 (可选，默认 qwen3-tts-flash)
}

// GenerateSpeechResponse 语音合成响应
type GenerateSpeechResponse struct {
	AudioBase64 string `json:"audio_base64"` // Base64 编码的音频
	Format      string `json:"format"`       // 音频格式 (mp3)
	Characters  int    `json:"characters"`   // 合成的字符数
}

// VoiceItem 音色信息
type VoiceItem struct {
	ID          string `json:"id"`          // 音色ID
	Name        string `json:"name"`        // 中文名称
	Description string `json:"description"` // 描述
	Gender      string `json:"gender"`      // 性别
}

// ModelItem 模型信息
type ModelItem struct {
	ID          string `json:"id"`          // 模型ID
	Name        string `json:"name"`        // 模型名称
	Description string `json:"description"` // 描述
	VoiceCount  int    `json:"voice_count"` // 音色数量
	Languages   string `json:"languages"`   // 支持语言
}

// GenerateSpeech 生成语音
// @Summary 语音合成
// @Tags Speech
// @Security BearerAuth
// @Param request body GenerateSpeechRequest true "语音合成请求"
// @Success 200 {object} response.Response{data=GenerateSpeechResponse}
// @Router /speech/generate [post]
func (h *SpeechHandler) GenerateSpeech(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	// 解析请求
	var req GenerateSpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证文本长度
	if len(req.Text) == 0 {
		response.BadRequest(c, "文本不能为空")
		return
	}

	if len(req.Text) > 10000 {
		response.BadRequest(c, "文本长度不能超过 10000 字符")
		return
	}

	// 默认模型
	model := req.Model
	if model == "" {
		model = tts.DefaultModel
	}

	// 验证模型
	if !tts.IsValidModel(model) {
		response.BadRequest(c, "无效的模型: "+model)
		return
	}

	// 验证音色（根据模型）
	if !tts.IsValidVoiceForModel(req.Voice, model) {
		response.BadRequest(c, "无效的音色: "+req.Voice+" (模型: "+model+")")
		return
	}

	// 获取用户设置
	settings, err := h.settingsRepo.GetByUserID(userID.(int64))
	if err != nil {
		response.ServerError(c, "获取用户设置失败")
		return
	}

	// 获取 API Key（使用 Qwen API Key，因为阿里云百炼使用同一个）
	apiKey := settings.QwenApiKey
	if apiKey == "" {
		response.BadRequest(c, "请先在设置中配置通义千问 (Qwen) API Key")
		return
	}

	// 创建 TTS 客户端并合成语音
	client := tts.NewClient(apiKey)
	result, err := client.Synthesize(req.Text, req.Voice, model)
	if err != nil {
		response.ServerError(c, "语音合成失败: "+err.Error())
		return
	}

	// 返回结果
	response.Success(c, GenerateSpeechResponse{
		AudioBase64: result.AudioBase64,
		Format:      result.Format,
		Characters:  result.Characters,
	})
}

// GetVoices 获取可用音色列表
// @Summary 获取可用音色列表
// @Tags Speech
// @Security BearerAuth
// @Param model query string false "模型名称"
// @Success 200 {object} response.Response{data=[]VoiceItem}
// @Router /speech/voices [get]
func (h *SpeechHandler) GetVoices(c *gin.Context) {
	model := c.Query("model")
	if model == "" {
		model = tts.DefaultModel
	}

	voices := tts.GetVoicesForModel(model)

	// 转换为响应格式
	items := make([]VoiceItem, len(voices))
	for i, v := range voices {
		items[i] = VoiceItem{
			ID:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			Gender:      v.Gender,
		}
	}

	response.Success(c, items)
}

// GetModels 获取可用模型列表
// @Summary 获取可用TTS模型列表
// @Tags Speech
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]ModelItem}
// @Router /speech/models [get]
func (h *SpeechHandler) GetModels(c *gin.Context) {
	models := tts.GetModels()

	// 转换为响应格式
	items := make([]ModelItem, len(models))
	for i, m := range models {
		items[i] = ModelItem{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			VoiceCount:  m.VoiceCount,
			Languages:   m.Languages,
		}
	}

	response.Success(c, items)
}
