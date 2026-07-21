package hemis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ttysi-fit/backend/config"
)

// OAuthClient — HEMIS OAuth (Authorization Code) klienti.
// Ikki provayder: "student" (student.ttyesi.uz) va "employee" (hemis.ttyesi.uz).
type OAuthClient struct {
	cfg  config.HEMISOAuthConfig
	http *http.Client
}

// NewOAuthClient — config asosida OAuth klientini yaratadi.
func NewOAuthClient(cfg config.HEMISOAuthConfig) *OAuthClient {
	return &OAuthClient{
		cfg:  cfg,
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// HemisProfile — resource (/oauth/api/user) javobidagi kerakli maydonlar.
// Javob tekis yoki {"data": {...}} ko'rinishida bo'lishi mumkin — ikkalasi ham qo'llab-quvvatlanadi.
type HemisProfile struct {
	ID    int64  `json:"id"`
	UUID  string `json:"uuid"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// Provider — nom bo'yicha provayder configini qaytaradi ("student"|"employee").
func (c *OAuthClient) Provider(name string) (config.HEMISOAuthProvider, error) {
	switch name {
	case "student":
		return c.cfg.Student, nil
	case "employee":
		return c.cfg.Employee, nil
	default:
		return config.HEMISOAuthProvider{}, fmt.Errorf("hemis oauth: noma'lum provayder %q", name)
	}
}

// AuthorizationURL — authorize URL va CSRF uchun state qaytaradi.
func (c *OAuthClient) AuthorizationURL(providerName string) (string, string, error) {
	p, err := c.Provider(providerName)
	if err != nil {
		return "", "", err
	}
	if !p.Configured() {
		return "", "", fmt.Errorf("hemis oauth: %q provayderi sozlanmagan (client_id/url)", providerName)
	}

	state, err := randomState()
	if err != nil {
		return "", "", err
	}

	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("response_type", "code")
	q.Set("state", state)
	if len(c.cfg.Scopes) > 0 {
		q.Set("scope", strings.Join(c.cfg.Scopes, " "))
	}

	return p.AuthorizeURL + "?" + q.Encode(), state, nil
}

// FetchUser — code'ni access_token'ga almashtirib, profilni qaytaradi.
func (c *OAuthClient) FetchUser(ctx context.Context, providerName, code string) (*HemisProfile, error) {
	p, err := c.Provider(providerName)
	if err != nil {
		return nil, err
	}
	if !p.Configured() {
		return nil, fmt.Errorf("hemis oauth: %q provayderi sozlanmagan", providerName)
	}

	token, err := c.exchangeCode(ctx, p, code)
	if err != nil {
		return nil, err
	}
	return c.resource(ctx, p, token)
}

// exchangeCode — authorization_code -> access_token.
func (c *OAuthClient) exchangeCode(ctx context.Context, p config.HEMISOAuthProvider, code string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("redirect_uri", p.RedirectURI)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("hemis oauth: token so'rovi: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("hemis oauth: token status %d: %s", res.StatusCode, string(body))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("hemis oauth: token javobi: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("hemis oauth: access_token bo'sh")
	}
	return tok.AccessToken, nil
}

// resource — access_token bilan profilni oladi.
func (c *OAuthClient) resource(ctx context.Context, p config.HEMISOAuthProvider, accessToken string) (*HemisProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.ResourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hemis oauth: resource so'rovi: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hemis oauth: resource status %d: %s", res.StatusCode, string(body))
	}

	// Javob tekis yoki {"data": {...}} bo'lishi mumkin.
	var env struct {
		Data *HemisProfile `json:"data"`
		HemisProfile
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("hemis oauth: resource javobi: %w", err)
	}
	if env.Data != nil && (env.Data.ID != 0 || env.Data.Login != "") {
		return env.Data, nil
	}
	return &env.HemisProfile, nil
}

func randomState() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("hemis oauth: state: %w", err)
	}
	return hex.EncodeToString(b), nil
}
