package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type key struct{}
type IdentityResolver interface {
	ResolvePrincipal(context.Context, string, []string, map[string][]string) (entity.Principal, error)
}

type Auth struct {
	verifier *oidc.IDTokenVerifier
	oauth    oauth2.Config
	cfg      config.Config
	service  IdentityResolver
	cipher   *secrets.Cipher
}

func New(ctx context.Context, cfg config.Config, service IdentityResolver, cipher *secrets.Cipher) (*Auth, error) {
	a := &Auth{cfg: cfg, service: service, cipher: cipher}
	if cfg.DevToken != "" {
		return a, nil
	}
	client := &http.Client{Timeout: 10 * time.Second}
	ctx = oidc.ClientContext(ctx, client)
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	a.verifier = provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	a.oauth = oauth2.Config{ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, Endpoint: provider.Endpoint(), RedirectURL: cfg.RedirectURL, Scopes: []string{oidc.ScopeOpenID, "profile", "email"}}
	return a, nil
}

func Principal(ctx context.Context) entity.Principal {
	p, _ := ctx.Value(key{}).(entity.Principal)
	return p
}

func (a *Auth) Verify(ctx context.Context, token string) (entity.Principal, error) {
	if a.cfg.DevToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(a.cfg.DevToken)) == 1 {
		return entity.Principal{Subject: "development", GlobalAdmin: true, Claims: []entity.Claim{}}, nil
	}
	if a.verifier == nil {
		return entity.Principal{}, fmt.Errorf("invalid token")
	}
	verified, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return entity.Principal{}, fmt.Errorf("invalid token")
	}
	var claims struct {
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
		NamespaceRoles map[string][]string `json:"namespace_roles"`
	}
	if err = verified.Claims(&claims); err != nil {
		return entity.Principal{}, err
	}
	return a.service.ResolvePrincipal(ctx, verified.Subject, claims.RealmAccess.Roles, claims.NamespaceRoles)
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			if cookie, err := r.Cookie("access_token"); err == nil {
				token = cookie.Value
			}
		}
		p, err := a.Verify(r.Context(), token)
		if err != nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), key{}, p)))
	})
}

type loginState struct {
	State, Verifier, Nonce string
	CreatedAt              time.Time
}

func random() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func (a *Auth) Start(w http.ResponseWriter, r *http.Request) {
	if a.verifier == nil {
		http.Error(w, "OIDC is not configured", http.StatusServiceUnavailable)
		return
	}
	s := loginState{random(), random(), random(), time.Now().UTC()}
	b, err := json.Marshal(s)
	if err != nil {
		http.Error(w, "cannot start login", 500)
		return
	}
	sealed, err := a.cipher.Seal("oidc-state", string(b))
	if err != nil {
		http.Error(w, "cannot start login", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oidc_state", Value: sealed, Path: "/api/v1/user", HttpOnly: true, Secure: strings.HasPrefix(a.cfg.RedirectURL, "https:"), SameSite: http.SameSiteLaxMode, MaxAge: 300})
	hash := sha256.Sum256([]byte(s.Verifier))
	url := a.oauth.AuthCodeURL(s.State, oidc.Nonce(s.Nonce), oauth2.SetAuthURLParam("code_challenge", base64.RawURLEncoding.EncodeToString(hash[:])), oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *Auth) Callback(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "oidc_state", Value: "", Path: "/api/v1/user", MaxAge: -1, HttpOnly: true, Secure: strings.HasPrefix(a.cfg.RedirectURL, "https:"), SameSite: http.SameSiteLaxMode})
	cookie, err := r.Cookie("oidc_state")
	if err != nil {
		http.Error(w, "invalid login state", 400)
		return
	}
	plain, err := a.cipher.Open("oidc-state", cookie.Value)
	if err != nil {
		http.Error(w, "invalid login state", 400)
		return
	}
	var state loginState
	if json.Unmarshal([]byte(plain), &state) != nil || time.Since(state.CreatedAt) > 5*time.Minute || subtle.ConstantTimeCompare([]byte(state.State), []byte(r.URL.Query().Get("state"))) != 1 {
		http.Error(w, "invalid login state", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	token, err := a.oauth.Exchange(ctx, r.URL.Query().Get("code"), oauth2.SetAuthURLParam("code_verifier", state.Verifier))
	if err != nil {
		http.Error(w, "login failed", 401)
		return
	}
	idToken, _ := token.Extra("id_token").(string)
	verified, err := a.verifier.Verify(ctx, idToken)
	if err != nil || verified.Nonce != state.Nonce {
		http.Error(w, "invalid identity token", 401)
		return
	}
	if _, err = a.Verify(ctx, token.AccessToken); err != nil {
		http.Error(w, "invalid access token audience", 401)
		return
	}
	a.cookie(w, token)
	http.Redirect(w, r, a.cfg.PublicURL, http.StatusFound)
}

func (a *Auth) Password(w http.ResponseWriter, r *http.Request) {
	if a.verifier == nil {
		http.Error(w, "OIDC is not configured", 503)
		return
	}
	var input struct{ Email, Password string }
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.Email == "" || input.Password == "" {
		http.Error(w, "email and password are required", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	token, err := a.oauth.PasswordCredentialsToken(ctx, input.Email, input.Password)
	if err != nil {
		http.Error(w, "login failed", 401)
		return
	}
	if _, err = a.Verify(ctx, token.AccessToken); err != nil {
		http.Error(w, "invalid access token audience", 401)
		return
	}
	a.cookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{"accessToken": token.AccessToken, "expiresIn": int(time.Until(token.Expiry).Seconds())})
}

func (a *Auth) cookie(w http.ResponseWriter, t *oauth2.Token) {
	maxAge := int(time.Until(t.Expiry).Seconds())
	if maxAge < 1 {
		maxAge = 300
	}
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: t.AccessToken, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(a.cfg.PublicURL, "https:"), SameSite: http.SameSiteLaxMode, MaxAge: maxAge})
}
