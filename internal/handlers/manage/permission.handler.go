package manage

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"net/http"
)

// GetPermissions
//
//	@Summary		批量分页获取权限
//	@Description	批量分页获取权限
//	@Tags			Permission
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort													true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Permission]]	"权限列表"
//	@Router			/manage/permission/list [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	var param entity.ParamPagingSort
	if err := c.ShouldBindQuery(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	param.PagingParam.WithMaxSize(10000)
	permissions, total, err := gorm_utils.GetByPageTotal[schema.Permission](
		h.Db.Preload(clause.Associations),
		param.PagingParam,
		param.SortParam,
	)
	// 过滤掉密码
	permissions = slice.Map(
		permissions, func(index int, permission schema.Permission) schema.Permission {
			return permission
		},
	)

	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Permission]{
			List:  permissions,
			Total: total,
		},
	)
}

// GetPermission
//
//	@Summary		获取权限
//	@Description	获取权限
//	@Tags			Permission
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64										true	"权限 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.Permission]	"权限"
//	@Router			/manage/permission/{id} [get]
func (h *Handler) GetPermission(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	permission, err := gorm_utils.GetByID[schema.Permission](h.Db.Preload(clause.Associations), uri.ID)
	if err != nil {
		ctx_utils.CustomError(c, 404, "permission not found")
		return
	}
	ctx_utils.Success(c, permission)
}

// UpdatePermission
//
//	@Summary		更新权限
//	@Description	更新权限
//	@Tags			Permission
//	@Accept			json
//	@Produce		json
//	@Param			id			path		uint64						true	"权限 ID"
//	@Param			permission	body		schema.Permission			true	"权限参数"
//	@Success		200			{object}	entity.CommonResponse[bool]	"更新成功与否"
//	@Router			/manage/permission/{id}/update [post]
func (h *Handler) UpdatePermission(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var permission schema.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	permission.ID = uri.ID
	updatePermission := schema.Permission{
		ID:     permission.ID,
		Active: permission.Active, // 更新时，只更新 Active 字段
	}
	if err := h.Db.Select("Active").Updates(&updatePermission); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update permission")
		return
	}
	ctx_utils.Success(c, true)
}
