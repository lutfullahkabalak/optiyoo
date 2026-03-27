package config

import "os"

const EnvUploadDir = "OPTYOO_UPLOAD_DIR"

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
	DefaultTheme:            string(ThemeWada2),
	Themes:                  []Theme{ThemeRoot, ThemeDark, ThemeWada1, ThemeWada2, ThemeWada3},
}
