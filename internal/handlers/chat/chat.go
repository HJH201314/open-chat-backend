package chat

import "github.com/fcraft/open-chat/internal/handlers"

type PathParamSessionId struct {
	SessionId string `uri:"session_id" binding:"required"`
}

type Handler struct {
	*handlers.BaseHandler
}

func NewChatHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}
