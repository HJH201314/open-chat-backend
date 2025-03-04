package chat

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetModels
//
//	@Summary		获取所有模型
//	@Description	获取所有模型
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[[]schema.ModelCache]
//	@Router			/chat/config/schema [get]
func (h *Handler) GetModels(c *gin.Context) {
	// 从缓存中查询
	cacheModels, err := h.Redis.GetCachedModels()
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get schema")
		return
	}
	// 将 config 隐藏
	slice.ForEach(
		cacheModels, func(_ int, item schema.ModelCache) {
			item.Config = schema.ModelConfig{}
		},
	)
	ctx_utils.Success(c, cacheModels)
}
