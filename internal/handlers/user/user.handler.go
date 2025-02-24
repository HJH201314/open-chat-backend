package user

import (
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/models"
	"github.com/fcraft/open-chat/internal/shared/constant"
	"github.com/fcraft/open-chat/internal/shared/entity"
	"github.com/fcraft/open-chat/internal/shared/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewUserHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}

// Ping 检测客户端登录态
//
//	@Summary		检测客户端登录态
//	@Description	检测客户端登录态
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[models.User]	"user is online"
//	@Failure		404	{object}	entity.CommonResponse[any]			"user not found"
//	@Router			/user/ping [post]
func (h *Handler) Ping(c *gin.Context) {
	if userId := util.GetUserId(c); userId > 0 {
		if user, err := h.Store.GetUser(userId); err == nil {
			util.NormalResponse(c, user)
		} else {
			util.CustomErrorResponse(c, 404, "user not found")
		}
	} else {
		util.NormalResponse(c, "boom")
	}
}

func signJwtTokenIntoHeader(c *gin.Context, user *models.User) {
	// 1. 创建 Claims
	claims := entity.UserClaims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                     // 生效时间
			Issuer:    "open-chat",                                        // 签发者
			Subject:   "user-auth",                                        // 主题
		},
	}

	// 2. 签发 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("AUTH_SECRET")))
	if err != nil {
		util.HttpErrorResponse(c, constant.ErrInternal)
		return
	}

	// 3. 将 token 写入 header
	c.Writer.Header().Set("OC-Auth-Token", tokenString)
}

// Login
//
//	@Summary		用户登录
//	@Description	用户登录
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	body		user.Login.loginRequest				true	"登录请求"
//	@Success		200	{object}	entity.CommonResponse[models.User]	"login successfully"
//	@Router			/user/login [post]
func (h *Handler) Login(c *gin.Context) {
	type loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req loginRequest
	if err := c.BindJSON(&req); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	var userRes models.User
	if err := h.Store.Db.Where(
		"username = ? AND password = ?",
		req.Username,
		req.Password,
	).First(&userRes).Error; err != nil {
		util.CustomErrorResponse(c, 401, "username or password is incorrect")
		return
	}

	signJwtTokenIntoHeader(c, &userRes)
	util.NormalResponse(c, userRes)
}

// Register
//
//	@Summary		用户注册
//	@Description	用户注册
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	body		user.Register.registerRequest	true	"注册请求"
//	@Success		200	{object}	entity.CommonResponse[bool]		"register successfully"
//	@Router			/user/register [post]
func (h *Handler) Register(c *gin.Context) {
	type registerRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req registerRequest
	if err := c.BindJSON(&req); err != nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	var userRes models.User
	if err := h.Store.Db.Where("username = ?", req.Username).First(&userRes).Error; err == nil {
		util.NormalResponse(c, false)
		return
	}
	user := models.User{
		Username: req.Username,
		Password: req.Password,
	}
	if err := h.Store.CreateUser(&user); err != nil {
		util.HttpErrorResponse(c, constant.ErrInternal)
		return
	}

	util.NormalResponse(c, true)
}
