package manage

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
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
//	@Param			provider	body		schema.Provider							true	"API 提供商参数"
//	@Success		200			{object}	entity.CommonResponse[schema.Provider]	"成功创建的 API 提供商"
//	@Router			/manage/provider/create [post]
func (h *Handler) CreateProvider(c *gin.Context) {
	var provider schema.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Store.AddProvider(&provider); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create provider")
		return
	}
	ctx_utils.Success(c, provider)
}

// GetProvider
//
//	@Summary		获取 API 提供商
//	@Description	获取 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			provider_id	path		uint64									true	"API 提供商 ID"
//	@Success		200			{object}	entity.CommonResponse[schema.Provider]	"API 提供商"
//	@Router			/manage/provider/{provider_id} [get]
func (h *Handler) GetProvider(c *gin.Context) {
	var uri struct {
		ProviderId uint64 `uri:"provider_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.ProviderId == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	provider, err := h.Store.GetProvider(uri.ProviderId)
	if err != nil {
		ctx_utils.CustomError(c, 404, "provider not found")
		return
	}
	ctx_utils.Success(c, provider)
}

// GetProviders
//
//	@Summary		批量获取 API 提供商
//	@Description	批量获取 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			req	query		entity.ParamPagingSort													true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.Provider]]	"API 提供商列表"
//	@Router			/manage/provider/list [get]
func (h *Handler) GetProviders(c *gin.Context) {
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	req.SortParam.WithDefault("created_at ASC", "id")
	providers, total, err := gorm_utils.GetByPageTotal[schema.Provider](
		h.Db.Preload("Models"),
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get providers")
		return
	}
	ctx_utils.Success(
		c, &entity.PaginatedTotalResponse[schema.Provider]{
			List:  providers,
			Total: total,
		},
	)
}

// GetAllProviders
//
//	@Summary		获取所有 API 提供商
//	@Description	获取所有 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[[]schema.Provider]	"API 提供商列表"
//	@Router			/manage/provider/all [get]
func (h *Handler) GetAllProviders(c *gin.Context) {
	providers, err := h.Store.GetProviders()
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get providers")
		return
	}
	ctx_utils.Success(c, providers)
}

// UpdateProvider
//
//	@Summary		更新 API 提供商
//	@Description	更新 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			id			path		uint64						true	"API 提供商 ID"
//	@Param			provider	body		schema.Provider				true	"API 提供商参数"
//	@Success		200			{object}	entity.CommonResponse[bool]	"更新成功与否"
//	@Router			/manage/provider/{id}/update [post]
func (h *Handler) UpdateProvider(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var provider schema.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	provider.ID = uri.ID
	if err := h.Store.UpdateProvider(&provider); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to update provider")
		return
	}
	ctx_utils.Success(c, true)
}

// DeleteProvider
//
//	@Summary		删除 API 提供商
//	@Description	删除 API 提供商
//	@Tags			Provider
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"API 提供商 ID"
//	@Success		200	{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/provider/{id}/delete [post]
func (h *Handler) DeleteProvider(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Store.DeleteProvider(uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete provider")
		return
	}
	ctx_utils.Success(c, true)
}

///////////////////////

// CreateAPIKey
//
//	@Summary		创建 APIKey
//	@Description	创建 APIKey 并绑定 到 API 提供商
//	@Tags			APIKey
//	@Accept			json
//	@Produce		json
//	@Param			apikey	body		schema.APIKey							true	"API Key"
//	@Success		200		{object}	entity.CommonResponse[schema.APIKey]	"成功创建的 API Key"
//	@Router			/manage/key/create [post]
func (h *Handler) CreateAPIKey(c *gin.Context) {
	var apiKey schema.APIKey
	if err := c.ShouldBindJSON(&apiKey); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	resKey, err := h.Store.AddAPIKey(apiKey.ProviderID, apiKey.Key)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create key")
		return
	}
	ctx_utils.Success(c, &resKey)
}

// DeleteAPIKey
//
//	@Summary		删除 APIKey
//	@Description	删除 APIKey
//	@Tags			APIKey
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64						true	"API Key ID"
//	@Success		200	{object}	entity.CommonResponse[bool]	"删除成功与否"
//	@Router			/manage/key/{id}/delete [post]
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Store.DeleteAPIKey(uri.ID); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete key")
		return
	}
	ctx_utils.Success(c, true)
}

// GetAPIKeyByProvider
//
//	@Summary		列出APIKey
//	@Description	列出供应商的 APIKey
//	@Tags			APIKey
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64																true	"API 提供商 ID"
//	@Param			req	query		entity.ParamPagingSort												true	"分页参数"
//	@Success		200	{object}	entity.CommonResponse[entity.PaginatedTotalResponse[schema.APIKey]]	"成功创建的 API Key"
//	@Router			/manage/key/list/provider/{id} [get]
func (h *Handler) GetAPIKeyByProvider(c *gin.Context) {
	var uri entity.PathParamId
	if err := c.BindUri(&uri); err != nil || uri.ID == 0 {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req entity.ParamPagingSort
	if err := c.ShouldBindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	apiKeys, total, err := gorm_utils.GetByPageTotal[schema.APIKey](
		h.Db.Where("provider_id = ?", uri.ID),
		req.PagingParam,
		req.SortParam,
	)
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to load keys")
		return
	}
	ctx_utils.Success(
		c, entity.PaginatedTotalResponse[schema.APIKey]{
			List:  apiKeys,
			Total: total,
		},
	)
}
