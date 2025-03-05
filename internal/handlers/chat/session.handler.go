package chat

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateSession
//
//	@Summary		创建会话
//	@Description	创建会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[string]
//	@Router			/chat/session/new [post]
func (h *Handler) CreateSession(c *gin.Context) {
	session := schema.Session{
		EnableContext: true, // 默认开启上下文
	}

	if err := h.Store.CreateSession(ctx_utils.GetUserId(c), &session); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create session")
		return
	}
	ctx_utils.Success(c, session.ID)
}

// DeleteSession
//
//	@Summary		删除会话
//	@Description	删除会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string	true	"会话 ID"
//	@Success		200			{object}	entity.CommonResponse[bool]
//	@Router			/chat/session/del/{session_id} [post]
func (h *Handler) DeleteSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.CustomError(c, 400, "no permission")
		return
	}
	// 执行删除操作
	if err := h.Helper.DeleteSession(uri.SessionId); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete session")
		return
	}
	ctx_utils.Success(c, true)
}

// GetSessions
//
//	@Summary		获取会话列表
//	@Description	获取会话列表
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort											true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PagingResponse[schema.Session]]	"返回数据"
//	@Router			/chat/session/list [get]
func (h *Handler) GetSessions(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 查询消息
	num, size := req.GetPageSize(20, 50)
	sessions, nextPage, err := h.Store.GetSessionsByPage(ctx_utils.GetUserId(c), num, size, req.SortParam)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(
		c, &entity.PagingResponse[schema.Session]{
			List:     sessions,
			NextPage: nextPage,
		},
	)
}
