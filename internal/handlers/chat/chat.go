package chat

import "github.com/fcraft/open-chat/internal/handlers"

// PathParamSessionId 路径参数SessionId
type PathParamSessionId struct {
	SessionId string `uri:"session_id" binding:"required"`
}

// PathParamId 路径参数ID
type PathParamId struct {
	ID uint64 `uri:"id" binding:"required"`
}

type Handler struct {
	*handlers.BaseHandler
}

func NewChatHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}
