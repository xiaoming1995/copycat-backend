package repository

import (
	"context"
	"fmt"

	"copycat/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectRepository 项目数据仓库接口
type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Project, int64, error)
	GetBySourceURL(ctx context.Context, userID int64, sourceURL string) (*model.Project, error)
	Update(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// projectRepository 项目数据仓库实现
type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库实例
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create 创建项目
func (r *projectRepository) Create(ctx context.Context, project *model.Project) error {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetByID 根据 ID 获取项目
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var project model.Project
	if err := r.db.WithContext(ctx).First(&project, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get project by id: %w", err)
	}
	return &project, nil
}

// GetByUserID 根据用户 ID 获取项目列表 (分页)
func (r *projectRepository) GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64

	// 查询总数
	if err := r.db.WithContext(ctx).Model(&model.Project{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get projects by user id: %w", err)
	}

	return projects, total, nil
}

// Update 更新项目
func (r *projectRepository) Update(ctx context.Context, project *model.Project) error {
	if err := r.db.WithContext(ctx).Save(project).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// Delete 删除项目
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Project{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// UpdateStatus 更新项目状态
func (r *projectRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := r.db.WithContext(ctx).Model(&model.Project{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}
	return nil
}

// GetBySourceURL 根据用户ID和来源URL获取项目（查找已有的分析结果）
func (r *projectRepository) GetBySourceURL(ctx context.Context, userID int64, sourceURL string) (*model.Project, error) {
	var project model.Project
	// 查找用户已分析的同链接项目（有分析结果的优先）
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND source_url = ? AND analysis_result IS NOT NULL", userID, sourceURL).
		Order("created_at DESC").
		First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}
