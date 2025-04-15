package chat

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
		ctx_utils.BizError(c, constants.BizErrNoPermission)
		return
	}
	// 执行删除操作
	if err := h.Helper.DeleteSession(uri.SessionId); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete session")
		return
	}
	ctx_utils.Success(c, true)
}

// GetSession
//
//	@Summary		获取会话
//	@Description	获取会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string									true	"会话 ID"
//	@Success		200			{object}	entity.CommonResponse[schema.Session]	"返回数据"
//	@Router			/chat/session/{session_id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.BizError(c, constants.BizErrNoPermission)
		return
	}
	// 查询会话
	session, err := h.Store.GetSession(uri.SessionId)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(c, session)
}

// GetUserSession
//
//	@Summary		获取用户会话
//	@Description	获取用户会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string										true	"会话 ID"
//	@Success		200			{object}	entity.CommonResponse[schema.UserSession]	"返回数据"
//	@Router			/chat/session/user/{session_id} [get]
func (h *Handler) GetUserSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.BizError(c, constants.BizErrNoPermission)
		return
	}
	// 查询会话
	var userSession schema.UserSession
	err := h.Db.Model(&schema.UserSession{}).Where(
		"session_id = ? AND user_id = ?",
		uri.SessionId,
		ctx_utils.GetUserId(c),
	).Last(&userSession).Error
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(c, userSession)
}

// GetSharedSession
//
//	@Summary		获取已分享的用户会话信息
//	@Description	获取已分享的用户会话信息（仅返回 Name）
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string										true	"会话 ID"
//	@Param			req			query		chat.GetSharedSession.Req					true	"请求参数"
//	@Success		200			{object}	entity.CommonResponse[schema.UserSession]	"返回数据"
//	@Router			/chat/session/{session_id}/shared [get]
func (h *Handler) GetSharedSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	type Req struct {
		Touch bool   `form:"touch" json:"touch"` // 尝试获取而不抛出错误
		Code  string `form:"code" json:"code"`
	}
	var req Req
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	session, err := h.Helper.GetSharedSession(uri.SessionId, req.Code)
	if err != nil {
		if !req.Touch {
			ctx_utils.BizError(c, err) // 返回错误
		} else {
			ctx_utils.SuccessBizError(c, err) // 返回200但错误
		}
		return
	}

	// 优先使用共享会话的标题
	if session.Session != nil && session.ShareInfo.Title != "" {
		session.Session.Name = session.ShareInfo.Title
	}

	finalSession := schema.UserSession{
		Session: session.Session,
		UserID:  session.UserID,
	}
	ctx_utils.Success(c, finalSession)
}

// GetSessions
//
//	@Summary		获取会话列表
//	@Description	获取会话列表
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort															true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedContinuationResponse[schema.UserSession]]	"返回数据"
//	@Router			/chat/session/list [get]
func (h *Handler) GetSessions(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 查询消息
	sessions, nextPage, err := h.Store.GetSessionsByPage(ctx_utils.GetUserId(c), req)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedContinuationResponse[schema.UserSession]{
			List:     sessions,
			NextPage: nextPage,
		},
	)
}

// SyncSessions
//
//	@Summary		同步会话列表
//	@Description	同步会话列表
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			req	query		chat.SyncSessions.syncSessionParam											true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedSyncListResponse[schema.UserSession]]	"返回数据"
//	@Router			/chat/session/sync [get]
func (h *Handler) SyncSessions(c *gin.Context) {
	type syncSessionParam struct {
		entity.ParamPagingSort
		LastSyncTime entity.MilliTime `json:"last_sync_time" form:"last_sync_time" swaggertype:"primitive,integer" binding:"required"` // 客户端上次同步时间戳
	}
	var req syncSessionParam
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 查询消息
	sessions, nextPage, err := h.Store.GetSessionsForSync(
		ctx_utils.GetUserId(c),
		req.LastSyncTime.Time,
		req.ParamPagingSort,
	)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	var deletedSessions []schema.UserSession
	var updatedSessions []schema.UserSession
	for _, session := range sessions {
		if session.DeletedAt.Valid {
			session.Session = nil
			deletedSessions = append(deletedSessions, session)
		} else {
			updatedSessions = append(updatedSessions, session)
		}
	}

	ctx_utils.Success(
		c, &entity.PaginatedSyncListResponse[schema.UserSession]{
			Updated:  updatedSessions,
			Deleted:  deletedSessions,
			NextPage: nextPage,
		},
	)
}

// UpdateSession
//
//	@Summary		更新会话
//	@Description	更新会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string									true	"会话 ID"
//	@Param			req			body		entity.ReqUpdateBody[schema.Session]	true	"会话信息"
//	@Success		200			{object}	entity.CommonResponse[bool]
//	@Router			/chat/session/update/{session_id} [post]
func (h *Handler) UpdateSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req entity.ReqUpdateBody[schema.Session]
	if err := c.BindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	req.Data.ID = uri.SessionId
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.BizError(c, constants.BizErrNoPermission)
		return
	}
	req.WithWhitelist("name", "system_prompt")
	if err := h.Db.Omit("LastActive").Select(req.Updates).Updates(&req.Data).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update session")
		return
	}
	ctx_utils.Success(c, true)
}

// UpdateSessionFlag
//
//	@Summary		更新用户会话标记
//	@Description	更新用户会话标记
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string					true	"会话 ID"
//	@Param			req			body		schema.SessionFlagInfo	true	"会话信息"
//	@Success		200			{object}	entity.CommonResponse[bool]
//	@Router			/chat/session/flag/{session_id} [post]
func (h *Handler) UpdateSessionFlag(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req schema.SessionFlagInfo
	if err := c.BindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	userKey := schema.UserSession{
		SessionID: uri.SessionId,
		UserID:    ctx_utils.GetUserId(c),
	}
	updateData := map[string]interface{}{
		"flag_star": req.Star,
	}
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.BizError(c, constants.BizErrNoPermission)
		return
	}
	if err := h.Db.Model(&schema.UserSession{}).Where(userKey).Updates(&updateData).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update session flags")
		return
	}
	ctx_utils.Success(c, true)
}

// ShareSession
//
//	@Summary		分享会话
//	@Description	分享会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string							true	"会话 ID"
//	@Param			req			body		chat.ShareSession.ShareRequest	true	"分享信息"
//	@Success		200			{object}	entity.CommonResponse[bool]
//	@Router			/chat/session/share/{session_id} [post]
func (h *Handler) ShareSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	type ShareRequest struct {
		ShareInfo schema.SessionShareInfo `json:"share_info"`
		Active    bool                    `json:"active"`
	}
	var req ShareRequest
	if err := c.BindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.BizError(c, constants.BizErrNoPermission)
		return
	}
	// 停用分享时，清除邀请码和过期时间
	if !req.Active {
		req.ShareInfo.Permanent = false
		req.ShareInfo.Code = ""
		req.ShareInfo.Title = ""
		req.ShareInfo.ExpiredAt = time.Unix(0, 0).Unix()
	}
	userSession := &schema.UserSession{
		SessionID: uri.SessionId,
		UserID:    ctx_utils.GetUserId(c),
		ShareInfo: req.ShareInfo,
	}
	if err := h.Store.UpdateUserSessionShare(userSession); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update share info")
		return
	}
	ctx_utils.Success(c, true)
}
