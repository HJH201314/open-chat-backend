package course

import "github.com/fcraft/open-chat/internal/handlers"

type PathParamId struct {
	ID uint64 `uri:"id" binding:"required"`
}

type Handler struct {
	*handlers.BaseHandler
}

func NewCourseHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}
