package exam

import (
	"context"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"strconv"

	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

// ExamHandler 考试处理器
type ExamHandler struct {
	handlers.BaseHandler
	examScoreService *services.ExamScoreService
}

// NewExamHandler 创建考试处理器
func NewExamHandler(handler *handlers.BaseHandler) *ExamHandler {
	return &ExamHandler{
		BaseHandler:      *handler,
		examScoreService: services.NewExamScoreService(handler.Db),
	}
}

// SubmitExam 提交考试
//
//	@Summary		提交考试答案
//	@Description	提交用户的考试答案并进行评分
//	@Tags			考试
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"考试ID"
//	@Param			request	body		SubmitExamRequest	true	"提交信息"
//	@Success		200		{object}	entity.CommonResponse[SubmitExamResponse]
//	@Router			/tue/exam/{id}/submit [post]
func (h *ExamHandler) SubmitExam(c *gin.Context) {
	// 获取考试ID
	examIDStr := c.Param("id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 获取提交请求
	var req SubmitExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 获取用户ID
	userID := ctx_utils.GetUserId(c)

	// 查询考试信息
	exam, err := h.Store.GetExamWithDetails(examID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 检查用户是否已经提交过该考试
	//var existingRecord schema.ExamUserRecord
	//if err := h.db.Where("user_id = ? AND exam_id = ?", userID, examID).First(&existingRecord).Error; err == nil {
	//	// 已存在记录
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "已经提交过该考试"})
	//	return
	//}

	// 创建用户答案列表
	userAnswers := make([]schema.ExamUserRecordAnswer, 0, len(req.Answers))
	for _, answer := range req.Answers {
		// 查找题目
		_, exists := slice.FindBy(
			exam.Problems, func(_ int, problem schema.ExamProblem) bool {
				return problem.ProblemID == answer.ProblemID
			},
		)

		if !exists {
			continue
		}

		// 添加用户答案
		userAnswers = append(
			userAnswers, schema.ExamUserRecordAnswer{
				ProblemID: answer.ProblemID,
				Answer:    answer.Answer,
				Status:    schema.StatusPending,
			},
		)
	}
	// 过滤重复 answer（防攻击）
	userAnswers = slice.UniqueBy(
		userAnswers, func(item schema.ExamUserRecordAnswer) uint64 {
			return item.ProblemID
		},
	)

	// 创建考试记录
	record := schema.ExamUserRecord{
		UserID:    userID,
		ExamID:    examID,
		Status:    schema.StatusPending,
		Answers:   userAnswers,
		TimeSpent: req.TimeSpent,
	}

	// 保存记录
	if err := gorm_utils.Save(h.Db, &record); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 异步执行评分
	go func() {
		h.examScoreService.ScoreExam(context.Background(), record.ID)
	}()

	// 返回成功
	ctx_utils.Success(c, SubmitExamResponse{RecordID: record.ID})
}

// GetExamResult 获取考试结果
//
//	@Summary		获取考试结果
//	@Description	获取用户的考试评分结果
//	@Tags			考试
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"考试记录ID"
//	@Success		200	{object}	entity.CommonResponse[schema.ExamUserRecord]
//	@Router			/tue/exam/{id}/records [get]
func (h *ExamHandler) GetExamResult(c *gin.Context) {
	// 获取记录ID
	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 64)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 获取用户ID
	userID := ctx_utils.GetUserId(c)

	// 查询考试记录
	var record schema.ExamUserRecord
	if err := h.Db.Preload("Exam").Preload("Answers").Where(
		"id = ? AND user_id = ?",
		recordID,
		userID,
	).First(&record).Error; err != nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}

	// 返回记录
	ctx_utils.Success(c, record)
}

// RescoreExam 重新评分
//
//	@Summary		重新评分考试
//	@Description	管理员重新评分考试
//	@Tags			考试
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"考试记录ID"
//	@Router			/tue/exam/{id}/rescore [post]
func (h *ExamHandler) RescoreExam(c *gin.Context) {
	// 获取记录ID
	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 64)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 查询考试记录
	record, err := gorm_utils.GetByID[schema.ExamUserRecord](h.Db, recordID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}

	// 异步执行评分
	go func() {
		h.examScoreService.ScoreExam(context.Background(), record.ID)
	}()

	// 返回成功
	ctx_utils.Success(c, true)
}

type SubmitExamRequestAnswer struct {
	ProblemID uint64      `json:"problem_id"`
	Answer    interface{} `json:"answer"`
}

// SubmitExamRequest 提交考试请求
type SubmitExamRequest struct {
	Answers   []SubmitExamRequestAnswer `json:"answers"`    // 答案列表
	TimeSpent int                       `json:"time_spent"` // 答题用时（秒）
}

// SubmitExamResponse 提交考试响应
type SubmitExamResponse struct {
	RecordID uint64 `json:"record_id"` // 记录ID
}
