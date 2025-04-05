package services

import (
	"context"
	"encoding/json"
	"github.com/fcraft/open-chat/internal/schema"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	modelCollectionServiceInstance *ModelCollectionService
	modelCollectionServiceOnce     sync.Once
)

type ModelCollectionService struct {
	BaseService *BaseService
}

func InitModelCollectionService(base *BaseService) *ModelCollectionService {
	modelCollectionServiceOnce.Do(
		func() {
			modelCollectionServiceInstance = &ModelCollectionService{
				BaseService: base,
			}
		},
	)
	return modelCollectionServiceInstance
}

func GetModelCollectionService() *ModelCollectionService {
	return modelCollectionServiceInstance
}

const (
	modelCollectionCacheKeyPrefix = "model_collection:"
	cacheExpiration               = 1 * time.Hour
)

// GetCollectionByName 从缓存或数据库获取模型集合
func (s *ModelCollectionService) GetCollectionByName(name string) (*schema.ModelCollection, error) {
	cacheKey := modelCollectionCacheKeyPrefix + name

	// 尝试从Redis获取
	cachedData, err := s.BaseService.Redis.Get(context.Background(), cacheKey).Bytes()
	if err == nil {
		var collection schema.ModelCollection
		if err := json.Unmarshal(cachedData, &collection); err == nil {
			return &collection, nil
		}
	}

	// 从数据库获取
	var collection schema.ModelCollection
	if err := s.BaseService.Gorm.
		Preload("Models").
		Preload("Models.Provider").
		Where("name = ?", name).
		First(&collection).Error; err != nil {
		return nil, err
	}

	// 存入Redis
	if data, err := json.Marshal(collection); err == nil {
		s.BaseService.Redis.Set(context.Background(), cacheKey, data, cacheExpiration)
	}

	return &collection, nil
}

// GetRandomModelFromCollection 从集合中随机获取一个模型
func (s *ModelCollectionService) GetRandomModelFromCollection(collectionName string) (*schema.Model, error) {
	collection, err := s.GetCollectionByName(collectionName)
	if err != nil {
		return nil, err
	}

	if len(collection.Models) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// 简单实现：返回第一个模型
	// 实际应用中可以实现更复杂的负载均衡逻辑
	return &collection.Models[0], nil
}
