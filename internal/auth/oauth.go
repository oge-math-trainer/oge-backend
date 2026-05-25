package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

const (
	ProviderGoogle = "google"
	ProviderYandex = "yandex"

	defaultGoogleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultGoogleTokenURL    = "https://oauth2.googleapis.com/token"
	defaultGoogleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

	defaultYandexAuthURL     = "https://oauth.yandex.ru/authorize"
	defaultYandexTokenURL    = "https://oauth.yandex.ru/token"
	defaultYandexUserInfoURL = "https://login.yandex.ru/info"

	oauthStateTTL      = 10 * time.Minute
	maxOAuthBodyBytes  = 1 << 20
	oauthStatePartSize = 2
)

type OAuthConfig struct {
	Google OAuthProviderConfig
	Yandex OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string

	AuthURL     string
	TokenURL    string
	UserInfoURL string
}

type OAuthStart struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
}

type oauthState struct {
	Provider  string `json:"provider"`
	Nonce     string `json:"nonce"`
	ExpiresAt int64  `json:"expires_at"`
}

type oauthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type oauthProfile struct {
	ID    string
	Email string
}

func (s *Service) OAuthStart(ctx context.Context, provider string) (OAuthStart, error) {
	_ = ctx

	provider, cfg, err := s.oauthProvider(provider)
	if err != nil {
		return OAuthStart{}, err
	}

	state, err := s.createOAuthState(provider)
	if err != nil {
		return OAuthStart{}, err
	}

	authURL, err := url.Parse(cfg.AuthURL)
	if err != nil {
		return OAuthStart{}, app.Internal(err)
	}
	query := authURL.Query()
	query.Set("client_id", cfg.ClientID)
	query.Set("redirect_uri", cfg.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(oauthScopes(provider), " "))
	query.Set("state", state)
	if provider == ProviderGoogle {
		query.Set("access_type", "online")
		query.Set("prompt", "select_account")
	}
	authURL.RawQuery = query.Encode()

	return OAuthStart{
		Provider: provider,
		URL:      authURL.String(),
	}, nil
}

func (s *Service) OAuthCallback(ctx context.Context, provider, code, state string) (Session, error) {
	provider, cfg, err := s.oauthProvider(provider)
	if err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(code) == "" {
		return Session{}, app.Validation("OAuth code is required")
	}
	if strings.TrimSpace(state) == "" {
		return Session{}, app.Validation("OAuth state is required")
	}
	if err := s.verifyOAuthState(provider, state); err != nil {
		return Session{}, err
	}

	token, err := s.exchangeOAuthCode(ctx, provider, cfg, code)
	if err != nil {
		return Session{}, err
	}
	profile, err := s.fetchOAuthProfile(ctx, provider, cfg, token.AccessToken)
	if err != nil {
		return Session{}, err
	}

	identity := OAuthIdentity{
		Provider:       provider,
		ProviderUserID: strings.TrimSpace(profile.ID),
		Email:          normalizeEmail(profile.Email),
	}
	if identity.ProviderUserID == "" {
		return Session{}, app.Unauthorized("OAuth provider did not return user id")
	}
	if identity.Email == "" {
		return Session{}, app.Unauthorized("OAuth provider did not return email")
	}

	user, err := s.repo.FindOrCreateOAuthUser(ctx, identity)
	if err != nil {
		return Session{}, err
	}
	signed, err := s.createToken(user.ID)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: signed, User: user}, nil
}

func (s *Service) oauthProvider(provider string) (string, OAuthProviderConfig, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderGoogle:
		cfg := s.oauth.Google.withDefaults(ProviderGoogle)
		if !cfg.isConfigured() {
			return "", OAuthProviderConfig{}, app.Validation("Google OAuth is not configured")
		}
		return ProviderGoogle, cfg, nil
	case ProviderYandex:
		cfg := s.oauth.Yandex.withDefaults(ProviderYandex)
		if !cfg.isConfigured() {
			return "", OAuthProviderConfig{}, app.Validation("Yandex OAuth is not configured")
		}
		return ProviderYandex, cfg, nil
	default:
		return "", OAuthProviderConfig{}, app.Validation("Unsupported OAuth provider")
	}
}

func (cfg OAuthProviderConfig) isConfigured() bool {
	return strings.TrimSpace(cfg.ClientID) != "" &&
		strings.TrimSpace(cfg.ClientSecret) != "" &&
		strings.TrimSpace(cfg.RedirectURL) != ""
}

func (cfg OAuthProviderConfig) withDefaults(provider string) OAuthProviderConfig {
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
	cfg.RedirectURL = strings.TrimSpace(cfg.RedirectURL)
	cfg.AuthURL = strings.TrimSpace(cfg.AuthURL)
	cfg.TokenURL = strings.TrimSpace(cfg.TokenURL)
	cfg.UserInfoURL = strings.TrimSpace(cfg.UserInfoURL)

	switch provider {
	case ProviderGoogle:
		if cfg.AuthURL == "" {
			cfg.AuthURL = defaultGoogleAuthURL
		}
		if cfg.TokenURL == "" {
			cfg.TokenURL = defaultGoogleTokenURL
		}
		if cfg.UserInfoURL == "" {
			cfg.UserInfoURL = defaultGoogleUserInfoURL
		}
	case ProviderYandex:
		if cfg.AuthURL == "" {
			cfg.AuthURL = defaultYandexAuthURL
		}
		if cfg.TokenURL == "" {
			cfg.TokenURL = defaultYandexTokenURL
		}
		if cfg.UserInfoURL == "" {
			cfg.UserInfoURL = defaultYandexUserInfoURL
		}
	}
	return cfg
}

func oauthScopes(provider string) []string {
	switch provider {
	case ProviderGoogle:
		return []string{"openid", "email", "profile"}
	case ProviderYandex:
		return []string{"login:email", "login:info"}
	default:
		return nil
	}
}

func (s *Service) createOAuthState(provider string) (string, error) {
	if strings.TrimSpace(s.secret) == "" {
		return "", app.Internal(errors.New("AUTH_SECRET is empty"))
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", app.Internal(err)
	}
	state := oauthState{
		Provider:  provider,
		Nonce:     base64.RawURLEncoding.EncodeToString(nonce),
		ExpiresAt: s.now().Add(oauthStateTTL).Unix(),
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return "", app.Internal(err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := s.signOAuthState(encodedPayload)
	return encodedPayload + "." + signature, nil
}

func (s *Service) verifyOAuthState(provider, rawState string) error {
	parts := strings.Split(rawState, ".")
	if len(parts) != oauthStatePartSize {
		return app.Unauthorized("Invalid OAuth state")
	}
	expected := s.signOAuthState(parts[0])
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return app.Unauthorized("Invalid OAuth state")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return app.Unauthorized("Invalid OAuth state")
	}
	var state oauthState
	if err := json.Unmarshal(payload, &state); err != nil {
		return app.Unauthorized("Invalid OAuth state")
	}
	if state.Provider != provider {
		return app.Unauthorized("Invalid OAuth state")
	}
	if state.ExpiresAt < s.now().Unix() {
		return app.Unauthorized("OAuth state expired")
	}
	return nil
}

func (s *Service) signOAuthState(payload string) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) exchangeOAuthCode(ctx context.Context, provider string, cfg OAuthProviderConfig, code string) (oauthTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", strings.TrimSpace(code))
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("redirect_uri", cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return oauthTokenResponse{}, app.Internal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	var token oauthTokenResponse
	if err := s.doOAuthJSON(req, &token); err != nil {
		return oauthTokenResponse{}, app.Unauthorized(fmt.Sprintf("%s OAuth code exchange failed", provider))
	}
	if token.ErrorCode != "" {
		return oauthTokenResponse{}, app.Unauthorized(token.ErrorDescription)
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return oauthTokenResponse{}, app.Unauthorized("OAuth provider did not return access token")
	}
	return token, nil
}

func (s *Service) fetchOAuthProfile(ctx context.Context, provider string, cfg OAuthProviderConfig, accessToken string) (oauthProfile, error) {
	userInfoURL := cfg.UserInfoURL
	if provider == ProviderYandex {
		u, err := url.Parse(userInfoURL)
		if err != nil {
			return oauthProfile{}, app.Internal(err)
		}
		query := u.Query()
		query.Set("format", "json")
		u.RawQuery = query.Encode()
		userInfoURL = u.String()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return oauthProfile{}, app.Internal(err)
	}
	req.Header.Set("Accept", "application/json")
	switch provider {
	case ProviderGoogle:
		req.Header.Set("Authorization", "Bearer "+accessToken)
	case ProviderYandex:
		req.Header.Set("Authorization", "OAuth "+accessToken)
	}

	switch provider {
	case ProviderGoogle:
		var out struct {
			Sub           string `json:"sub"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := s.doOAuthJSON(req, &out); err != nil {
			return oauthProfile{}, app.Unauthorized("Google user info request failed")
		}
		if !out.EmailVerified {
			return oauthProfile{}, app.Unauthorized("Google email is not verified")
		}
		return oauthProfile{ID: out.Sub, Email: out.Email}, nil
	case ProviderYandex:
		var out struct {
			ID           string   `json:"id"`
			DefaultEmail string   `json:"default_email"`
			Emails       []string `json:"emails"`
		}
		if err := s.doOAuthJSON(req, &out); err != nil {
			return oauthProfile{}, app.Unauthorized("Yandex user info request failed")
		}
		email := out.DefaultEmail
		if email == "" && len(out.Emails) > 0 {
			email = out.Emails[0]
		}
		return oauthProfile{ID: out.ID, Email: email}, nil
	default:
		return oauthProfile{}, app.Validation("Unsupported OAuth provider")
	}
}

func (s *Service) doOAuthJSON(req *http.Request, dst any) error {
	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body := io.LimitReader(resp.Body, maxOAuthBodyBytes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, body)
		return fmt.Errorf("oauth request failed: status=%d", resp.StatusCode)
	}
	if err := json.NewDecoder(body).Decode(dst); err != nil {
		return err
	}
	return nil
}
