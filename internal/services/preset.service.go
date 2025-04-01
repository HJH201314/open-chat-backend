// Package services 预设缓存服务
package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/schema"
	gormStore "github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
)

type PresetService struct {
	BuiltinPresets map[string]*schema.Preset // 系统内置预设缓存
	BaseService
}

var (
	presetServiceInstance *PresetService
	presetServiceOnce     sync.Once
)

// InitPresetService 初始化预设缓存服务
func InitPresetService(base *BaseService) *PresetService {
	presetServiceOnce.Do(
		func() {
			presetServiceInstance = &PresetService{
				BuiltinPresets: make(map[string]*schema.Preset),
				BaseService:    *base,
			}
		},
	)
	return presetServiceInstance
}

// GetPresetService 获取预设缓存服务
func GetPresetService() *PresetService {
	if presetServiceInstance == nil {
		panic("PresetService not initialized")
	}
	return presetServiceInstance
}

// RegisterBuiltinPresets 注册内置预设
// 注册并保存至数据库，当且仅当版本号高于数据库时覆盖更新
//
// Parameters:
//   - name: 预设名称
//   - preset: 预设完整数据（也可使用关联 ID，但不推荐）
//   - version：版本号，用于更新时判断是否需要覆盖更新
func (s *PresetService) RegisterBuiltinPresets(name string, preset *schema.Preset, version int64) error {
	preset.Name = name
	preset.Version = version
	preset.Module = "builtin"
	// 查询数据库，若预设不存在则创建，否则从数据库中获取
	var presetData schema.Preset
	// 查询包括关联 session 和 messages 的完整数据
	err := s.Gorm.Scopes(gormStore.PresetFullScope).Where(
		"name = ? AND module = ?",
		name,
		"builtin",
	).Find(&presetData).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 数据库查询失败报错
		s.Logger.Error("failed to get builtin preset", "error", err)
		return err
	}

	// 无数据，创建并保存
	if errors.Is(err, gorm.ErrRecordNotFound) || presetData.Version < version {
		if s.Gorm.Session(&gorm.Session{FullSaveAssociations: true}).Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			},
		).Save(preset); err != nil {
			s.Logger.Error("failed to create builtin preset", "error", err)
			return err
		}
		// 因为可能插入的时候是使用的 ID 关联，所以插入后重新查询完整数据
		if err := s.Gorm.Scopes(gormStore.PresetFullScope).Where(
			"name = ? AND module = ?",
			name,
			"builtin",
		).Find(&presetData).Error; err != nil {
			s.Logger.Error("failed to load builtin preset", "error", err)
			return err
		}
		s.BuiltinPresets[name] = &presetData
	} else {
		s.BuiltinPresets[name] = &presetData
	}

	return nil
}

// GetBuiltinPreset 获取内置预设
// 从内存中获取预设，若不存在则从 redis 缓存中获取
func (s *PresetService) GetBuiltinPreset(name string) *schema.Preset {
	if maputil.HasKey(s.BuiltinPresets, name) {
		// 若在内存中有，直接返回
		return s.BuiltinPresets[name]
	} else {
		// 没有，查询 redis 缓存
		preset, err := s.RedisStore.GetCachedPresetByName(name)
		if err != nil {
			return nil
		}
		s.BuiltinPresets[name] = preset
		return preset
	}
}

// GetBuiltinPresets 获取所有内置预设
func (s *PresetService) GetBuiltinPresets() map[string]*schema.Preset {
	return s.BuiltinPresets
}

// RegisterBuiltinPresetsSimple 注册内置预设
func (s *PresetService) RegisterBuiltinPresetsSimple(name string, desc string, version int64, systemPrompt string, messages []chat_utils.Message) {
	// 将参数转换为 schema.Preset
	schemaMessages := slice.Map(
		messages, func(_ int, m chat_utils.Message) schema.Message {
			return schema.Message{
				Role:    m.Role,
				Content: m.Content,
			}
		},
	)
	preset := schema.Preset{
		Name:        name,
		Version:     version,
		Description: desc,
		PromptSession: &schema.Session{
			Name:         desc,
			SystemPrompt: systemPrompt,
			Messages:     schemaMessages,
		},
	}
	if err := s.RegisterBuiltinPresets(name, &preset, version); err != nil {
		s.Logger.Error("failed to register builtin preset", "error", err)
	} else {
		s.Logger.Info("register builtin preset successfully", "name", name)
	}
}

// BuiltinPresetCompletion 内置预设补全
func BuiltinPresetCompletion(presetName string, params map[string]string) (string, uint64, error) {
	presetService := GetPresetService()
	if presetService == nil {
		return "", 0, errors.New("preset service not found")
	}
	preset := presetService.GetBuiltinPreset(presetName)

	// 查询配置，获取默认的AI模型提供商 TODO：目前临时使用 deepseek-v3，后续更新可配置聚合 API
	var model schema.Model
	if err := presetService.Gorm.Model(&model).Preload("Provider").Preload("Provider.APIKeys").Where(
		"id = ?",
		19,
	).First(&model).Error; err != nil {
		return "", 0, fmt.Errorf("failed to get default AI provider: %w", err)
	}

	// 记录调用
	presetRecord := &schema.PresetCompletionRecord{
		PresetID: preset.ID,
		Status:   constants.StatusHandling,
		Params:   datatypes.NewJSONType[map[string]string](params),
	}
	if err := presetService.Gorm.Create(presetRecord).Error; err != nil {
		return "", 0, fmt.Errorf("failed to create preset record: %w", err)
	}

	// 调用AI接口进行补全
	resp, err := chat_utils.Completion(
		context.Background(), chat_utils.GetCommonCompletionOptions(
			&model, &chat_utils.CompletionOptions{
				CompletionModelConfig: chat_utils.CompletionModelConfig{
					MaxTokens:   1000, // 输出长度限制 TODO：跟随更新可配置后可自定义
					Temperature: 1.6,  // 较高的温度，提高灵活性 TODO：跟随更新可配置后可自定义
				},
				SystemPrompt: preset.PromptSession.SystemPrompt,
				Messages:     chat_utils.ConvertSchemaToMessages(preset.PromptSession.Messages, params),
			},
		),
	)

	// 更新记录
	defer func() {
		var status constants.CommonStatus
		if err != nil || resp.Content == "" {
			// 失败
			status = constants.StatusFailed
		} else {
			// 成功
			status = constants.StatusCompleted
		}
		if err := presetService.Gorm.Model(&presetRecord).Updates(
			schema.PresetCompletionRecord{
				Content: resp.Content,
				Status:  status,
			},
		).Error; err != nil {
			// do nothing
		}
	}()

	if err != nil {
		return "", presetRecord.ID, fmt.Errorf("failed to complete: %w", err)
	}
	if resp.Content == "" {
		return "", presetRecord.ID, errors.New("no content")
	}
	return resp.Content, presetRecord.ID, nil
}
