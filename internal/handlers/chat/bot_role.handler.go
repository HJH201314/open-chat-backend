package chat

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"

	"github.com/gin-gonic/gin"
)

// CreateBotRole
//
//	@Summary		创建机器人角色
//	@Description	创建一个新的机器人角色，包含名称、描述和引用的会话ID
//	@Tags			BotRole
//	@Accept			json
//	@Produce		json
//	@Param			role	body		schema.BotRole	true	"机器人角色信息"
//	@Success		200		{object}	schema.BotRole	"成功创建的机器人角色"
//	@Router			/bot/create [post]
func (h *Handler) CreateBotRole(c *gin.Context) {
	var role schema.BotRole
	if err := c.ShouldBindJSON(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 创建角色
	if err := h.Helper.CreateBotRole(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, role)
}

// GetBotRole
//
//	@Summary		获取机器人角色
//	@Description	根据ID获取指定的机器人角色信息
//	@Tags			BotRole
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"机器人角色ID"
//	@Success		200	{object}	schema.BotRole	"机器人角色信息"
//	@Router			/bot/{id} [get]
func (h *Handler) GetBotRole(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 获取角色
	role, err := h.Helper.GetBotRole(param.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}

	ctx_utils.Success(c, role)
}

// ListBotRoles
//
//	@Summary		获取机器人角色列表
//	@Description	获取所有机器人角色的列表
//	@Tags			BotRole
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	schema.BotRole	"机器人角色列表"
//	@Router			/bot/list [get]
func (h *Handler) ListBotRoles(c *gin.Context) {
	// 获取角色列表
	roles, err := h.Helper.ListBotRoles()
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, roles)
}

// UpdateBotRole
//
//	@Summary		更新机器人角色
//	@Description	更新指定ID的机器人角色信息
//	@Tags			BotRole
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"机器人角色ID"
//	@Param			role	body		schema.BotRole	true	"更新的机器人角色信息"
//	@Success		200		{object}	schema.BotRole	"更新后的机器人角色信息"
//	@Router			/bot/{id}/update [post]
func (h *Handler) UpdateBotRole(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	var role schema.BotRole
	if err := c.ShouldBindJSON(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 确保ID一致
	role.ID = param.ID

	// 更新角色
	if err := h.Helper.UpdateBotRole(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, role)
}

// DeleteBotRole
//
//	@Summary		删除机器人角色
//	@Description	删除指定ID的机器人角色
//	@Tags			BotRole
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int		true	"机器人角色ID"
//	@Success		200	{object}	bool	"删除成功"
//	@Router			/bot/{id}/delete [post]
func (h *Handler) DeleteBotRole(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 删除角色
	if err := h.Helper.DeleteBotRole(param.ID); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, true)
}
