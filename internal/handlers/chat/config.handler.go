package chat

import (
	"encoding/json"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

// GetModelConfig
//
//	@Summary		获取模型配置
//	@Description	获取模型配置
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[[]entity.ConfigChatModel]
//	@Router			/chat/config/models [get]
func (h *Handler) GetModelConfig(c *gin.Context) {
	config, err := services.GetSystemConfigService().GetConfig(services.ConfigAvailableChatModelCollection)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	var modelConfig []entity.ConfigChatModel
	if err := json.Unmarshal(config.Value, &modelConfig); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	ctx_utils.Success(c, modelConfig)
}

// GetBotConfig
//
//	@Summary		获取 bot 角色配置
//	@Description	获取 bot 角色配置
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[[]schema.Preset]
//	@Router			/chat/config/bots [get]
func (h *Handler) GetBotConfig(c *gin.Context) {
	// 从缓存中查询
	cachedPresets, err := h.Redis.GetCachedPresetsByModule("chat")
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	// 将 session 信息隐藏
	cachedPresets = slice.Map(
		cachedPresets, func(_ int, item schema.Preset) schema.Preset {
			item.PromptSession = nil
			item.PromptSessionId = ""
			return item
		},
	)
	ctx_utils.Success(c, cachedPresets)
}
