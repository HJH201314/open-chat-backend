package handlers

import "github.com/fcraft/open-chat/internal/storage/gorm"

type BaseHandler struct {
	Store *gorm.GormStore
}
