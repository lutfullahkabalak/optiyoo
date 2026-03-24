package config

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
