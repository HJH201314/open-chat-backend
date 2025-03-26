package chat

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"

	"github.com/gin-gonic/gin"
)

// CreatePreset
//
//	@Summary		创建预设
//	@Description	创建一个新的预设，包含名称、描述和引用的会话ID
//	@Tags			Preset
//	@Accept			json
//	@Produce		json
//	@Param			role	body		schema.Preset	true	"预设信息"
//	@Success		200		{object}	schema.Preset	"成功创建的预设"
//	@Router			/preset/create [post]
func (h *Handler) CreatePreset(c *gin.Context) {
	var role schema.Preset
	if err := c.ShouldBindJSON(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 创建角色
	if err := h.Helper.CreatePreset(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, role)
}

// GetPreset
//
//	@Summary		获取预设
//	@Description	根据ID获取指定的预设信息
//	@Tags			Preset
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"预设ID"
//	@Success		200	{object}	schema.Preset	"预设信息"
//	@Router			/preset/{id} [get]
func (h *Handler) GetPreset(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 获取角色
	role, err := h.Helper.GetPreset(param.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}

	ctx_utils.Success(c, role)
}

// ListPresets
//
//	@Summary		获取预设列表
//	@Description	获取所有预设的列表
//	@Tags			Preset
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	schema.Preset	"预设列表"
//	@Router			/preset/list [get]
func (h *Handler) ListPresets(c *gin.Context) {
	// 获取角色列表
	roles, err := h.Helper.ListPresets()
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, roles)
}

// UpdatePreset
//
//	@Summary		更新预设
//	@Description	更新指定ID的预设信息
//	@Tags			Preset
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"预设ID"
//	@Param			role	body		schema.Preset	true	"更新的预设信息"
//	@Success		200		{object}	schema.Preset	"更新后的预设信息"
//	@Router			/preset/{id}/update [post]
func (h *Handler) UpdatePreset(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	var role schema.Preset
	if err := c.ShouldBindJSON(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 确保ID一致
	role.ID = param.ID

	// 更新角色
	if err := h.Helper.UpdatePreset(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, role)
}

// DeletePreset
//
//	@Summary		删除预设
//	@Description	删除指定ID的预设
//	@Tags			Preset
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int		true	"预设ID"
//	@Success		200	{object}	bool	"删除成功"
//	@Router			/preset/{id}/delete [post]
func (h *Handler) DeletePreset(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 删除角色
	if err := h.Helper.DeletePreset(param.ID); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, true)
}
