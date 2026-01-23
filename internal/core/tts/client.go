package tts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"copycat/pkg/logger"
)

// 阿里云百炼 TTS API 配置
const (
	DashScopeAPIURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	DefaultModel    = "qwen3-tts-flash"
)

// Model TTS 模型信息
type Model struct {
	ID          string `json:"id"`          // 模型ID
	Name        string `json:"name"`        // 模型名称
	Description string `json:"description"` // 描述
	VoiceCount  int    `json:"voice_count"` // 音色数量
	Languages   string `json:"languages"`   // 支持语言
}

// 可用的 TTS 模型
var AvailableModels = []Model{
	{ID: "qwen3-tts-flash", Name: "通义千问TTS-Flash", Description: "官方推荐，49种音色，支持多语言", VoiceCount: 49, Languages: "中/英/法/德/俄/意/西/葡/日/韩"},
	{ID: "qwen-tts", Name: "通义千问TTS", Description: "标准版，7种音色，仅中英", VoiceCount: 7, Languages: "中/英"},
}

// Voice 音色信息
type Voice struct {
	ID          string `json:"id"`          // 音色参数值
	Name        string `json:"name"`        // 中文名称
	Description string `json:"description"` // 描述
	Gender      string `json:"gender"`      // 性别
	Languages   string `json:"languages"`   // 支持语种
	Model       string `json:"model"`       // 所属模型
}

// qwen3-tts-flash 专用音色列表 (49种)
var Qwen3TTSFlashVoices = []Voice{
	{ID: "Cherry", Name: "芊悦", Description: "阳光积极、亲切自然小姐姐", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Serena", Name: "苏瑶", Description: "温柔小姐姐", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Ethan", Name: "晨煦", Description: "阳光、温暖、活力、朝气", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Chelsie", Name: "千雪", Description: "二次元虚拟女友", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Momo", Name: "茉兔", Description: "撒娇搞怪，逗你开心", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Vivian", Name: "十三", Description: "拽拽的、可爱的小暴躁", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Moon", Name: "月白", Description: "率性帅气", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Maia", Name: "四月", Description: "知性与温柔的碰撞", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Kai", Name: "凯", Description: "耳朵的一场SPA", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Nofish", Name: "不吃鱼", Description: "不会翘舌音的设计师", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Bella", Name: "萌宝", Description: "喝酒不打醉拳的小萝莉", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Jennifer", Name: "詹妮弗", Description: "品牌级、电影质感般美语女声", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Ryan", Name: "甜茶", Description: "节奏拉满，戏感炸裂", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Katerina", Name: "卡捷琳娜", Description: "御姐音色，韵律回味十足", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Aiden", Name: "艾登", Description: "精通厨艺的美语大男孩", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Eldric Sage", Name: "沧明子", Description: "沉稳睿智的老者", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Mia", Name: "乖小妹", Description: "温顺如春水，乖巧如初雪", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Mochi", Name: "沙小弥", Description: "聪明伶俐的小大人", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Bellona", Name: "燕铮莺", Description: "声音洪亮，金戈铁马入梦来", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Vincent", Name: "田叔", Description: "独特的沙哑烟嗓", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Bunny", Name: "萌小姬", Description: "萌属性爆棚的小萝莉", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Neil", Name: "阿闻", Description: "最专业的新闻主持人", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Elias", Name: "墨讲师", Description: "严谨且具叙事技巧的讲师", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Arthur", Name: "徐大爷", Description: "质朴嗓音", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Nini", Name: "邻家妹妹", Description: "糯米糍一样又软又黏的嗓音", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Ebona", Name: "诡婆婆", Description: "像生锈的钥匙转动幽暗角落", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Seren", Name: "小婉", Description: "温和舒缓，助你进入睡眠", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Pip", Name: "顽屁小孩", Description: "调皮捣蛋却充满童真", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Stella", Name: "少女阿月", Description: "迷糊少女音与正义感切换", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Bodega", Name: "博德加", Description: "热情的西班牙大叔", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Sonrisa", Name: "索尼莎", Description: "热情开朗的拉美大姐", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Alek", Name: "阿列克", Description: "战斗民族的冷与暖", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Dolce", Name: "多尔切", Description: "慵懒的意大利大叔", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Sohee", Name: "素熙", Description: "温柔开朗的韩国欧尼", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Ono Anna", Name: "小野杏", Description: "鬼灵精怪的青梅竹马", Gender: "女", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Lenn", Name: "莱恩", Description: "穿西装也听后朋克的德国青年", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Emilien", Name: "埃米尔安", Description: "浪漫的法国大哥哥", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Andre", Name: "安德雷", Description: "声音磁性，自然舒服、沉稳男生", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	{ID: "Radio Gol", Name: "拉迪奥·戈尔", Description: "足球解说诗人", Gender: "男", Languages: "中/英/法/德/俄/意/西/葡/日/韩", Model: "qwen3-tts-flash"},
	// 方言音色
	{ID: "Jada", Name: "上海-阿珍", Description: "风风火火的沪上阿姐", Gender: "女", Languages: "上海话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Dylan", Name: "北京-晓东", Description: "北京胡同里长大的少年", Gender: "男", Languages: "北京话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Li", Name: "南京-老李", Description: "耐心的瑜伽老师", Gender: "男", Languages: "南京话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Marcus", Name: "陕西-秦川", Description: "面宽话短、心实声沉的老陕味", Gender: "男", Languages: "陕西话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Roy", Name: "闽南-阿杰", Description: "诙谐直爽、市井活泼的台湾哥仔", Gender: "男", Languages: "闽南语/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Peter", Name: "天津-李彼得", Description: "天津相声，专业捧哏", Gender: "男", Languages: "天津话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Sunny", Name: "四川-晴儿", Description: "甜到你心里的川妹子", Gender: "女", Languages: "四川话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Eric", Name: "四川-程川", Description: "跳脱市井的四川成都男子", Gender: "男", Languages: "四川话/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Rocky", Name: "粤语-阿强", Description: "幽默风趣的阿强，在线陪聊", Gender: "男", Languages: "粤语/中/英等", Model: "qwen3-tts-flash"},
	{ID: "Kiki", Name: "粤语-阿清", Description: "甜美的港妹闺蜜", Gender: "女", Languages: "粤语/中/英等", Model: "qwen3-tts-flash"},
}

// qwen-tts 专用音色列表 (7种)
var QwenTTSVoices = []Voice{
	{ID: "longxiaochun", Name: "龙小淳", Description: "知性女声", Gender: "女", Languages: "中/英", Model: "qwen-tts"},
	{ID: "longxiaoxia", Name: "龙小夏", Description: "温柔女声", Gender: "女", Languages: "中/英", Model: "qwen-tts"},
	{ID: "longlaotie", Name: "龙老铁", Description: "东北男声", Gender: "男", Languages: "中/英", Model: "qwen-tts"},
	{ID: "longshu", Name: "龙舒", Description: "儒雅男声", Gender: "男", Languages: "中/英", Model: "qwen-tts"},
	{ID: "longshuo", Name: "龙硕", Description: "成熟男声", Gender: "男", Languages: "中/英", Model: "qwen-tts"},
	{ID: "longyue", Name: "龙悦", Description: "甜美女声", Gender: "女", Languages: "中/英", Model: "qwen-tts"},
	{ID: "longfei", Name: "龙飞", Description: "激情男声", Gender: "男", Languages: "中/英", Model: "qwen-tts"},
}

// Client TTS 客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建 TTS 客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// TTSRequest 阿里云百炼 TTS 请求结构
type TTSRequest struct {
	Model string   `json:"model"`
	Input TTSInput `json:"input"`
}

// TTSInput TTS 输入参数
type TTSInput struct {
	Text         string `json:"text"`
	Voice        string `json:"voice"`
	LanguageType string `json:"language_type,omitempty"`
}

// TTSResponse 阿里云百炼 TTS 响应结构
type TTSResponse struct {
	StatusCode int `json:"status_code"`
	Output     struct {
		Text         *string `json:"text"`
		FinishReason string  `json:"finish_reason"`
		Choices      *string `json:"choices"`
		Audio        struct {
			Data      string `json:"data"`       // Base64 编码的音频
			URL       string `json:"url"`        // 音频 URL
			ID        string `json:"id"`         // 音频 ID
			ExpiresAt int64  `json:"expires_at"` // 过期时间
		} `json:"audio"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		Characters   int `json:"characters"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
	// 错误信息
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SynthesizeResult 合成结果
type SynthesizeResult struct {
	AudioBase64 string `json:"audio_base64"` // Base64 编码的音频
	Format      string `json:"format"`       // 音频格式
	Characters  int    `json:"characters"`   // 字符数
}

// Synthesize 文本转语音
func (c *Client) Synthesize(text, voice, model string) (*SynthesizeResult, error) {
	// 如果没有指定模型，使用默认模型
	if model == "" {
		model = DefaultModel
	}

	// 验证模型是否有效
	if !IsValidModel(model) {
		return nil, fmt.Errorf("无效的模型: %s", model)
	}

	// 验证音色是否有效（根据模型）
	if !IsValidVoiceForModel(voice, model) {
		return nil, fmt.Errorf("无效的音色: %s (模型: %s)", voice, model)
	}

	// 构建请求
	req := TTSRequest{
		Model: model,
		Input: TTSInput{
			Text:         text,
			Voice:        voice,
			LanguageType: "Chinese",
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求
	httpReq, err := http.NewRequest("POST", DashScopeAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	logger.Info("TTS 请求: model=%s, voice=%s, text_length=%d", model, voice, len(text))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var ttsResp TTSResponse
	if err := json.Unmarshal(body, &ttsResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if ttsResp.Code != "" {
		logger.Error("TTS API 错误: code=%s, message=%s", ttsResp.Code, ttsResp.Message)
		return nil, fmt.Errorf("TTS API 错误: %s - %s", ttsResp.Code, ttsResp.Message)
	}

	// 获取音频数据
	audioData := ttsResp.Output.Audio.Data
	if audioData == "" && ttsResp.Output.Audio.URL != "" {
		// 如果 Data 为空但有 URL，从 URL 下载
		audioData, err = downloadAudioAsBase64(ttsResp.Output.Audio.URL)
		if err != nil {
			return nil, fmt.Errorf("下载音频失败: %w", err)
		}
	}

	if audioData == "" {
		return nil, fmt.Errorf("TTS 响应中没有音频数据")
	}

	logger.Info("TTS 成功: characters=%d, request_id=%s", ttsResp.Usage.Characters, ttsResp.RequestID)

	return &SynthesizeResult{
		AudioBase64: audioData,
		Format:      "mp3",
		Characters:  ttsResp.Usage.Characters,
	}, nil
}

// downloadAudioAsBase64 从 URL 下载音频并转换为 Base64
func downloadAudioAsBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("下载音频失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载音频失败: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取音频数据失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// IsValidModel 验证模型是否有效
func IsValidModel(model string) bool {
	for _, m := range AvailableModels {
		if m.ID == model {
			return true
		}
	}
	return false
}

// IsValidVoiceForModel 验证音色是否对指定模型有效
func IsValidVoiceForModel(voice, model string) bool {
	voices := GetVoicesForModel(model)
	for _, v := range voices {
		if v.ID == voice {
			return true
		}
	}
	return false
}

// GetVoicesForModel 获取指定模型的可用音色
func GetVoicesForModel(model string) []Voice {
	switch model {
	case "qwen3-tts-flash":
		return Qwen3TTSFlashVoices
	case "qwen-tts":
		return QwenTTSVoices
	default:
		return Qwen3TTSFlashVoices
	}
}

// GetModels 获取所有可用模型
func GetModels() []Model {
	return AvailableModels
}

// GetVoices 获取所有可用音色 (deprecated, use GetVoicesForModel)
func GetVoices() []Voice {
	return Qwen3TTSFlashVoices
}
