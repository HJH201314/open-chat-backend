package course

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ProblemHandler 考试处理器
type ProblemHandler struct {
	handlers.BaseHandler
	makeQuestionService *services.MakeQuestionService
}

// NewProblemHandler 创建考试处理器
func NewProblemHandler(handler *handlers.BaseHandler) *ProblemHandler {
	return &ProblemHandler{
		BaseHandler:         *handler,
		makeQuestionService: services.GetMakeQuestionService(),
	}
}

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
func (h *ProblemHandler) CreateProblem(c *gin.Context) {
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
func (h *ProblemHandler) GetProblem(c *gin.Context) {
	// 从 path 中获取题目 ID
	var uri entity.PathParamId
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
//	@Param			req	query		entity.ParamPagingSort													true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Problem]]	"返回数据"
//	@Router			/tue/problem/list [get]
func (h *ProblemHandler) GetProblems(c *gin.Context) {
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

// UpdateProblem
//
//	@Summary		更新题目
//	@Description	更新题目
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uint64									true	"题目 ID"
//	@Param			problem	body		entity.ReqUpdateBody[schema.Problem]	true	"题目参数"
//	@Success		200		{object}	entity.CommonResponse[bool]				"更新成功与否"
//	@Router			/tue/problem/{id}/update [post]
func (h *ProblemHandler) UpdateProblem(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var problem entity.ReqUpdateBody[schema.Problem]
	if err := c.ShouldBindJSON(&problem); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	problem.Data.ID = uri.ID
	if err := h.Db.Select(problem.Updates).Updates(&problem.Data).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update problem")
		return
	}
	ctx_utils.Success(c, true)
}

// DeleteProblem
//
//	@Summary		删除题目
//	@Description	删除题目
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"题目 ID"
//	@Success		200	{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/tue/problem/{id}/delete [post]
func (h *ProblemHandler) DeleteProblem(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Delete[schema.Problem](h.Db, uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete problem")
		return
	}
	ctx_utils.Success(c, true)
}

// MakeQuestion 创建题目
//
//	@Summary		创建题目
//	@Description	创建题目
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			req	body		MakeQuestionRequest						true	"题目要求"
//	@Success		200	{object}	entity.CommonResponse[schema.Problem]	"生成的题目"
//	@Router			/tue/problem/make [post]
func (h *ProblemHandler) MakeQuestion(c *gin.Context) {
	var req MakeQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	problem, err := services.GetMakeQuestionService().MakeQuestion(req.Type, req.Description)
	if err != nil {
		ctx_utils.CustomError(c, 500, "make question failed")
		return
	}
	ctx_utils.Success(c, &problem)
}

type MakeQuestionRequest struct {
	Type        schema.ProblemType `json:"type" binding:"required"`
	Description string             `json:"description" binding:"required"`
}
