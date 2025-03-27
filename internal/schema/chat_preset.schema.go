package schema

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
