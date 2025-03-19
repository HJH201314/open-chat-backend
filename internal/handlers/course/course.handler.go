package course

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

// GetCourse 获取单个课程
//
//	@Summary		获取单个课程
//	@Description	获取单个课程
//	@Tags			Course
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string									true	"课程 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.Course]	"返回数据"
//	@Router			/tue/course/{id} [get]
func (h *Handler) GetCourse(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	course, err := h.Store.GetCourseWithDetails(param.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}

	ctx_utils.Success(c, course)
}

// CreateCourse 创建课程
//
//	@Summary		创建课程
//	@Description	创建课程基础参数，绑定或创建题目、资源
//	@Tags			Course
//	@Accept			json
//	@Produce		json
//	@Param			req	body		schema.Course							true	"课程内容"
//	@Success		200	{object}	entity.CommonResponse[schema.Course]	"返回数据"
//	@Router			/tue/course/create [post]
func (h *Handler) CreateCourse(c *gin.Context) {
	var req schema.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	if err := h.Store.CreateCourse(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, req)
}

// UpdateCourse 更新课程
//
//	@Summary		更新课程
//	@Description	更新课程基础参数，增量更新 题目、资源
//	@Tags			Course
//	@Accept			json
//	@Produce		json
//	@Param			req	body		schema.Course							true	"课程内容"
//	@Success		200	{object}	entity.CommonResponse[schema.Course]	"返回数据"
//	@Router			/tue/course/update [post]
func (h *Handler) UpdateCourse(c *gin.Context) {
	var req schema.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	if err := h.Store.UpdateCourse(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, req)
}

// DeleteCourse 删除课程
//
//	@Summary		删除课程
//	@Description	删除课程
//	@Tags			Course
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"课程 ID"
//	@Success		200	{object}	entity.CommonResponse[any]	"返回数据"
//	@Router			/tue/course/{id} [post]
func (h *Handler) DeleteCourse(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	if err := h.Store.DeleteCourse(param.ID); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, true)
}

// GetCourses 获取课程列表
//
//	@Summary		获取课程列表
//	@Description	获取课程列表
//	@Tags			Course
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Course]]	"返回数据"
//	@Router			/tue/course/list [get]
func (h *Handler) GetCourses(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	req.WithDefaultSize(20).WithMaxSize(100)
	courses, lastPage, err := h.Store.GetCourses(req.PagingParam, req.SortParam)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Course]{
			List:     courses,
			LastPage: &lastPage,
		},
	)
}
