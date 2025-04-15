package base

import (
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

type KeyHandler struct {
	handlers.BaseHandler
}

func NewKeyHandler(handler *handlers.BaseHandler) *KeyHandler {
	return &KeyHandler{
		BaseHandler: *handler,
	}
}

// GetPublicKey
//
//	@Summary		获取公钥
//	@Description	获取公钥
//	@Tags			Base
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[string]	"RSA 公钥"
//	@Router			/base/public-key [get]
func (h *KeyHandler) GetPublicKey(c *gin.Context) {
	ctx_utils.Success(c, services.GetEncryptService().PublicKeyPEM)
}
