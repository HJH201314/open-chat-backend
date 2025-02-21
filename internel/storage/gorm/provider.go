package gorm

import (
	"github.com/fcraft/open-chat/internel/models"
	"gorm.io/gorm"
)

// AddProvider 添加提供商
func (s *GormStore) AddProvider(provider *models.Provider) error {
	return s.Db.Create(provider).Error
}

// UpdateProvider 更新提供商
func (s *GormStore) UpdateProvider(provider *models.Provider) error {
	return s.Db.Model(provider).Updates(provider).Error
}

// GetProvider 获取提供商
func (s *GormStore) GetProvider(providerId uint64) (*models.Provider, error) {
	var provider models.Provider
	return &provider, s.Db.Where("id = ?", providerId).First(&provider).Error
}

// GetProviders 获取提供商
func (s *GormStore) GetProviders() ([]models.Provider, error) {
	var providers []models.Provider
	return providers, s.Db.Find(&providers).Error
}

// DeleteProvider 删除提供商
func (s *GormStore) DeleteProvider(providerId uint64) error {
	return s.Db.Transaction(func(tx *gorm.DB) error {
		// 删除模型
		if err := tx.Where("provider_id = ?", providerId).Delete(&models.Model{}).Error; err != nil {
			return err
		}
		// 删除提供商
		return tx.Where("id = ?", providerId).Delete(&models.Provider{}).Error
	})
}

// AddModel 添加模型
func (s *GormStore) AddModel(model *models.Model) error {
	return s.Db.Create(model).Error
}

// UpdateModel 更新模型
func (s *GormStore) UpdateModel(model *models.Model) error {
	return s.Db.Model(model).Updates(model).Error
}

// GetModel 获取模型
func (s *GormStore) GetModel(modelId uint64) (*models.Model, error) {
	var model models.Model
	return &model, s.Db.Where("id = ?", modelId).First(&model).Error
}

// GetModelsByProvider 获取模型
func (s *GormStore) GetModelsByProvider(providerId uint64) ([]models.Model, error) {
	var aiModels []models.Model
	return aiModels, s.Db.Where("provider_id = ?", providerId).Find(&aiModels).Error
}

// DeleteModel 删除模型
func (s *GormStore) DeleteModel(modelId uint64) error {
	return s.Db.Where("id = ?", modelId).Delete(&models.Model{}).Error
}

// DeleteModelsByProvider 删除模型
func (s *GormStore) DeleteModelsByProvider(providerId uint64) error {
	return s.Db.Where("provider_id = ?", providerId).Delete(&models.Model{}).Error
}
