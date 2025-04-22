package user

import (
	"context"
	"encoding/json"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"io"
	"time"
)

// GetAuthUrl 获取 OAuth 认证 URL
//
//	@Summary		前往 OAuth 认证
//	@Description	前往 OAuth 认证
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string							true	"OAuth 名称"
//	@Success		200		{object}	entity.CommonResponse[string]	"OAuth 认证地址"
//	@Router			/auth/{name}/url [get]
func (h *Handler) GetAuthUrl(c *gin.Context) {
	var uri entity.PathParamName
	if err := c.BindUri(&uri); err != nil || uri.Name == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 获取 OAuth 认证信息
	config := services.GetOAuthService().GetConfig(uri.Name)
	if config == nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}
	// 使用随机字符串作为 state 并写入 redis
	randStr := random.RandString(5)
	if err := h.Redis.Client.Set(c, "oauth:state:"+randStr, uri.Name, 1*time.Hour).Err(); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	url := config.AuthCodeURL(randStr, oauth2.AccessTypeOffline)
	ctx_utils.Success(c, url)
}

type LoginByOAuthReq struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// LoginByOAuth OAuth 回调登录
//
//	@Summary		OAuth 回调登录
//	@Description	OAuth 回调登录
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string								true	"OAuth 名称"
//	@Param			req		body		LoginByOAuthReq						true	"OAuth 回调登录信息"
//	@Success		200		{object}	entity.CommonResponse[schema.User]	"用户信息"
//	@Router			/auth/{name}/do [post]
func (h *Handler) LoginByOAuth(c *gin.Context) {
	var uri entity.PathParamName
	if err := c.BindUri(&uri); err != nil || uri.Name == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req LoginByOAuthReq
	if err := c.BindJSON(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 1. 基本信息
	provider := services.GetOAuthService().GetProvider(uri.Name)
	config := services.GetOAuthService().GetConfig(uri.Name)
	if provider == nil || config == nil {
		ctx_utils.HttpError(c, constants.ErrNotFound)
		return
	}
	token, err := config.Exchange(c, req.Code)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	// 2. 获取用户信息
	var uniqueID string
	switch uri.Name {
	case "github":
		client := config.Client(context.Background(), token)
		get, err := client.Get("https://api.github.com/user")
		if err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
			return
		}
		// 解析用户信息
		defer get.Body.Close()
		var user struct {
			Login string `json:"login"`
			ID    uint64 `json:"id"`
		}
		body, err := io.ReadAll(get.Body)
		if err := json.Unmarshal(body, &user); err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
		}
		uniqueID = "github_" + convertor.ToString(user.ID)
	}
	if uniqueID == "" {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 3. 查询用户是否存在
	var oauthUser schema.OAuthUser
	if err := h.Db.Where(
		"oauth_provider_id = ? AND oauth_user_name = ?",
		provider.ID,
		uniqueID,
	).Find(&oauthUser).Error; err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal) // 数据库其它问题
		return
	}

	// 4. 没有数据，执行注册
	if oauthUser.ID == 0 {
		// 没有数据，则创建用户
		user := schema.User{
			Nickname: uniqueID,
			Type:     schema.UserTypeThirdParty,
		}
		// 创建用户
		if err := h.Store.CreateUser(&user); err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
			return
		}
		// 后续注册步骤
		if err := h.doUserRegister(oauthUser.UserID); err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
			return
		}
		// 创建用户与 OAuth 的关联
		oauthUser = schema.OAuthUser{
			OAuthProviderID: provider.ID,
			OAuthUserName:   uniqueID,
			UserID:          user.ID,
		}
		if err := h.Db.Create(&oauthUser); err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
			return
		}
	}

	// 5. 执行登录
	if user, err := h.doUserLogin(c, oauthUser.UserID); err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	} else {
		ctx_utils.Success(c, user)
	}
}
