package manage

import (
	"github.com/fcraft/open-chat/internal/models"
	"github.com/fcraft/open-chat/internal/shared/constant"
	_ "github.com/fcraft/open-chat/internal/shared/entity"
	"github.com/fcraft/open-chat/internal/shared/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateProvider
//
//	@Summary		创建 API 提供商
//	@Description	创建 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			provider	body		models.Provider							true	"API 提供商参数"
//	@Success		200			{object}	entity.CommonResponse[models.Provider]	"成功创建的 API 提供商"
//	@Router			/manage/provider/create [post]
func (h *Handler) CreateProvider(c *gin.Context) {
	var provider models.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.AddProvider(&provider); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to create provider")
		return
	}
	util.NormalResponse(c, provider)
}

// GetProvider
//
//	@Summary		获取 API 提供商
//	@Description	获取 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			provider_id	path		uint64									true	"API 提供商 ID"
//	@Success		200			{object}	entity.CommonResponse[models.Provider]	"API 提供商"
//	@Router			/manage/provider/{provider_id} [get]
func (h *Handler) GetProvider(c *gin.Context) {
	var uri struct {
		ProviderId uint64 `uri:"provider_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ProviderId == 0 {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	provider, err := h.Store.GetProvider(uri.ProviderId)
	if err != nil {
		util.CustomErrorResponse(c, 404, "provider not found")
		return
	}
	util.NormalResponse(c, provider)
}

// GetProviders
//
//	@Summary		批量获取 API 提供商
//	@Description	批量获取 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[[]models.Provider]	"API 提供商列表"
//	@Router			/manage/provider/list [get]
func (h *Handler) GetProviders(c *gin.Context) {
	providers, err := h.Store.GetProviders()
	if err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to get providers")
		return
	}
	util.NormalResponse(c, providers)
}

// UpdateProvider
//
//	@Summary		更新 API 提供商
//	@Description	更新 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			provider	body		models.Provider				true	"API 提供商参数"
//	@Success		200			{object}	entity.CommonResponse[bool]	"更新成功与否"
//	@Router			/manage/provider/update [post]
func (h *Handler) UpdateProvider(c *gin.Context) {
	var provider models.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.UpdateProvider(&provider); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to update provider")
		return
	}
	util.NormalResponse(c, true)
}

// DeleteProvider
//
//	@Summary		删除 API 提供商
//	@Description	删除 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			provider_id	path		uint64						true	"API 提供商 ID"
//	@Success		200			{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/provider/delete/{provider_id} [post]
func (h *Handler) DeleteProvider(c *gin.Context) {
	var uri struct {
		ProviderId uint64 `uri:"provider_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ProviderId == 0 {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.DeleteProvider(uri.ProviderId); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to delete provider")
		return
	}
	util.NormalResponse(c, true)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	var apiKey models.APIKey
	if err := c.ShouldBindJSON(&apiKey); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.AddAPIKey(apiKey.ProviderID, apiKey.Key); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to create key")
		return
	}
	util.NormalResponse(c, true)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
	var uri struct {
		KeyId uint64 `uri:"key_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.KeyId == 0 {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.DeleteAPIKey(uri.KeyId); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to delete key")
		return
	}
	util.NormalResponse(c, true)
}

func (h *Handler) CreateModel(c *gin.Context) {
	var model models.Model
	if err := c.ShouldBindJSON(&model); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.AddModel(&model); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to create model")
		return
	}
	util.NormalResponse(c, model)
}

func (h *Handler) GetModel(c *gin.Context) {
	var uri struct {
		ModelId uint64 `uri:"model_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ModelId == 0 {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	model, err := h.Store.GetModel(uri.ModelId)
	if err != nil {
		util.CustomErrorResponse(c, 404, "model not found")
		return
	}
	util.NormalResponse(c, model)
}

func (h *Handler) GetModelsByProvider(c *gin.Context) {
	var uri struct {
		ProviderId uint64 `uri:"provider_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ProviderId == 0 {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	aiModels, err := h.Store.GetModelsByProvider(uri.ProviderId)
	if err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to get models")
		return
	}
	util.NormalResponse(c, aiModels)
}

func (h *Handler) UpdateModel(c *gin.Context) {
	var model models.Model
	if err := c.ShouldBindJSON(&model); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.UpdateModel(&model); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to update model")
		return
	}
	util.NormalResponse(c, true)
}

func (h *Handler) DeleteModel(c *gin.Context) {
	var uri struct {
		ModelId uint64 `uri:"model_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ModelId == 0 {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := h.Store.DeleteModel(uri.ModelId); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to delete model")
		return
	}
	util.NormalResponse(c, true)
}
