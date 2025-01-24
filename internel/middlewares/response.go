package middlewares

import (
	"github.com/fcraft/open-chat/internel/shared/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ResponseMiddleware 响应中间件
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先处理请求
		c.Next()

		// 检查是否有错误发生
		if len(c.Errors) > 0 {
			// 如果有错误，返回错误信息
			errorResponse := entity.ErrorResponse(c.Errors.Last())
			var httpCode int
			if errorResponse.Code >= 100 && errorResponse.Code < 600 {
				httpCode = errorResponse.Code
			} else {
				httpCode = http.StatusInternalServerError
			}
			c.JSON(httpCode, errorResponse)
			return
		}

		// 获取原始的响应数据
		responseData, exists := c.Get("data")
		if !exists {
			// 如果没有数据，不作处理
			return
		}

		// 封装响应数据
		okResponse := entity.OK.WithData(responseData)
		c.JSON(okResponse.Code, okResponse)
	}
}
