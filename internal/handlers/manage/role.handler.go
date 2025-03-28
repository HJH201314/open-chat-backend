package manage

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
)

// GetRoles
//
//	@Summary		批量分页获取角色
//	@Description	批量分页获取角色
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Role]]	"角色列表"
//	@Router			/manage/role/list [get]
func (h *Handler) GetRoles(c *gin.Context) {
	var param entity.ParamPagingSort
	if err := c.ShouldBindQuery(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	roles, total, err := gorm_utils.GetByPageTotal[schema.Role](
		h.Db.Preload(clause.Associations),
		param.PagingParam,
		param.SortParam,
	)
	// 过滤掉密码
	roles = slice.Map(
		roles, func(index int, role schema.Role) schema.Role {
			return role
		},
	)

	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Role]{
			List:  roles,
			Total: total,
		},
	)
}

// CreateRole
//
//	@Summary		创建角色
//	@Description	创建角色
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			role	body		schema.Role								true	"角色参数"
//	@Success		200		{object}	entity.CommonResponse[schema.Provider]	"成功创建的角色"
//	@Router			/manage/role/create [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var role schema.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Save(h.Db, &role); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create role")
		return
	}
	ctx_utils.Success(c, role)
}

// GetRole
//
//	@Summary		获取角色
//	@Description	获取角色
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64								true	"角色 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.Role]	"角色"
//	@Router			/manage/role/{id} [get]
func (h *Handler) GetRole(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	role, err := gorm_utils.GetByID[schema.Role](h.Db.Preload(clause.Associations), uri.ID)
	if err != nil {
		ctx_utils.CustomError(c, 404, "role not found")
		return
	}
	ctx_utils.Success(c, role)
}

// UpdateRole
//
//	@Summary		更新角色
//	@Description	更新角色
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uint64						true	"角色 ID"
//	@Param			role	body		schema.Role					true	"角色参数"
//	@Success		200		{object}	entity.CommonResponse[bool]	"更新成功与否"
//	@Router			/manage/role/{id}/update [post]
func (h *Handler) UpdateRole(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var role schema.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	role.ID = uri.ID
	if err := gorm_utils.Update(h.Db.Session(&gorm.Session{FullSaveAssociations: true}), &role); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update role")
		return
	}
	ctx_utils.Success(c, true)
}

// DeleteRole
//
//	@Summary		删除角色
//	@Description	删除角色
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"角色 ID"
//	@Success		200	{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/role/{id}/delete [post]
func (h *Handler) DeleteRole(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Delete[schema.Role](h.Db.Select(clause.Associations), uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete role")
		return
	}
	ctx_utils.Success(c, true)
}
