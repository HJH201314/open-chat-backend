package services

import (
	"fmt"
	"github.com/fcraft/open-chat/internal/schema"
	"golang.org/x/oauth2"
	"os"
	"sync"
)

type OAuthService struct {
	ProviderMap map[string]*schema.OAuthProvider // provider name -> provider
	ConfigMap   map[string]*oauth2.Config        // provider name -> config
	BaseService
}

var (
	oauthServiceInstance *OAuthService
	oauthServiceOnce     sync.Once
)

func InitOAuthService(base *BaseService) *OAuthService {
	oauthServiceOnce.Do(
		func() {
			oauthServiceInstance = &OAuthService{
				ProviderMap: make(map[string]*schema.OAuthProvider),
				ConfigMap:   make(map[string]*oauth2.Config),
				BaseService: *base,
			}
		},
	)
	return oauthServiceInstance
}

func GetOAuthService() *OAuthService {
	if oauthServiceInstance == nil {
		panic("oauth service not initialized")
	}
	return oauthServiceInstance
}

func (s *OAuthService) GetConfig(name string) *oauth2.Config {
	// 1.  从内存读取
	if config, ok := s.ConfigMap[name]; ok {
		return config
	}
	// 2. 从 provider 创建
	provider := s.GetProvider(name)
	if provider == nil {
		return nil
	}
	config := &oauth2.Config{
		ClientID:     provider.ClientId,
		ClientSecret: provider.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.AuthUrl,
			TokenURL: provider.TokenUrl,
		},
		Scopes:      provider.Scopes.Data(),
		RedirectURL: fmt.Sprintf("%s/auth/redirect/%s", os.Getenv("APP_BASE_URL"), name),
	}
	// 3. 保存到内存并返回
	s.ConfigMap[name] = config
	return config
}

func (s *OAuthService) GetProvider(name string) *schema.OAuthProvider {
	// 1. 从内存中读取
	if provider, ok := s.ProviderMap[name]; ok {
		return provider
	}
	// 2. 从数据库中刷新读取
	provider, err := s.RefreshProvider(name)
	if err != nil {
		return nil
	}
	return provider
}

func (s *OAuthService) RefreshProvider(name string) (*schema.OAuthProvider, error) {
	var provider schema.OAuthProvider
	if err := s.Gorm.Where("name = ?", name).First(&provider).Error; err != nil {
		return nil, err
	}
	s.ProviderMap[name] = &provider
	delete(s.ConfigMap, name) // 删除旧的配置，下次 GetConfig 再计算
	return &provider, nil
}

func (s *OAuthService) DeleteProvider(name string) error {
	// 1. 缓存中删除删除
	delete(s.ProviderMap, name)
	// 2. 数据库中删除
	if err := s.Gorm.Delete("name = ?", name).Error; err != nil {
		return err
	}
	return nil
}
