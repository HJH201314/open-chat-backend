package course

import (
	"github.com/fcraft/open-chat/internal/constants"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
)

// GetExam 获取单个测验
//
//	@Summary		获取单个测验
//	@Description	获取单个测验
//	@Tags			Exam
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string								true	"测验 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.Exam]	"返回数据"
//	@Router			/tue/exam/{id} [get]
func (h *Handler) GetExam(c *gin.Context) {
	var param PathParamId
	if err := c.BindUri(&param); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	exam, err := h.Helper.GetExam(param.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}

	ctx_utils.Success(c, exam)
}

// CreateExam 保存单个测验
//
//	@Summary		保存单个测验
//	@Description	保存单个测验
//	@Tags			Exam
//	@Accept			json
//	@Produce		json
//	@Param			req	body		schema.Exam							true	"测验内容"
//	@Success		200	{object}	entity.CommonResponse[schema.Exam]	"返回数据"
//	@Router			/tue/exam/create [post]
func (h *Handler) CreateExam(c *gin.Context) {
	var req schema.Exam
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 1. 保存测验
	if err := gorm_utils.Save(h.Store.Db, &req); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 2. 更新测验总分
	err := h.Store.UpdateExamTotalScore(req.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, req)
}

// RandomExam 随机测验
//
//	@Summary		随机测验
//	@Description	随机测验
//	@Tags			Exam
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[schema.Exam]	"返回数据"
//	@Router			/tue/exam/random [post]
func (h *Handler) RandomExam(c *gin.Context) {
	var problems []schema.Problem

	// 1. 查询随机题目
	if err := h.Db.Order("RANDOM()").Limit(20).Find(&problems).Error; err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 2. 生成测验
	var examProblems []schema.ExamProblem
	for _, problem := range problems {
		examProblems = append(
			examProblems, schema.ExamProblem{
				ProblemID: problem.ID,
				Score:     100,
			},
		)
	}
	req := schema.Exam{
		Name:        "智能测验",
		Description: "智能测验",
		LimitTime:   0,
		Problems:    examProblems,
		TotalScore:  uint64(len(examProblems) * 100),
		Subjects:    "",
	}
	if err := gorm_utils.Save(h.Store.Db, &req); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 3. 重新查询结果
	exam, err := h.Store.GetExamWithDetails(req.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, exam)
}
