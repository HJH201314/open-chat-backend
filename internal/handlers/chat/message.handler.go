package chat

import (
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// GetMessages
//
//	@Summary		获取消息
//	@Description	获取消息
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string																		true	"会话 ID"
//	@Param			req			query		entity.ParamPagingSort														true	"分页参数"
//	@Success		200			{object}	entity.CommonResponse[entity.PaginatedContinuationResponse[schema.Message]]	"返回数据"
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
	req.WithDefaultSize(20).WithMaxSize(50)
	messages, nextPage, err := h.Store.GetMessagesByPage(uri.SessionId, req.PagingParam, req.SortParam)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedContinuationResponse[schema.Message]{
			List:     messages,
			NextPage: nextPage,
		},
	)
}

// UpdateMessage
//
//	@Summary		更新消息
//	@Description	更新消息
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string									true	"消息 ID"
//	@Param			req	body		schema.Message							true	"更新的消息数据"
//	@Success		200	{object}	entity.CommonResponse[schema.Message]	"返回数据"
//	@Router			/chat/message/{id}/update [post]
func (h *Handler) UpdateMessage(c *gin.Context) {
	var uri PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req schema.Message
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 查消息
	message := schema.Message{
		ID: uri.ID,
	}
	if err := h.Db.Find(&message).Error; err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
	}

	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), message.SessionID) {
		ctx_utils.CustomError(c, 400, "no permission")
		return
	}

	updateMessage := schema.Message{
		Extra: datatypes.NewJSONType(maputil.Merge(message.Extra.Data(), req.Extra.Data())),
	}
	if err := h.Db.Model(&message).Updates(&updateMessage).Error; err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, true)
}
