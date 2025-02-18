package handlers

import "github.com/fcraft/open-chat/internel/storage/gorm"

type BaseHandler struct {
	Store *gorm.GormStore
}
