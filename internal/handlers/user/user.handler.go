package user

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/auth_utils"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"strings"
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
//	@Success		200	{object}	entity.CommonResponse[schema.User]	"user is online"
//	@Failure		404	{object}	entity.CommonResponse[any]			"user not found"
//	@Router			/user/ping [post]
func (h *Handler) Ping(c *gin.Context) {
	if userId := ctx_utils.GetUserId(c); userId > 0 {
		if user, err := h.Store.GetUserDetailed(userId); err == nil {
			ctx_utils.Success(c, user)
		} else {
			ctx_utils.CustomError(c, 404, "user not found")
		}
	} else {
		ctx_utils.Success(c, "boom")
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
	if user, err := h.Store.GetUserDetailed(refreshClaims.ID); err == nil {
		token, _ := signJwtTokenIntoHeader(c, user)
		if err := h.Redis.CacheUserToken(user.ID, token, constants.RefreshTokenExpire); err != nil {
			return
		}
	}
}

func signJwtTokenIntoHeader(c *gin.Context, user *schema.User) (string, string) {
	authToken, err := auth_utils.SignAuthTokenForUser(user.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return "", ""
	}
	refreshToken, err := auth_utils.SignRefreshTokenForUser(user.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return "", ""
	}
	// 将 token 写入 header
	c.Writer.Header().Set("OC-Auth-Token", authToken)
	c.Writer.Header().Set("OC-Refresh-Token", refreshToken)
	return authToken, refreshToken
}

// Current
//
//	@Summary		获取当前用户信息
//	@Description	获取当前用户信息
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[schema.User]	"get current user info successfully"
//	@Router			/user/current [get]
func (h *Handler) Current(c *gin.Context) {
	if userId := ctx_utils.GetUserId(c); userId > 0 {
		if user, err := h.Store.GetUserDetailed(userId); err == nil {
			user.Password = ""
			ctx_utils.Success(c, user)
		} else {
			ctx_utils.CustomError(c, 404, "user not found")
		}
	}
}

// Login
//
//	@Summary		用户登录
//	@Description	用户登录
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	body		user.Login.loginRequest				true	"登录请求"
//	@Success		200	{object}	entity.CommonResponse[schema.User]	"login successfully"
//	@Router			/user/login [post]
func (h *Handler) Login(c *gin.Context) {
	type loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req loginRequest
	if err := c.BindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var userRes schema.User
	pwd, err := services.GetEncryptService().Decrypt(req.Password)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := h.Db.Where(
		"username = ? AND password = ?",
		req.Username,
		pwd,
	).First(&userRes).Error; err != nil {
		ctx_utils.CustomError(c, 401, "username or password is incorrect")
		return
	}

	user, err := h.doUserLogin(c, userRes.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, user)
}

// BackdoorLogin
//
//	@Summary		后门登录
//	@Description	后门登录
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	body		user.Login.loginRequest				true	"登录请求"
//	@Success		200	{object}	entity.CommonResponse[schema.User]	"login successfully"
//	@Router			/user/backdoor/login [post]
func (h *Handler) BackdoorLogin(c *gin.Context) {
	type loginRequest struct {
		Username string `json:"username" binding:"required"`
	}
	var req loginRequest
	if err := c.BindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	var userRes schema.User
	if err := h.Db.Where(
		"username = ?",
		req.Username,
	).First(&userRes).Error; err != nil {
		ctx_utils.CustomError(c, 401, "username is incorrect")
		return
	}

	// 刷新角色
	if err := h.Redis.DeleteUserRolesCache(userRes.ID); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 获取详细用户信息
	detailedUser, err := h.Store.GetUserDetailed(userRes.ID)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 签发并缓存 token
	token, _ := signJwtTokenIntoHeader(c, detailedUser)
	if err := h.Redis.CacheUserToken(userRes.ID, token, constants.RefreshTokenExpire); err != nil {
		return
	}
	ctx_utils.Success(c, detailedUser)
}

// Logout
//
//	@Summary		用户登出
//	@Description	用户登出
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Router			/user/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	if err := h.Redis.InvalidUserToken(ctx_utils.GetUserId(c), ctx_utils.GetRawAuthToken(c)); err != nil {
		return
	}
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
	if err := c.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" || strings.Contains(
		req.Username,
		"_",
	) {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var userRes schema.User
	if err := h.Db.Where("username = ?", req.Username).First(&userRes).Error; err == nil {
		ctx_utils.Success(c, false)
		return
	}
	pwd, err := services.GetEncryptService().Decrypt(req.Password)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	user := schema.User{
		Username: req.Username,
		Password: pwd,
		Type:     schema.UserTypeNormal,
	}
	if err := h.Store.CreateUser(&user); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	if err := h.doUserRegister(user.ID); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	ctx_utils.Success(c, true)
}

// userLogin 用户登录处理，包含 角色授权、token 签发；返回用户数据
func (h *Handler) doUserLogin(c *gin.Context, userId uint64) (*schema.User, error) {
	// 赠送用量
	if _, err := h.Store.CreateUserUsage(userId, 100000); err != nil {
		return nil, err
	}

	// 刷新角色
	if err := h.Redis.DeleteUserRolesCache(userId); err != nil {
		return nil, err
	}

	// 获取详细用户信息
	detailedUser, err := h.Store.GetUserDetailed(userId)
	if err != nil {
		return nil, err
	}

	// 签发并缓存 token
	token, _ := signJwtTokenIntoHeader(c, detailedUser)
	if err := h.Redis.CacheUserToken(userId, token, constants.RefreshTokenExpire); err != nil {
		return nil, err
	}

	return detailedUser, nil
}

func (h *Handler) doUserRegister(userId uint64) error {
	// 赠送用量
	if _, err := h.Store.CreateUserUsage(userId, 100000); err != nil {
		return err
	}

	// 添加默认USER角色
	var userRole schema.Role
	if err := h.Db.Where("name = ?", "USER").First(&userRole).Error; err != nil {
		// 如果USER角色不存在，创建它
		userRole = schema.Role{
			Name:        "USER",
			Description: "普通用户",
		}
		if err := h.Db.Create(&userRole).Error; err != nil {
			return err
		}
	}
	// 绑定USER角色到新用户
	if err := h.Helper.BindRolesToUser(userId, []uint64{userRole.ID}); err != nil {
		return err
	}

	return nil
}
