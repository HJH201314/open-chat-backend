package chat

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

// GetMessages
//
//	@Summary		获取消息
//	@Description	获取消息
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string															true	"会话 ID"
//	@Param			req			query		entity.ParamPagingSort											true	"分页参数"
//	@Success		200			{object}	entity.CommonResponse[entity.PagingResponse[schema.Message]]	"返回数据"
//	@Router			/chat/message/list/{session_id} [get]
func (h *Handler) GetMessages(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req entity.ParamPagingSort
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.CustomError(c, 400, "no permission")
		return
	}
	// 查询消息
	num, size := req.GetPageSize(20, 50)
	messages, nextPage, err := h.Store.GetMessagesByPage(uri.SessionId, num, size, req.SortParam)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(
		c, &entity.PagingResponse[schema.Message]{
			List:     messages,
			NextPage: nextPage,
		},
	)
}
