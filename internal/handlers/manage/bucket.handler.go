package manage

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateBucket
//
//	@Summary		创建 储存桶
//	@Description	创建 储存桶
//	@Tags			Bucket
//	@Accept			json
//	@Produce		json
//	@Param			bucket	body		schema.Bucket							true	"储存桶参数"
//	@Success		200		{object}	entity.CommonResponse[schema.Bucket]	"成功创建的 储存桶"
//	@Router			/manage/bucket/create [post]
func (h *Handler) CreateBucket(c *gin.Context) {
	var bucket schema.Bucket
	if err := c.ShouldBindJSON(&bucket); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Db.Create(&bucket).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create bucket")
		return
	}
	ctx_utils.Success(c, bucket)
}

// GetBucket
//
//	@Summary		获取 储存桶
//	@Description	获取 储存桶
//	@Tags			Bucket
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64									true	"储存桶 ID"
//	@Success		200	{object}	entity.CommonResponse[schema.Bucket]	"储存桶"
//	@Router			/manage/bucket/{id} [get]
func (h *Handler) GetBucket(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	bucket, err := gorm_utils.GetByID[schema.Bucket](h.Db, uri.ID)
	if err != nil {
		ctx_utils.CustomError(c, 404, "bucket not found")
		return
	}
	ctx_utils.Success(c, bucket)
}

// GetBuckets
//
//	@Summary		批量获取 储存桶
//	@Description	批量获取 储存桶
//	@Tags			Bucket
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Bucket]]	"储存桶列表"
//	@Router			/manage/bucket/list [get]
func (h *Handler) GetBuckets(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	req.SortParam.WithDefault("created_at ASC", "id")
	buckets, total, err := gorm_utils.GetByPageTotal[schema.Bucket](
		h.Db,
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get buckets")
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Bucket]{
			List:  buckets,
			Total: total,
		},
	)
}

// UpdateBucket
//
//	@Summary		更新 储存桶
//	@Description	更新 储存桶
//	@Tags			Bucket
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uint64								true	"储存桶 ID"
//	@Param			bucket	body		entity.ReqUpdateBody[schema.Bucket]	true	"储存桶参数"
//	@Success		200		{object}	entity.CommonResponse[bool]			"更新成功与否"
//	@Router			/manage/bucket/{id}/update [post]
func (h *Handler) UpdateBucket(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var bucket entity.ReqUpdateBody[schema.Bucket]
	if err := c.ShouldBindJSON(&bucket); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	bucket.Data.ID = uri.ID
	if err := h.Db.Select(bucket.Updates).Updates(&bucket.Data).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update bucket")
		return
	}
	ctx_utils.Success(c, true)
}

// DeleteBucket
//
//	@Summary		删除 储存桶
//	@Description	删除 储存桶
//	@Tags			Bucket
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"储存桶 ID"
//	@Success		200	{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/bucket/{id}/delete [post]
func (h *Handler) DeleteBucket(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Delete[schema.Bucket](h.Db, uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete bucket")
		return
	}
	ctx_utils.Success(c, true)
}
