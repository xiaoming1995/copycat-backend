package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Project 创作项目模型
type Project struct {
	ID               uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid();comment:项目唯一ID(UUID)" json:"id"`
	UserID           int64          `gorm:"column:user_id;not null;index;comment:关联用户ID" json:"user_id"`
	BatchTaskID      *uuid.UUID     `gorm:"column:batch_task_id;type:uuid;index;comment:关联批量任务ID(可选)" json:"batch_task_id,omitempty"`
	SourceURL        string         `gorm:"column:source_url;type:text;comment:原始文案来源URL(小红书/公众号)" json:"source_url"`
	SourceContent    string         `gorm:"column:source_content;type:text;not null;comment:爬取/输入的原始文案内容" json:"source_content"`
	AnalysisResult   datatypes.JSON `gorm:"column:analysis_result;type:jsonb;comment:LLM分析结果(情绪/结构/关键词)" json:"analysis_result"`
	NewTopic         string         `gorm:"column:new_topic;type:varchar(500);comment:用户输入的新主题" json:"new_topic"`
	GeneratedContent string         `gorm:"column:generated_content;type:text;comment:LLM生成的仿写文案" json:"generated_content"`
	Status           string         `gorm:"column:status;type:varchar(50);default:draft;index;comment:项目状态(draft/analyzed/completed)" json:"status"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime;index:idx_projects_created_at,sort:desc;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`

	// 关联关系（仅用于代码层面加载，不创建数据库外键）
	User *User `gorm:"-" json:"user,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectStatus 项目状态常量
const (
	ProjectStatusDraft     = "draft"     // 草稿
	ProjectStatusAnalyzed  = "analyzed"  // 已分析
	ProjectStatusCompleted = "completed" // 已完成
)
