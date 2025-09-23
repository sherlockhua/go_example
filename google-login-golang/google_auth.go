package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
//	"os"
	//"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/people/v1"
)

// GoogleAuth 管理Google OAuth2认证流程
type GoogleAuth struct {
	config       *oauth2.Config
	stateStore   StateStore
	logger       *log.Logger
	tokenCache   TokenCache
	peopleService *people.Service
}

// StateStore 用于存储和验证状态参数，防止CSRF攻击
type StateStore interface {
	SaveState(state string, expiration time.Duration) error
	VerifyState(state string) (bool, error)
}

// TokenCache 用于缓存用户令牌
type TokenCache interface {
	SaveToken(userID string, token *oauth2.Token) error
	GetToken(userID string) (*oauth2.Token, error)
}

// User 表示从Google获取的用户信息
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture   string `json:"picture"`
}

// 内存实现的StateStore，生产环境可替换为Redis等
type MemoryStateStore struct {
	states map[string]time.Time
	mu     sync.RWMutex
}

// 内存实现的TokenCache，生产环境可替换为数据库
type MemoryTokenCache struct {
	tokens map[string]*oauth2.Token
	mu     sync.RWMutex
}

// NewGoogleAuth 创建一个新的GoogleAuth实例
func NewGoogleAuth(clientID, clientSecret, redirectURL string, scopes []string, logger *log.Logger) (*GoogleAuth, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	// 创建people服务用于获取用户信息
	ctx := context.Background()
	client := config.Client(ctx, nil)
	peopleService, err := people.New(client)
	if err != nil {
		return nil, fmt.Errorf("创建people服务失败: %v", err)
	}

	return &GoogleAuth{
		config:       config,
		stateStore:   NewMemoryStateStore(),
		logger:       logger,
		tokenCache:   NewMemoryTokenCache(),
		peopleService: peopleService,
	}, nil
}

// NewMemoryStateStore 创建内存状态存储
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		states: make(map[string]time.Time),
	}
}

// SaveState 保存状态并设置过期时间
func (s *MemoryStateStore) SaveState(state string, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(expiration)
	return nil
}

// VerifyState 验证状态是否有效
func (s *MemoryStateStore) VerifyState(state string) (bool, error) {
	s.mu.RLock()
	expiration, exists := s.states[state]
	s.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// 检查是否过期
	if time.Now().After(expiration) {
		s.mu.Lock()
		delete(s.states, state)
		s.mu.Unlock()
		return false, nil
	}

	// 一次性使用，验证后删除
	s.mu.Lock()
	delete(s.states, state)
	s.mu.Unlock()

	return true, nil
}

// NewMemoryTokenCache 创建内存令牌缓存
func NewMemoryTokenCache() *MemoryTokenCache {
	return &MemoryTokenCache{
		tokens: make(map[string]*oauth2.Token),
	}
}

// SaveToken 保存用户令牌
func (c *MemoryTokenCache) SaveToken(userID string, token *oauth2.Token) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[userID] = token
	return nil
}

// GetToken 获取用户令牌
func (c *MemoryTokenCache) GetToken(userID string) (*oauth2.Token, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	token, exists := c.tokens[userID]
	if !exists {
		return nil, errors.New("令牌不存在")
	}
	return token, nil
}

// GenerateAuthURL 生成Google登录授权URL
func (g *GoogleAuth) GenerateAuthURL() (string, error) {
	// 生成随机状态参数，防止CSRF攻击
	state, err := generateRandomState()
	if err != nil {
		g.logger.Printf("生成状态参数失败: %v", err)
		return "", err
	}

	// 保存状态，有效期10分钟
	if err := g.stateStore.SaveState(state, 10*time.Minute); err != nil {
		g.logger.Printf("保存状态失败: %v", err)
		return "", err
	}

	// 生成授权URL
	url := g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, nil
}

// HandleCallback 处理Google登录回调
func (g *GoogleAuth) HandleCallback(code, state string) (*User, error) {
	// 验证状态参数
	valid, err := g.stateStore.VerifyState(state)
	if err != nil {
		g.logger.Printf("验证状态失败: %v", err)
		return nil, fmt.Errorf("验证状态失败")
	}
	if !valid {
		g.logger.Println("无效的状态参数")
		return nil, fmt.Errorf("无效的状态参数")
	}

	// 交换令牌
	token, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		g.logger.Printf("交换令牌失败: %v", err)
		return nil, fmt.Errorf("获取访问令牌失败")
	}

	// 使用令牌获取用户信息
	user, err := g.getUserInfo(token)
	if err != nil {
		g.logger.Printf("获取用户信息失败: %v", err)
		return nil, fmt.Errorf("获取用户信息失败")
	}

	// 缓存令牌
	if err := g.tokenCache.SaveToken(user.ID, token); err != nil {
		g.logger.Printf("缓存令牌失败: %v", err)
		// 这里不返回错误，因为令牌缓存失败不影响登录流程
	}

	return user, nil
}

// RefreshToken 刷新用户令牌
func (g *GoogleAuth) RefreshToken(userID string) (*oauth2.Token, error) {
	// 获取缓存的令牌
	token, err := g.tokenCache.GetToken(userID)
	if err != nil {
		g.logger.Printf("获取缓存令牌失败: %v", err)
		return nil, err
	}

	// 检查令牌是否需要刷新
	/*if !token.Expired() {
		return token, nil
	}*/

	// 刷新令牌
	ctx := context.Background()
	source := g.config.TokenSource(ctx, token)
	newToken, err := source.Token()
	if err != nil {
		g.logger.Printf("刷新令牌失败: %v", err)
		return nil, err
	}

	// 更新缓存
	if err := g.tokenCache.SaveToken(userID, newToken); err != nil {
		g.logger.Printf("更新令牌缓存失败: %v", err)
	}

	return newToken, nil
}

// 获取用户信息
func (g *GoogleAuth) getUserInfo(token *oauth2.Token) (*User, error) {
	// 方法1: 使用people API获取详细信息
	person, err := g.peopleService.People.Get("people/me").
		PersonFields("names,emailAddresses,photos").
		Do()
	if err != nil {
		// 尝试备用方法获取用户信息
		return g.getUserInfoFallback(token)
	}

	user := &User{}

	// 提取姓名信息
	if len(person.Names) > 0 {
		user.Name = person.Names[0].DisplayName
		user.GivenName = person.Names[0].GivenName
		user.FamilyName = person.Names[0].FamilyName
	}

	// 提取邮箱信息
	if len(person.EmailAddresses) > 0 {
		user.Email = person.EmailAddresses[0].Value
		user.ID = person.EmailAddresses[0].Metadata.Source.Id
	}

	// 提取头像信息
	if len(person.Photos) > 0 && person.Photos[0].Url != "" {
		user.Picture = person.Photos[0].Url
	}

	return user, nil
}

// 备用方法获取用户信息
func (g *GoogleAuth) getUserInfoFallback(token *oauth2.Token) (*User, error) {
	client := g.config.Client(context.Background(), token)
	
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %v", err)
	}

	return &user, nil
}

// 生成随机状态字符串
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
    
