package schema

import (
	"github.com/fcraft/open-chat/internal/constants"
	"gorm.io/datatypes"
)

type Preset struct {
	// 原始数据
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"unique;index" json:"name"` // 角色名称
	Description     string `json:"description"`              // 角色描述
	PromptSessionId string `json:"prompt_session_id"`        // 引用一个 session 中的对话作为 prompt
	Module          string `gorm:"index" json:"type"`        // 角色所属模块（chat、tue 等）
	Version         int64  `gorm:"default:0" json:"version"` // 预设版本号，可能被用于标记是否需要强制更新
	AutoCreateUpdateDeleteAt

	// 组装数据
	PromptSession *Session `gorm:"foreignKey:PromptSessionId;references:ID" json:"prompt_session"`
}

// PresetCompletionRecord 记录预设的补全记录
type PresetCompletionRecord struct {
	// 原始数据
	ID       uint64                                `gorm:"primaryKey;autoIncrement" json:"id"`
	PresetID uint64                                `gorm:"index;not null" json:"preset_id"`
	Params   datatypes.JSONType[map[string]string] `gorm:"type:json" json:"params"`
	Content  string                                `json:"content"`
	Status   constants.CommonStatus                `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	AutoCreateUpdateAt

	Preset *Preset `gorm:"foreignKey:ID;references:PresetID" json:"preset"`
}
