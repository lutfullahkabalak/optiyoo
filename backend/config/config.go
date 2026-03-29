package config

import (
	"net/url"
	"os"
	"strings"
)

const EnvUploadDir = "OPTYOO_UPLOAD_DIR"

const (
	EnvJWTSecret    = "OPTYOO_JWT_SECRET"
	EnvCORSOrigin   = "OPTYOO_CORS_ORIGIN"
	devJWTSecret    = "optiyoo-dev-jwt-secret-do-not-use-in-production"
	defaultCORSOrig = "http://localhost:5173"
)

// JWTSecret returns the HMAC key for access tokens. Production must set OPTYOO_JWT_SECRET.
func JWTSecret() string {
	if s := os.Getenv(EnvJWTSecret); s != "" {
		return s
	}
	return devJWTSecret
}

// UsingDevJWTSecret reports whether the process relies on the built-in dev signing key.
func UsingDevJWTSecret() bool {
	return os.Getenv(EnvJWTSecret) == ""
}

func corsOriginsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv(EnvCORSOrigin))
	if raw == "" {
		return []string{defaultCORSOrig}
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return []string{defaultCORSOrig}
	}
	return out
}

// CORSAllowOrigin returns the first configured origin (backward compatible).
func CORSAllowOrigin() string {
	return corsOriginsFromEnv()[0]
}

// ResolveCORSAllowOrigin returns the value for Access-Control-Allow-Origin, or "" if the request must not be allowed.
// OPTYOO_CORS_ORIGIN virgülle ayrılmış birden fazla köken içerebilir.
// Geliştirmede (OPTYOO_JWT_SECRET yok) https://*.trycloudflare.com kökenleri de kabul edilir (Cloudflare Quick Tunnel).
func ResolveCORSAllowOrigin(requestOrigin string) string {
	list := corsOriginsFromEnv()
	if requestOrigin == "" {
		return list[0]
	}
	for _, o := range list {
		if o == requestOrigin {
			return requestOrigin
		}
	}
	if UsingDevJWTSecret() && isTryCloudflareTunnelOrigin(requestOrigin) {
		return requestOrigin
	}
	return ""
}

func isTryCloudflareTunnelOrigin(s string) bool {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return false
	}
	return strings.HasSuffix(u.Hostname(), ".trycloudflare.com")
}

// UploadDir returns the root directory for uploaded blobs (disk store). Override with OPTYOO_UPLOAD_DIR.
func UploadDir() string {
	if s := os.Getenv(EnvUploadDir); s != "" {
		return s
	}
	return "data/uploads"
}

type Theme string

const (
	ThemeRoot  Theme = "theme-root"
	ThemeDark  Theme = "theme-dark"
	ThemeWada1 Theme = "theme-wada-1"
	ThemeWada2 Theme = "theme-wada-2"
	ThemeWada3 Theme = "theme-wada-3"
)

var AppConfig = struct {
	AllowOpenEndedQuestions bool    `json:"allow_open_ended"`
	ActiveTheme             string  `json:"active_theme"`
	DefaultTheme            string  `json:"default_theme"`
	Themes                  []Theme `json:"themes"`
}{
	AllowOpenEndedQuestions: false, // Kullanıcı isteği doğrultusunda mevcuttur ancak varsayılan olarak kapalıdır.
	ActiveTheme:             string(ThemeRoot),
	DefaultTheme:            string(ThemeDark),
	Themes:                  []Theme{ThemeRoot, ThemeDark, ThemeWada1, ThemeWada2, ThemeWada3},
}
