package ctx_utils

import (
	"errors"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/gin-gonic/gin"
)

// Success 调用此函数，统一成功返回
func Success[T any](c *gin.Context, data T) {
	c.JSON(
		200, entity.CommonResponse[T]{
			Code: 200,
			Msg:  "success",
			Data: data,
		},
	)
}

// SuccessBizError 调用此函数，以 200 返回错误信息
func SuccessBizError(c *gin.Context, err error) {
	var bizError constants.BizError
	if errors.As(err, &bizError) {
		c.AbortWithStatusJSON(200, entity.ERR.WithCode(bizError.BizCode).WithMsg(bizError.Msg))
	} else {
		HttpError(c, err)
	}
}

// BizError 调用此函数，中断并返回有代码标记的错误
func BizError(c *gin.Context, err error) {
	var bizError constants.BizError
	if errors.As(err, &bizError) {
		httpCode := 400
		if bizError.HttpCode != 0 {
			httpCode = bizError.HttpCode
		}
		c.AbortWithStatusJSON(httpCode, entity.ERR.WithCode(bizError.BizCode).WithMsg(bizError.Msg))
	} else {
		HttpError(c, err)
	}
}

// CustomError 调用此函数，中断并返回格式化错误
func CustomError(c *gin.Context, code int, msg string) {
	var httpCode int
	if code >= 400 && code < 600 {
		httpCode = code
	} else {
		httpCode = 500
	}
	c.AbortWithStatusJSON(httpCode, entity.ERR.WithCode(code).WithMsg(msg))
}

// HttpError 调用此函数，中断并返回错误，error 必须为 constants.ErrStatusMap 中的 key
func HttpError(c *gin.Context, err error) {
	responseEntity := entity.ERR.WithError(err)
	c.AbortWithStatusJSON(responseEntity.Code, responseEntity)
}
