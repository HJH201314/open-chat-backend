package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/fcraft/open-chat/internal/schema"
	"github.com/redis/go-redis/v9"
	"github.com/xeipuuv/gojsonschema"

	"gorm.io/gorm"
)

const (
	ConfigAvailableChatModelCollection = "available_chat_model_collections"
)

// RegisterConfigParams 注册配置的参数结构
type RegisterConfigParams struct {
	Name        string
	DisplayName string
	Schema      map[string]any
	Default     any
	Description string
	IsPublic    bool
}

type SystemConfigService struct {
	BaseService *BaseService
	db          *gorm.DB
	redis       *redis.Client
	cache       sync.Map // 内存缓存
}

var (
	systemConfigServiceInstance *SystemConfigService
	systemConfigOnce            sync.Once
)

// InitSystemConfigService 初始化系统配置服务
func InitSystemConfigService(base *BaseService) *SystemConfigService {
	systemConfigOnce.Do(
		func() {
			systemConfigServiceInstance = &SystemConfigService{
				BaseService: base,
				db:          base.Gorm,
				redis:       base.Redis,
			}
			// 自动迁移数据库表
			err := base.Gorm.AutoMigrate(&schema.SystemConfig{})
			if err != nil {
				panic(err)
			}
		},
	)

	if err := systemConfigServiceInstance.RegisterSystemConfig(
		RegisterConfigParams{
			Name:        ConfigAvailableChatModelCollection,
			DisplayName: "可选聊天模型集合",
			Schema: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]string{
							"type":        "string",
							"description": "the name of the collection",
						},
						"display_name": map[string]string{
							"type":        "string",
							"description": "the display name of the collection",
						},
						"order": map[string]string{
							"type":        "number",
							"description": "the sort order of the collection",
						},
						"is_default": map[string]string{
							"type":        "boolean",
							"description": "is default",
						},
					},
					"required": []string{"name", "display_name"},
				},
			},
			IsPublic: true,
		},
	); err != nil {
		return nil
	}
	if err := systemConfigServiceInstance.RegisterSystemConfig(
		RegisterConfigParams{
			Name:        "temp_gift_card",
			DisplayName: "礼品卡",
			Schema: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":        "string",
					"description": "the gift card code",
				},
			},
			Default:  datatypes.NewJSONType[any]([]string{""}),
			IsPublic: true,
		},
	); err != nil {
		return nil
	}
	return systemConfigServiceInstance
}

// GetSystemConfigService 获取系统配置服务
func GetSystemConfigService() *SystemConfigService {
	return systemConfigServiceInstance
}

// RegisterSystemConfig 注册系统配置
func (s *SystemConfigService) RegisterSystemConfig(params RegisterConfigParams) error {
	schemaByte, err := json.Marshal(params.Schema)
	if err != nil {
		return err
	}
	defaultValueStr, err := convertor.ToJson(params.Default)
	if err != nil {
		return err
	}
	config := schema.SystemConfig{
		Name:        params.Name,
		DisplayName: params.DisplayName,
		Schema:      string(schemaByte),
		Default:     defaultValueStr,
		Description: params.Description,
		IsPublic:    params.IsPublic,
	}

	// 尝试获取数据库中已有配置
	var savedConfig schema.SystemConfig
	if err := s.db.Where(
		"name = ?",
		params.Name,
	).First(&savedConfig).Error; err == nil && savedConfig.Value != "" {
		// 检验现存配置对于当前注册的 schema 的合法性
		if err := s.validateAgainstSchema(savedConfig.Value, config.Schema); err != nil {
			slog.Default().Error("Failed to validate existing config", "name", params.Name, "error", err.Error())
			// 备份现存配置
			savedConfig.Name = params.Name + "_backup_" + time.Now().Format("20060102150405")
			savedConfig.AutoCreateUpdateDeleteAt = schema.AutoCreateUpdateDeleteAt{} // 清除查询所获取的时间数据
			s.db.Create(&savedConfig)
		}
	}

	// 使用upsert方式保存配置
	result := s.db.Where(schema.SystemConfig{Name: params.Name}).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"display_name", "default", "schema", "description"}),
		},
	).Create(&config)
	if result.Error != nil {
		return result.Error
	}
	// 保存默认值
	if config.Value == "" && config.Default != "" {
		if err := s.SetConfig(params.Name, params.Default); err != nil {
			return err
		}
	}

	// 更新缓存
	if err := s.db.Where("name = ?", params.Name).First(&config).Error; err != nil {
		return err
	}
	s.cache.Store(params.Name, &config)
	s.updateRedisCache(params.Name, config)

	return nil
}

// GetConfig 获取系统配置
func (s *SystemConfigService) GetConfig(name string) (*schema.SystemConfig, error) {
	// 1. 检查内存缓存
	if val, ok := s.cache.Load(name); ok {
		return val.(*schema.SystemConfig), nil
	}

	// 2. 检查Redis缓存
	ctx := context.Background()
	redisKey := "system_config:" + name
	val, err := s.redis.Get(ctx, redisKey).Result()
	if err == nil {
		var config schema.SystemConfig
		if err := json.Unmarshal([]byte(val), &config); err == nil {
			s.cache.Store(name, &config)
			return &config, nil
		}
	}

	// 3. 从数据库查询
	var config schema.SystemConfig
	if err := s.db.Where("name = ?", name).First(&config).Error; err != nil {
		return nil, err
	}

	// 更新缓存
	s.cache.Store(name, &config)
	s.updateRedisCache(name, config)

	return &config, nil
}

// updateRedisCache 更新Redis缓存
func (s *SystemConfigService) updateRedisCache(name string, config schema.SystemConfig) {
	ctx := context.Background()
	redisKey := "system_config:" + name
	configBytes, _ := json.Marshal(config)
	s.redis.Set(ctx, redisKey, configBytes, 24*time.Hour)
}

// SetConfig 设置系统配置值
func (s *SystemConfigService) SetConfig(name string, value any) error {
	// 1. 获取当前配置
	config, err := s.GetConfig(name)
	if err != nil {
		return err
	}

	// 2. 校验value是否符合schema
	if err := s.validateAgainstSchema(value, config.Schema); err != nil {
		return err
	}

	// 3. 更新数据库
	valueStr, err := convertor.ToJson(value)
	if err != nil {
		return err
	}
	config.Value = valueStr
	if err := s.db.Where("name = ?", name).Save(config).Error; err != nil {
		return err
	}

	// 4. 更新缓存
	s.cache.Store(name, config)
	s.updateRedisCache(name, *config)

	return nil
}

// validateAgainstSchema 使用 gojsonschema 进行完整验证
func (s *SystemConfigService) validateAgainstSchema(value interface{}, schemaDef string) error {
	var valueLoader gojsonschema.JSONLoader
	switch value.(type) {
	case string:
		valueLoader = gojsonschema.NewStringLoader(value.(string))
	default:
		valueLoader = gojsonschema.NewGoLoader(value)
	}
	schemaLoader := gojsonschema.NewStringLoader(schemaDef)

	result, err := gojsonschema.Validate(schemaLoader, valueLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return fmt.Errorf("schema validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// ResetConfig 重置系统配置值
func (s *SystemConfigService) ResetConfig(name string) error {
	// 1. 从数据库读取默认值
	config, err := s.GetConfig(name)
	if err != nil {
		return err
	}

	// 2. 往数据库写入默认值
	if err := s.SetConfig(name, config.Default); err != nil {
		return err
	}

	return nil
}
