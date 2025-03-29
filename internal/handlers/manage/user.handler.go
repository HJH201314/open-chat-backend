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

// GetUsers
//
//	@Summary		批量分页获取用户
//	@Description	批量分页获取用户
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.User]]	"用户列表"
//	@Router			/manage/user/list [get]
func (h *Handler) GetUsers(c *gin.Context) {
	var param entity.ParamPagingSort
	if err := c.ShouldBindQuery(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	users, total, err := gorm_utils.GetByPageTotal[schema.User](
		h.Db.Preload(clause.Associations),
		param.PagingParam,
		param.SortParam,
	)
	// 过滤掉密码
	users = slice.Map(
		users, func(index int, user schema.User) schema.User {
			user.Password = ""
			return user
		},
	)

	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.User]{
			List:  users,
			Total: total,
		},
	)
}

// CreateUser
//
//	@Summary		创建用户
//	@Description	创建用户
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			user	body		schema.User								true	"用户参数"
//	@Success		200		{object}	entity.CommonResponse[schema.Provider]	"成功创建的用户"
//	@Router			/manage/user/create [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var user schema.User
	if err := c.ShouldBindJSON(&user); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Save(h.Db, &user); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create user")
		return
	}
	ctx_utils.Success(c, user)
}

// GetUser
//
//	@Summary		获取用户
//	@Description	获取用户
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64								true	"用户 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.User]	"用户"
//	@Router			/manage/user/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	user, err := gorm_utils.GetByID[schema.User](h.Db.Preload(clause.Associations), uri.ID)
	if err != nil {
		ctx_utils.CustomError(c, 404, "user not found")
		return
	}
	user.Password = ""
	ctx_utils.Success(c, user)
}

// UpdateUser
//
//	@Summary		更新用户
//	@Description	更新用户
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uint64						true	"用户 ID"
//	@Param			user	body		schema.User					true	"用户参数"
//	@Success		200		{object}	entity.CommonResponse[bool]	"更新成功与否"
//	@Router			/manage/user/{id}/update [post]
func (h *Handler) UpdateUser(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var user schema.User
	if err := c.ShouldBindJSON(&user); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	user.ID = uri.ID
	if user.Roles != nil {
		// 更新关联的 role
		if err := h.Db.Model(&user).Association("Roles").Replace(user.Roles); err != nil {
			ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update user role")
			return
		}
		// 删除用户角色缓存
		if err := h.Redis.DeleteUserRolesCache(uri.ID); err != nil {
			ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete user role cache: "+err.Error())
			return
		}
	}
	// 更新用户信息
	if err := h.Db.Model(&user).Omit("Username", "Roles").Updates(&user).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update user: "+err.Error())
		return
	}
	ctx_utils.Success(c, true)
}

// DeleteUser
//
//	@Summary		删除用户
//	@Description	删除用户
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"用户 ID"
//	@Success		200	{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/user/{id}/delete [post]
func (h *Handler) DeleteUser(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Delete[schema.User](h.Db.Select(clause.Associations), uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete user")
		return
	}
	if _, err := h.Redis.InvalidUserAllToken(uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to logout user")
		return
	}
	ctx_utils.Success(c, true)
}

// LogoutUser
//
//	@Summary		强制登出用户
//	@Description	强制登出用户
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"用户 ID"
//	@Success		200	{object}	entity.CommonResponse[int]	"成功登出的设备数量"
//	@Router			/manage/user/{id}/logout [post]
func (h *Handler) LogoutUser(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	count, err := h.Redis.InvalidUserAllToken(uri.ID)
	if err != nil {
		return
	}
	ctx_utils.Success(c, count)
}
