package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr, DatabaseURL, RedisURL             string
	Issuer, ClientID, ClientSecret, RedirectURL string
	PublicURL, AllowedOrigin, DevToken          string
	SecretKey                                   []byte
	HTTPTimeout                                 time.Duration
	AllowedHosts                                []string
}

func Load() (Config, error) {
	c := Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"), RedisURL: os.Getenv("REDIS_URL"), Issuer: os.Getenv("OIDC_ISSUER"), ClientID: os.Getenv("OIDC_CLIENT_ID"), ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"), RedirectURL: os.Getenv("OIDC_REDIRECT_URL"), PublicURL: env("PUBLIC_URL", "http://localhost:4200"), AllowedOrigin: env("ALLOWED_ORIGIN", "http://localhost:4200"), DevToken: os.Getenv("DEV_TOKEN"), HTTPTimeout: 15 * time.Second}
	var err error
	c.SecretKey, err = base64.StdEncoding.DecodeString(os.Getenv("SECRET_KEY"))
	if err != nil || len(c.SecretKey) != 32 {
		return c, fmt.Errorf("SECRET_KEY must be a base64-encoded 32-byte key")
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	if c.DevToken != "" {
		if os.Getenv("APP_ENV") != "development" || len(c.DevToken) < 32 {
			return c, fmt.Errorf("DEV_TOKEN requires APP_ENV=development and at least 32 characters")
		}
	} else if c.Issuer == "" || c.ClientID == "" || c.RedirectURL == "" {
		return c, fmt.Errorf("OIDC_ISSUER, OIDC_CLIENT_ID and OIDC_REDIRECT_URL are required")
	}
	for _, raw := range []string{c.PublicURL, c.AllowedOrigin} {
		u, e := url.Parse(raw)
		if e != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
			return c, fmt.Errorf("invalid public URL or allowed origin")
		}
	}
	for _, host := range strings.Split(os.Getenv("OUTBOUND_ALLOWED_HOSTS"), ",") {
		if host = strings.TrimSpace(host); host != "" {
			c.AllowedHosts = append(c.AllowedHosts, host)
		}
	}
	return c, nil
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
