package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/user/model"
	"github.com/travel-booking/server/internal/user/repository"
)

const (
	wechatTokenURL  = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL = "https://api.weixin.qq.com/sns/userinfo"
)

// WechatService handles WeChat OAuth integration.
type WechatService struct {
	repo       *repository.UserRepository
	smsService *SMSService
	jwtManager *middleware.JWTManager
	cfg        *config.Config
	logger     *zap.Logger
}

// NewWechatService creates a new WechatService.
func NewWechatService(
	repo *repository.UserRepository,
	smsService *SMSService,
	jwtManager *middleware.JWTManager,
	cfg *config.Config,
	logger *zap.Logger,
) *WechatService {
	return &WechatService{
		repo:       repo,
		smsService: smsService,
		jwtManager: jwtManager,
		cfg:        cfg,
		logger:     logger,
	}
}

// WechatLoginRequest is the request body for WeChat login.
type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	BindPhone string `json:"bind_phone"`
	BindCode  string `json:"bind_code"`
}

// WechatLoginResponse is the response for WeChat login.
type WechatLoginResponse struct {
	User          *UserResponse `json:"user,omitempty"`
	AccessToken   string        `json:"access_token,omitempty"`
	RefreshToken  string        `json:"refresh_token,omitempty"`
	NeedBindPhone bool          `json:"need_bindphone"`
}

// WechatUserInfo holds info from WeChat API.
type WechatUserInfo struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
}

// Login handles WeChat OAuth login.
// Step 1: Exchange code for OpenID. If OpenID exists, log in directly.
// Step 2: If OpenID is new and no bind_phone, return need_bindphone=true.
// Step 3: If bind_phone+bind_code provided, verify SMS and bind.
func (s *WechatService) Login(req WechatLoginRequest) (*WechatLoginResponse, error) {
	// Exchange code for WeChat user info
	wxUser, err := s.exchangeCode(req.Code)
	if err != nil {
		return nil, fmt.Errorf("wechat code exchange: %w", err)
	}

	// Check if WeChat OpenID is already linked
	user, err := s.repo.FindByWechatOpenID(wxUser.OpenID)
	if err == nil {
		// Already linked — log in directly
		accessToken, refreshToken, tokenErr := s.jwtManager.GenerateTokenPair(
			user.ID, "user", nil, nil,
		)
		if tokenErr != nil {
			return nil, fmt.Errorf("generate tokens: %w", tokenErr)
		}
		return &WechatLoginResponse{
			User:         toUserResponseFromModel(user),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find user by openid: %w", err)
	}

	// New WeChat user — need phone binding
	if req.BindPhone == "" || req.BindCode == "" {
		return &WechatLoginResponse{
			NeedBindPhone: true,
		}, nil
	}

	// Verify SMS code for phone binding
	if err := s.smsService.VerifyCode(context.Background(), req.BindPhone, req.BindCode); err != nil {
		return nil, err
	}

	// Check if phone is already registered
	existingUser, err := s.repo.FindByPhone(req.BindPhone)
	if err == nil {
		// Phone already registered — bind WeChat to existing account
		if bindErr := s.repo.BindWechat(existingUser.ID, wxUser.OpenID, wxUser.UnionID); bindErr != nil {
			return nil, fmt.Errorf("bind wechat: %w", bindErr)
		}
		accessToken, refreshToken, tokenErr := s.jwtManager.GenerateTokenPair(
			existingUser.ID, "user", nil, nil,
		)
		if tokenErr != nil {
			return nil, fmt.Errorf("generate tokens: %w", tokenErr)
		}
		return &WechatLoginResponse{
			User:         toUserResponseFromModel(existingUser),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	// Create new user with phone binding
	nickname := wxUser.Nickname
	if nickname == "" {
		nickname = fmt.Sprintf("%s%s", defaultNicknamePrefix, req.BindPhone[7:])
	}
	avatarURL := wxUser.HeadImgURL
	if avatarURL == "" {
		avatarURL = defaultAvatarURL
	}

	newUser := &model.UserAccount{
		Phone:          req.BindPhone,
		Nickname:       nickname,
		AvatarURL:      avatarURL,
		RealNameStatus: model.RNStatusUnverified,
		MemberLevel:    1,
		Status:         model.UserStatusActive,
		WechatOpenID:   wxUser.OpenID,
		WechatUnionID:  wxUser.UnionID,
	}
	if err := s.repo.Create(newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(
		newUser.ID, "user", nil, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &WechatLoginResponse{
		User:         toUserResponseFromModel(newUser),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// exchangeCode exchanges a WeChat OAuth code for user info.
func (s *WechatService) exchangeCode(code string) (*WechatUserInfo, error) {
	appID := s.cfg.Payment.Wechat.AppID
	// For WeChat OAuth, we use a separate app secret. For now, use the payment config's app ID.
	// In production, this should have its own config for WeChat OAuth app secret.

	url := fmt.Sprintf("%s?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		wechatTokenURL, appID, "", code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("wechat token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wechat response: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		OpenID       string `json:"openid"`
		UnionID      string `json:"unionid"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse wechat token response: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error %d: %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// Get user info
	userInfoURL := fmt.Sprintf("%s?access_token=%s&openid=%s&lang=zh_CN",
		wechatUserInfoURL, tokenResp.AccessToken, tokenResp.OpenID)

	infoResp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("wechat user info request: %w", err)
	}
	defer infoResp.Body.Close()

	infoBody, err := io.ReadAll(infoResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wechat user info: %w", err)
	}

	var userInfo WechatUserInfo
	if err := json.Unmarshal(infoBody, &userInfo); err != nil {
		return nil, fmt.Errorf("parse wechat user info: %w", err)
	}
	userInfo.OpenID = tokenResp.OpenID
	userInfo.UnionID = tokenResp.UnionID

	return &userInfo, nil
}

// toUserResponseFromModel converts a UserAccount model to UserResponse.
func toUserResponseFromModel(user *model.UserAccount) *UserResponse {
	maskedPhone := user.Phone
	if len(maskedPhone) >= 7 {
		maskedPhone = maskedPhone[:3] + "****" + maskedPhone[len(maskedPhone)-4:]
	}
	return &UserResponse{
		ID:             user.ID,
		Phone:          maskedPhone,
		Nickname:       user.Nickname,
		AvatarURL:      user.AvatarURL,
		RealNameStatus: user.RealNameStatus,
		MemberLevel:    user.MemberLevel,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
	}
}
