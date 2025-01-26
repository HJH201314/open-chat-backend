package util

import (
	"github.com/fcraft/open-chat/internel/shared/entity"
	"github.com/gin-gonic/gin"
)

// NormalResponse 调用此函数，通过中间件统一成功返回
func NormalResponse(c *gin.Context, data interface{}) {
	c.JSON(200, entity.OK.WithData(data))
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
