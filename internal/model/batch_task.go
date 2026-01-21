package model

import (
	"time"

	"github.com/google/uuid"
)

// BatchTask 批量分析任务模型
type BatchTask struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid();comment:批次任务ID" json:"id"`
	UserID       int64     `gorm:"column:user_id;not null;index;comment:关联用户ID" json:"user_id"`
	TotalCount   int       `gorm:"column:total_count;not null;comment:总链接数" json:"total_count"`
	SuccessCount int       `gorm:"column:success_count;default:0;comment:成功数" json:"success_count"`
	FailedCount  int       `gorm:"column:failed_count;default:0;comment:失败数" json:"failed_count"`
	Status       string    `gorm:"column:status;type:varchar(50);default:pending;index;comment:任务状态(pending/processing/completed/failed)" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index:idx_batch_tasks_created_at,sort:desc;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`

	// 关联关系（仅用于代码层面加载，不创建数据库外键）
	User     *User     `gorm:"-" json:"user,omitempty"`
	Projects []Project `gorm:"-" json:"projects,omitempty"`
}

// TableName 指定表名
func (BatchTask) TableName() string {
	return "batch_tasks"
}

// 批量任务状态常量
const (
	BatchTaskStatusPending    = "pending"    // 等待处理
	BatchTaskStatusProcessing = "processing" // 处理中
	BatchTaskStatusCompleted  = "completed"  // 已完成
	BatchTaskStatusFailed     = "failed"     // 失败
)

// BatchTaskItem 批量任务中的单个项目（用于请求/响应）
type BatchTaskItem struct {
	URL    string `json:"url"`
	Status string `json:"status"` // pending/success/failed
	Error  string `json:"error,omitempty"`
}
