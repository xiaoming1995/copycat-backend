package repository

import (
	"copycat/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BatchTaskRepository 批量任务仓库
type BatchTaskRepository struct {
	db *gorm.DB
}

// NewBatchTaskRepository 创建批量任务仓库实例
func NewBatchTaskRepository(db *gorm.DB) *BatchTaskRepository {
	return &BatchTaskRepository{db: db}
}

// Create 创建批量任务
func (r *BatchTaskRepository) Create(task *model.BatchTask) error {
	return r.db.Create(task).Error
}

// FindByID 根据ID查询批量任务
func (r *BatchTaskRepository) FindByID(id uuid.UUID) (*model.BatchTask, error) {
	var task model.BatchTask
	if err := r.db.First(&task, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// FindByIDWithProjects 根据ID查询批量任务（包含关联的项目）
func (r *BatchTaskRepository) FindByIDWithProjects(id uuid.UUID) (*model.BatchTask, error) {
	var task model.BatchTask
	if err := r.db.First(&task, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// 手动加载关联的项目（不使用外键约束）
	var projects []model.Project
	if err := r.db.Where("batch_task_id = ?", id).Find(&projects).Error; err != nil {
		return nil, err
	}
	task.Projects = projects

	return &task, nil
}

// FindByUserID 根据用户ID查询批量任务列表
func (r *BatchTaskRepository) FindByUserID(userID int64, limit, offset int) ([]model.BatchTask, int64, error) {
	var tasks []model.BatchTask
	var total int64

	// 统计总数
	if err := r.db.Model(&model.BatchTask{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Update 更新批量任务
func (r *BatchTaskRepository) Update(task *model.BatchTask) error {
	return r.db.Save(task).Error
}

// UpdateStatus 更新任务状态
func (r *BatchTaskRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&model.BatchTask{}).Where("id = ?", id).Update("status", status).Error
}

// IncrementSuccessCount 增加成功计数
func (r *BatchTaskRepository) IncrementSuccessCount(id uuid.UUID) error {
	return r.db.Model(&model.BatchTask{}).Where("id = ?", id).
		UpdateColumn("success_count", gorm.Expr("success_count + 1")).Error
}

// IncrementFailedCount 增加失败计数
func (r *BatchTaskRepository) IncrementFailedCount(id uuid.UUID) error {
	return r.db.Model(&model.BatchTask{}).Where("id = ?", id).
		UpdateColumn("failed_count", gorm.Expr("failed_count + 1")).Error
}

// Delete 删除批量任务
func (r *BatchTaskRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.BatchTask{}, "id = ?", id).Error
}
