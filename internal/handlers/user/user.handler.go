package user

import (
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/models"
	"github.com/fcraft/open-chat/internal/shared/constant"
	"github.com/fcraft/open-chat/internal/shared/entity"
	"github.com/fcraft/open-chat/internal/shared/util"
	"github.com/fcraft/open-chat/internal/utils/auth_utils"
	"github.com/gin-gonic/gin"
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

// Refresh 使用 auth_token 和 refresh_token 刷新登录态
//
//	@Summary		刷新登录态
//	@Description	刷新登录态
//	@Tags			User
//	@Param			X-Refresh-Token	header		string	true	"刷新用 Token"
//	@Success		200				{string}	string	"nothing"
//	@Router			/user/refresh [get]
func (h *Handler) Refresh(c *gin.Context) {
	// 1. 验证 auth_token 是否存在、是否真的坏了
	authToken := auth_utils.ValidateAuthToken(c)
	if authToken == nil || authToken.Valid {
		return
	}
	authClaims, ok := authToken.Claims.(*entity.UserClaims)
	if !ok {
		return
	}

	// 2. 验证 auth_token 和 refresh_token 是否匹配
	refreshToken := auth_utils.ValidateRefreshToken(c, authClaims)
	if refreshToken == nil {
		return
	}
	refreshClaims, ok := authToken.Claims.(*entity.UserClaims)
	if !ok {
		return
	}

	// 3. 确认用户存在并重新签发
	if user, err := h.Store.GetUser(refreshClaims.ID); err == nil {
		signJwtTokenIntoHeader(c, user)
	}
}

func signJwtTokenIntoHeader(c *gin.Context, user *models.User) {
	authToken, err := auth_utils.SignAuthTokenForUser(user.ID)
	if err != nil {
		util.HttpErrorResponse(c, constant.ErrInternal)
		return
	}
	refreshToken, err := auth_utils.SignRefreshTokenForUser(user.ID)
	if err != nil {
		util.HttpErrorResponse(c, constant.ErrInternal)
		return
	}
	// 将 token 写入 header
	c.Writer.Header().Set("OC-Auth-Token", authToken)
	c.Writer.Header().Set("OC-Refresh-Token", refreshToken)
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
