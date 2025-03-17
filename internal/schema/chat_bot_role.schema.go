package schema

type BotRole struct {
	// 原始数据
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `json:"name"`              // 角色名称
	Description     uint64 `json:"description"`       // 角色描述
	PromptSessionId string `json:"prompt_session_id"` // 引用一个 session 中的对话作为 prompt
	AutoCreateAt

	// 组装数据
	PromptSession *Session `gorm:"foreignKey:PromptSessionId;references:ID" json:"prompt_session"`
}
