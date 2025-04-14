package manage

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
)

// GetSchedule
//
//	@Summary		获取 定时任务
//	@Description	获取 定时任务
//	@Tags			Schedule
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string									true	"定时任务 name"
//	@Success		200		{object}	entity.CommonResponse[schema.Schedule]	"定时任务"
//	@Router			/manage/schedule/{name} [get]
func (h *Handler) GetSchedule(c *gin.Context) {
	var uri entity.PathParamName
	if err := c.BindUri(&uri); err != nil || uri.Name == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	schedule, err := gorm_utils.GetByName[schema.Schedule](h.Db, uri.Name)
	if err != nil {
		ctx_utils.CustomError(c, 404, "schedule not found")
		return
	}

	// 获取定时任务状态
	if services.GetScheduleService().IsJobRunning(schedule.Name) {
		schedule.Status = schema.ScheduleStatusRunning
	} else {
		schedule.Status = schema.ScheduleStatusStopped
	}
	ctx_utils.Success(c, schedule)
}

// GetSchedules
//
//	@Summary		批量获取 定时任务
//	@Description	批量获取 定时任务
//	@Tags			Schedule
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort													true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Schedule]]	"定时任务列表"
//	@Router			/manage/schedule/list [get]
func (h *Handler) GetSchedules(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	req.SortParam.WithDefault("created_at ASC", "id")
	schedules, total, err := gorm_utils.GetByPageTotal[schema.Schedule](
		h.Db,
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get schedules")
		return
	}

	// 获取定时任务状态
	for i, schedule := range schedules {
		if services.GetScheduleService().IsJobRunning(schedule.Name) {
			schedules[i].Status = schema.ScheduleStatusRunning
		} else {
			schedules[i].Status = schema.ScheduleStatusStopped
		}
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Schedule]{
			List:  schedules,
			Total: total,
		},
	)
}

// UpdateSchedule
//
//	@Summary		更新 定时任务
//	@Description	更新 定时任务
//	@Tags			Schedule
//	@Accept			json
//	@Produce		json
//	@Param			name		path		string									true	"定时任务 name"
//	@Param			schedule	body		entity.ReqUpdateBody[schema.Schedule]	true	"定时任务参数"
//	@Success		200			{object}	entity.CommonResponse[bool]				"更新成功与否"
//	@Router			/manage/schedule/{name}/update [post]
func (h *Handler) UpdateSchedule(c *gin.Context) {
	var uri entity.PathParamName
	if err := c.BindUri(&uri); err != nil || uri.Name == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var schedule entity.ReqUpdateBody[schema.Schedule]
	if err := c.ShouldBindJSON(&schedule); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	schedule.Data.Name = uri.Name
	schedule.WithWhitelist("duration", "status")
	if err := h.Db.Where("name = ?", uri.Name).Select(schedule.Updates).Updates(&schedule.Data).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update schedule")
		return
	}

	// 处理定时任务启停
	if slices.Contains(schedule.Updates, "status") {
		var handler func(name string) error
		if schedule.Data.Status == schema.ScheduleStatusRunning {
			handler = services.GetScheduleService().StartJob
		} else {
			handler = services.GetScheduleService().StopJob
		}
		if err := handler(schedule.Data.Name); err != nil {
			ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update schedule")
			return
		}
	}

	ctx_utils.Success(c, true)
}

// RunSchedule
//
//	@Summary		立即运行 定时任务
//	@Description	立即运行 定时任务
//	@Tags			Schedule
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string						true	"定时任务 name"
//	@Success		200		{object}	entity.CommonResponse[bool]	"成功与否"
//	@Router			/manage/schedule/{name}/run [post]
func (h *Handler) RunSchedule(c *gin.Context) {
	var uri entity.PathParamName
	if err := c.BindUri(&uri); err != nil || uri.Name == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	err := services.GetScheduleService().RunJobNow(uri.Name)
	if err != nil {
		ctx_utils.Success(c, false)
		return
	}

	ctx_utils.Success(c, true)
}
