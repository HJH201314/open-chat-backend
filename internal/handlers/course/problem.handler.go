package course

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
)

// CreateProblem 创建单个题目
//
//	@Summary		创建单个题目
//	@Description	创建单个题目
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			req	body		schema.Problem							true	"题目结构"
//	@Success		200	{object}	entity.CommonResponse[schema.Problem]	"返回数据"
//	@Router			/tue/problem/create [post]
func (h *Handler) CreateProblem(c *gin.Context) {
	// 从 path 中获取题目 ID
	var req schema.Problem
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 查询题目
	if err := gorm_utils.Save[schema.Problem](h.Db, &req); err != nil {
		ctx_utils.CustomError(c, 500, "create problem failed")
		return
	}
	ctx_utils.Success(c, req)
}

// GetProblem 获取单个题目
//
//	@Summary		获取单个题目
//	@Description	获取单个题目
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string									true	"题目 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.Problem]	"返回数据"
//	@Router			/tue/problem/{id} [get]
func (h *Handler) GetProblem(c *gin.Context) {
	// 从 path 中获取题目 ID
	var uri PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 查询题目
	problem, err := gorm_utils.GetByID[schema.Problem](h.Db, uri.ID)
	if err != nil {
		ctx_utils.CustomError(c, 404, "problem not found")
		return
	}
	ctx_utils.Success(c, problem)
}

// GetProblems 分页获取题目列表
//
//	@Summary		分页获取题目列表
//	@Description	分页获取题目列表
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort														true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedContinuationResponse[schema.Problem]]	"返回数据"
//	@Router			/tue/problem/list [get]
func (h *Handler) GetProblems(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	problems, total, err := gorm_utils.GetByPageTotal[schema.Problem](h.Db, req.PagingParam, req.SortParam)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Problem]{
			List:  problems,
			Total: total,
		},
	)
}
