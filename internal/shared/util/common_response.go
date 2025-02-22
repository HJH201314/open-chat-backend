package util

import (
	"github.com/fcraft/open-chat/internal/shared/entity"
	"github.com/gin-gonic/gin"
)

// NormalResponse 调用此函数，统一成功返回
func NormalResponse[T any](c *gin.Context, data T) {
	c.JSON(200, entity.CommonResponse[T]{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// CustomErrorResponse 调用此函数，中断并返回格式化错误
func CustomErrorResponse(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(500, entity.ERR.WithCode(code).WithMsg(msg))
}

// HttpErrorResponse 调用此函数，中断并返回错误，error 必须为 constant.ErrStatusMap 中的 key
func HttpErrorResponse(c *gin.Context, err error) {
	responseEntity := entity.ERR.WithError(err)
	c.AbortWithStatusJSON(responseEntity.Code, responseEntity)
}
