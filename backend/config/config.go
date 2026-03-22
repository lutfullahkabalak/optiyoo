package config

var AppConfig = struct {
	AllowOpenEndedQuestions bool `json:"allow_open_ended"`
}{
	AllowOpenEndedQuestions: false, // Kullanıcı isteği doğrultusunda mevcuttur ancak varsayılan olarak kapalıdır.
}
