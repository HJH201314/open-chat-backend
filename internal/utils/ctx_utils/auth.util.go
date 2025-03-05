package ctx_utils

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func GetRawAuthToken(c *gin.Context) string {
	return strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", 1)
}

func GetRawRefreshToken(c *gin.Context) string {
	return c.GetHeader("X-Refresh-Token")
}
