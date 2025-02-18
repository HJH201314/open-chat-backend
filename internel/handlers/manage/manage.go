package manage

import (
	"github.com/fcraft/open-chat/internel/handlers"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewManageHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}
