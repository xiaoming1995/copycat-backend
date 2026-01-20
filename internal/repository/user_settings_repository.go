package repository

import (
	"copycat/internal/model"

	"gorm.io/gorm"
)

// UserSettingsRepository 用户设置仓库
type UserSettingsRepository struct {
	db *gorm.DB
}

// NewUserSettingsRepository 创建用户设置仓库
func NewUserSettingsRepository(db *gorm.DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

// GetByUserID 根据用户ID获取设置
func (r *UserSettingsRepository) GetByUserID(userID int64) (*model.UserSettings, error) {
	var settings model.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Create 创建用户设置
func (r *UserSettingsRepository) Create(settings *model.UserSettings) error {
	return r.db.Create(settings).Error
}

// Update 更新用户设置
func (r *UserSettingsRepository) Update(settings *model.UserSettings) error {
	return r.db.Save(settings).Error
}

// Upsert 创建或更新用户设置
func (r *UserSettingsRepository) Upsert(settings *model.UserSettings) error {
	// 先尝试查询是否存在
	existing, err := r.GetByUserID(settings.UserID)
	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return r.Create(settings)
	}
	if err != nil {
		return err
	}
	// 存在，更新记录
	settings.ID = existing.ID
	settings.CreatedAt = existing.CreatedAt
	return r.Update(settings)
}

// Delete 删除用户设置
func (r *UserSettingsRepository) Delete(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserSettings{}).Error
}
