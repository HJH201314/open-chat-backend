package entity

type ConfigChatModel struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	IsDefault   bool   `json:"is_default"`
	Order       int64  `json:"order"`
}
