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

// CreateModel
//
//	@Summary		创建模型
//	@Description	创建模型并绑定到 API 供应商
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			model	body		schema.Model							true	"模型"
//	@Success		200		{object}	entity.CommonResponse[schema.Model ]	"成功创建的模型"
//	@Router			/manage/model/create [post]
func (h *Handler) CreateModel(c *gin.Context) {
	var model schema.Model
	if err := c.ShouldBindJSON(&model); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Store.AddModel(&model); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create model")
		return
	}
	ctx_utils.Success(c, model)
}

// GetModel
//
//	@Summary		获取模型
//	@Description	获取模型
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			model_id	path		uint64								true	"Model ID"
//	@Success		200			{object}	entity.CommonResponse[schema.Model]	"模型"
//	@Router			/manage/model/{model_id} [get]
func (h *Handler) GetModel(c *gin.Context) {
	var uri struct {
		ModelId uint64 `uri:"model_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ModelId == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	model, err := h.Store.GetModel(uri.ModelId)
	if err != nil {
		ctx_utils.CustomError(c, 404, "model not found")
		return
	}
	ctx_utils.Success(c, model)
}

// GetModels
//
//	@Summary		批量获取模型
//	@Description	批量获取模型
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Model]]	"模型列表"
//	@Router			/manage/model/list [get]
func (h *Handler) GetModels(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	models, total, err := gorm_utils.GetByPageTotal[schema.Model](
		h.Db.Preload("Provider"),
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get model")
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Model]{
			List:  models,
			Total: total,
		},
	)
}

// GetModelsByProvider
//
//	@Summary		获取 API 提供商的模型
//	@Description	获取 API 提供商的模型
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			provider_id	path		uint64																true	"Provider ID"
//	@Param			req			query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200			{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Model]]	"模型列表"
//	@Router			/manage/model/provider/{provider_id} [get]
func (h *Handler) GetModelsByProvider(c *gin.Context) {
	var uri struct {
		ProviderId uint64 `uri:"provider_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ProviderId == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	models, total, err := gorm_utils.GetByPageTotal[schema.Model](
		h.Db.Preload("Provider").Where("provider_id = ?", uri.ProviderId),
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get model")
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Model]{
			List:  models,
			Total: total,
		},
	)
}

// UpdateModel
//
//	@Summary		更新模型
//	@Description	Update 更新模型，若参数不传入或为空，则不会更新
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			model	body		schema.Model				true	"模型"
//	@Success		200		{object}	entity.CommonResponse[bool]	"更新成功与否"
//	@Router			/manage/model/update [post]
func (h *Handler) UpdateModel(c *gin.Context) {
	var model schema.Model
	if err := c.ShouldBindJSON(&model); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Store.UpdateModel(&model); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update model")
		return
	}
	ctx_utils.Success(c, true)
}

// DeleteModel
//
//	@Summary		删除模型
//	@Description	删除模型
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			model_id	path		uint64						true	"Model ID"
//	@Success		200			{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/model/delete/{model_id} [post]
func (h *Handler) DeleteModel(c *gin.Context) {
	var uri struct {
		ModelId uint64 `uri:"model_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ModelId == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Store.DeleteModel(uri.ModelId); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete model")
		return
	}
	ctx_utils.Success(c, true)
}

///////////////////////

// CreateModelCollection
//
//	@Summary		创建模型集合
//	@Description	创建模型集合
//	@Tags			ModelCollection
//	@Accept			json
//	@Produce		json
//	@Param			model	body		schema.ModelCollection							true	"模型集合"
//	@Success		200		{object}	entity.CommonResponse[schema.ModelCollection]	"成功创建的模型集合"
//	@Router			/manage/collection/create [post]
func (h *Handler) CreateModelCollection(c *gin.Context) {
	var collection schema.ModelCollection
	if err := c.ShouldBindJSON(&collection); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Save[schema.ModelCollection](h.Db, &collection); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create model collection")
		return
	}
	ctx_utils.Success(c, collection)
}

// GetModelCollection
//
//	@Summary		获取模型集合
//	@Description	获取模型集合
//	@Tags			ModelCollection
//	@Accept			json
//	@Produce		json
//	@Param			collection_id	path		uint64											true	"ModelCollection ID"
//	@Success		200				{object}	entity.CommonResponse[schema.ModelCollection]	"模型"
//	@Router			/manage/collection/{collection_id} [get]
func (h *Handler) GetModelCollection(c *gin.Context) {
	var uri struct {
		CollectionID uint64 `uri:"collection_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.CollectionID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	model, err := gorm_utils.GetByID[schema.ModelCollection](h.Db.Preload("Models"), uri.CollectionID)
	if err != nil {
		ctx_utils.CustomError(c, 404, "model collection not found")
		return
	}
	ctx_utils.Success(c, model)
}

// GetModelCollections
//
//	@Summary		批量获取模型集合
//	@Description	批量获取模型集合
//	@Tags			ModelCollection
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort															true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.ModelCollection]]	"模型集合列表"
//	@Router			/manage/collection/list [get]
func (h *Handler) GetModelCollections(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	models, total, err := gorm_utils.GetByPageTotal[schema.ModelCollection](
		h.Db.Preload("Models"),
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get model collection")
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.ModelCollection]{
			List:  models,
			Total: total,
		},
	)
}

// DeleteModelCollection
//
//	@Summary		删除模型集合
//	@Description	删除模型集合
//	@Tags			Model
//	@Accept			json
//	@Produce		json
//	@Param			collection_id	path		uint64						true	"ModelCollection ID"
//	@Success		200				{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/collection/delete/{collection_id} [post]
func (h *Handler) DeleteModelCollection(c *gin.Context) {
	var uri struct {
		CollectionID uint64 `uri:"collection_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.CollectionID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := gorm_utils.Delete[schema.ModelCollection](h.Db, uri.CollectionID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete model collection")
		return
	}
	ctx_utils.Success(c, true)
}
