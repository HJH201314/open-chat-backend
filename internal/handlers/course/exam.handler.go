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

	exam, err := h.Store.GetExamWithDetails(param.ID)
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
//	@Router			/tue/exam/create [get]
func (h *Handler) CreateExam(c *gin.Context) {
	var req schema.Exam
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	if err := gorm_utils.Save(h.Store.Db, &req); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, req)
}
