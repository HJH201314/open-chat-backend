package manage

import (
	"github.com/fcraft/open-chat/internal/constants"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// UpdateSystemConfig 更新系统配置
//
//	@Summary		更新系统配置
//	@Description	更新系统配置
//	@Tags			SystemConfig
//	@Accept			json
//	@Produce		json
//	@Param			params	body		UpdateSystemConfigParams	true	"更新系统配置参数"
//	@Success		200		{object}	entity.CommonResponse[bool]	"更新系统配置成功
//	@Route			/system-config/update
func (h *Handler) UpdateSystemConfig(c *gin.Context) {
	var params UpdateSystemConfigParams
	if err := c.ShouldBindJSON(&params); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	if err := services.GetSystemConfigService().SetConfig(params.Name, params.Value); err != nil {
		ctx_utils.CustomError(c, 500, err.Error())
		return
	}

	ctx_utils.Success(c, true)
}

// ResetSystemConfig 重置系统配置
//
//	@Summary		重置系统配置
//	@Description	重置系统配置
//	@Tags			SystemConfig
//	@Accept			json
//	@Produce		json
//	@Param			params	body		UpdateSystemConfigParams	true	"重置系统配置参数"
//	@Success		200		{object}	entity.CommonResponse[bool]	"重置系统配置成功
//	@Route			/system-config/reset
func (h *Handler) ResetSystemConfig(c *gin.Context) {
	var params ResetSystemConfigParams
	if err := c.ShouldBindJSON(&params); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	if err := services.GetSystemConfigService().ResetConfig(params.Name); err != nil {
		ctx_utils.CustomError(c, 500, err.Error())
		return
	}

	ctx_utils.Success(c, true)
}

type UpdateSystemConfigParams struct {
	Name  string                  `json:"name"`
	Value datatypes.JSONType[any] `json:"value" swaggertype:"object,string"`
}

type ResetSystemConfigParams struct {
	Name string `json:"name"`
}
