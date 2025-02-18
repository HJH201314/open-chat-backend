package manage

import (
	"github.com/fcraft/open-chat/internel/models"
	"github.com/fcraft/open-chat/internel/shared/constant"
	"github.com/fcraft/open-chat/internel/shared/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

func (h *Handler) GetProviders(c *gin.Context) {
	providers, err := h.Store.GetProviders()
	if err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to get providers")
		return
	}
	util.NormalResponse(c, providers)
}

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
	util.NormalResponse(c, provider)
}

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
	util.NormalResponse(c, model)
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
