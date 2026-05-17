package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/utils"
)

type AuthService struct {
	userDAO    *dao.UserDAO
	jwtManager *utils.JWTManager
	wechat     *WeChatConfig
}

type WeChatConfig struct {
	AppID  string
	Secret string
}

type WeChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func NewAuthService(userDAO *dao.UserDAO, jwtManager *utils.JWTManager, wechat *WeChatConfig) *AuthService {
	return &AuthService{
		userDAO:    userDAO,
		jwtManager: jwtManager,
		wechat:     wechat,
	}
}

func (s *AuthService) Login(code string) (string, time.Time, error) {
	weChatResp, err := s.code2Session(code)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get WeChat session: %w", err)
	}

	if weChatResp.ErrCode != 0 {
		return "", time.Time{}, fmt.Errorf("WeChat error: %s", weChatResp.ErrMsg)
	}

	user, err := s.userDAO.GetByOpenID(weChatResp.OpenID)
	if err != nil {
		log.Printf("[AUTH] Unregistered OpenID: %s", weChatResp.OpenID)
		return "", time.Time{}, fmt.Errorf("user not registered")
	}

	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, expiresAt, nil
}

func (s *AuthService) code2Session(code string) (*WeChatSessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.wechat.AppID, s.wechat.Secret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call WeChat API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result WeChatSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
