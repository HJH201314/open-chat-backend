package user

import (
	"github.com/fcraft/open-chat/internel/shared/util"
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	util.SuccessResponse(c, "pong")
}
