package ctx_utils

import (
	"github.com/fcraft/open-chat/internal/entities"
	"github.com/gin-gonic/gin"
)

// Success 调用此函数，统一成功返回
func Success[T any](c *gin.Context, data T) {
	c.JSON(
		200, entities.CommonResponse[T]{
			Code: 200,
			Msg:  "success",
			Data: data,
		},
	)
}

// CustomError 调用此函数，中断并返回格式化错误
func CustomError(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(500, entities.ERR.WithCode(code).WithMsg(msg))
}

// HttpError 调用此函数，中断并返回错误，error 必须为 constants.ErrStatusMap 中的 key
func HttpError(c *gin.Context, err error) {
	responseEntity := entities.ERR.WithError(err)
	c.AbortWithStatusJSON(responseEntity.Code, responseEntity)
}
