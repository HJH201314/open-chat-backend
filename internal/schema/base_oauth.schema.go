package schema

import "gorm.io/datatypes"

type OAuthProvider struct {
	ID           uint64                       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string                       `gorm:"not null" json:"name"`
	AuthUrl      string                       `gorm:"not null" json:"auth_url"`
	TokenUrl     string                       `gorm:"not null" json:"token_url"`
	Scopes       datatypes.JSONType[[]string] `gorm:"type:json;" json:"scopes"`
	ClientId     string                       `gorm:"not null" json:"client_id"`
	ClientSecret string                       `gorm:"not null" json:"client_secret"`

	AutoCreateUpdateDeleteAt
}

func (o *OAuthProvider) TableName() string {
	return "oauth_providers"
}

type OAuthUser struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	OAuthProviderID uint64 `gorm:"index;column:oauth_provider_id" json:"oauth_provider_id"`
	OAuthUserName   string `gorm:"index;unique;column:oauth_user_name" json:"oauth_user_name"` // 用户在第三方平台中的用户名
	UserID          uint64 `gorm:"index" json:"user_id"`

	AutoCreateUpdateDeleteAt
}

func (o *OAuthUser) TableName() string {
	return "oauth_users"
}
