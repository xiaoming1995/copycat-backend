package model

import (
	"time"
)

// UserSettings 用户设置模型（存储 LLM 配置等）
type UserSettings struct {
	ID     int64 `gorm:"column:id;primaryKey;autoIncrement;comment:设置ID(自增)" json:"id"`
	UserID int64 `gorm:"column:user_id;uniqueIndex;not null;comment:关联用户ID" json:"user_id"`

	// 文案分析 LLM 配置（也用于仿写生成）
	LLMProvider string `gorm:"column:llm_provider;type:varchar(50);default:openai;comment:文案LLM服务商" json:"llm_provider"`
	LLMApiKey   string `gorm:"column:llm_api_key;type:varchar(500);comment:文案LLM API密钥" json:"llm_api_key"`
	LLMModel    string `gorm:"column:llm_model;type:varchar(100);default:gpt-3.5-turbo;comment:文案LLM模型名称" json:"llm_model"`
	LLMBaseURL  string `gorm:"column:llm_base_url;type:varchar(500);comment:文案LLM API基础URL" json:"llm_base_url"`

	// 图片分析 LLM 配置
	ImageLLMProvider string `gorm:"column:image_llm_provider;type:varchar(50);default:openai;comment:图片LLM服务商" json:"image_llm_provider"`
	ImageLLMApiKey   string `gorm:"column:image_llm_api_key;type:varchar(500);comment:图片LLM API密钥" json:"image_llm_api_key"`
	ImageLLMModel    string `gorm:"column:image_llm_model;type:varchar(100);default:gpt-4o;comment:图片LLM模型名称" json:"image_llm_model"`
	ImageLLMBaseURL  string `gorm:"column:image_llm_base_url;type:varchar(500);comment:图片LLM API基础URL" json:"image_llm_base_url"`

	// 视频分析 LLM 配置
	VideoLLMProvider string `gorm:"column:video_llm_provider;type:varchar(50);default:openai;comment:视频LLM服务商" json:"video_llm_provider"`
	VideoLLMApiKey   string `gorm:"column:video_llm_api_key;type:varchar(500);comment:视频LLM API密钥" json:"video_llm_api_key"`
	VideoLLMModel    string `gorm:"column:video_llm_model;type:varchar(100);default:gpt-4o;comment:视频LLM模型名称" json:"video_llm_model"`
	VideoLLMBaseURL  string `gorm:"column:video_llm_base_url;type:varchar(500);comment:视频LLM API基础URL" json:"video_llm_base_url"`

	// 各提供商 API Key（所有分析类型共享）
	OpenAIApiKey    string `gorm:"column:openai_api_key;type:varchar(500);comment:OpenAI API密钥" json:"openai_api_key"`
	DeepSeekApiKey  string `gorm:"column:deepseek_api_key;type:varchar(500);comment:DeepSeek API密钥" json:"deepseek_api_key"`
	MoonshotApiKey  string `gorm:"column:moonshot_api_key;type:varchar(500);comment:Moonshot API密钥" json:"moonshot_api_key"`
	QwenApiKey      string `gorm:"column:qwen_api_key;type:varchar(500);comment:通义千问 API密钥" json:"qwen_api_key"`
	HunyuanApiKey   string `gorm:"column:hunyuan_api_key;type:varchar(500);comment:腾讯混元 API密钥" json:"hunyuan_api_key"`
	DoubaoApiKey    string `gorm:"column:doubao_api_key;type:varchar(500);comment:豆包 API密钥" json:"doubao_api_key"`
	ZhipuApiKey     string `gorm:"column:zhipu_api_key;type:varchar(500);comment:智谱 API密钥" json:"zhipu_api_key"`
	AnthropicApiKey string `gorm:"column:anthropic_api_key;type:varchar(500);comment:Anthropic API密钥" json:"anthropic_api_key"`

	// 仿写生成配置
	GenerateCount int `gorm:"column:generate_count;default:1;comment:一次生成的仿写条数(1-10)" json:"generate_count"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (UserSettings) TableName() string {
	return "user_settings"
}

// LLMProvider 服务商常量
const (
	LLMProviderOpenAI    = "openai"
	LLMProviderDeepSeek  = "deepseek"
	LLMProviderAnthropic = "anthropic"
	LLMProviderMoonshot  = "moonshot"
	LLMProviderQwen      = "qwen"
	LLMProviderHunyuan   = "hunyuan"
	LLMProviderDoubao    = "doubao"
	LLMProviderZhipu     = "zhipu"
)
