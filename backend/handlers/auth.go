package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/middleware"
	"optiyoo-backend/models"
	"strings"
	"time"
)

// generateID creates a safe unique ID using crypto/rand
func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeAuthJSON(w http.ResponseWriter, status int, u models.User, token string) {
	fillUserAvatarFields(&u)
	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		ID                            string    `json:"id"`
		Email                         string    `json:"email"`
		Name                          string    `json:"name"`
		Username                      string    `json:"username"`
		AvatarURL                     string    `json:"avatar_url,omitempty"`
		AvatarColor                   string    `json:"avatar_color,omitempty"`
		CanCreateMultiQuestionSurveys bool      `json:"can_create_multi_question_surveys"`
		CreatedAt                     time.Time `json:"created_at"`
		UpdatedAt                     time.Time `json:"updated_at"`
		Token                         string    `json:"token"`
	}{
		ID:                            u.ID,
		Email:                         u.Email,
		Name:                          u.Name,
		Username:                      u.Username,
		AvatarURL:                     u.AvatarURL,
		AvatarColor:                   u.AvatarColor,
		CanCreateMultiQuestionSurveys: u.CanCreateMultiQuestionSurveys,
		CreatedAt:                     u.CreatedAt,
		UpdatedAt:                     u.UpdatedAt,
		Token:                         token,
	})
}

// RegisterHandler Handles POST /api/register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Geçersiz istek gövdesi.", http.StatusBadRequest)
		return
	}

	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	if user.Username == "" {
		http.Error(w, "Kullanıcı adı zorunludur.", http.StatusBadRequest)
		return
	}
	if user.Email == "" {
		http.Error(w, "E-posta zorunludur.", http.StatusBadRequest)
		return
	}
	if len(user.Password) < 6 {
		http.Error(w, "Şifre en az 6 karakter olmalıdır.", http.StatusBadRequest)
		return
	}

	hash, err := hashPassword(user.Password)
	if err != nil {
		http.Error(w, "Kayıt işlenemedi.", http.StatusInternalServerError)
		return
	}
	user.Password = hash

	user.ID = generateID()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err = db.DB.Exec("INSERT INTO users (id, email, password, name, username, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		user.ID, user.Email, user.Password, user.Name, user.Username, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		http.Error(w, "E-posta veya kullanıcı adı kullanımda; ayrıca sistem hatası da oluşmuş olabilir.", http.StatusInternalServerError)
		return
	}

	_ = db.DB.QueryRow(`SELECT can_create_multi_question_surveys FROM users WHERE id = $1`, user.ID).Scan(&user.CanCreateMultiQuestionSurveys)

	token, err := middleware.SignUserToken(user.ID)
	if err != nil {
		http.Error(w, "Oturum oluşturulamadı.", http.StatusInternalServerError)
		return
	}

	writeAuthJSON(w, http.StatusCreated, user, token)
}

// LoginHandler Handles POST /api/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Geçersiz istek gövdesi.", http.StatusBadRequest)
		return
	}
	creds.Email = strings.TrimSpace(creds.Email)

	var user models.User
	var storedPassword string
	err := db.DB.QueryRow(`SELECT id, email, password, name, username, can_create_multi_question_surveys, created_at, updated_at FROM users WHERE email = $1`,
		creds.Email).Scan(&user.ID, &user.Email, &storedPassword, &user.Name, &user.Username, &user.CanCreateMultiQuestionSurveys, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "Hatalı şifre veya e-posta", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
		return
	}

	if !passwordsMatch(storedPassword, creds.Password) {
		http.Error(w, "Hatalı şifre veya e-posta", http.StatusUnauthorized)
		return
	}

	if !isBcryptHash(storedPassword) {
		newHash, hErr := hashPassword(creds.Password)
		if hErr == nil {
			_, _ = db.DB.Exec(`UPDATE users SET password = $1, updated_at = $2 WHERE id = $3`, newHash, time.Now(), user.ID)
		}
	}

	token, err := middleware.SignUserToken(user.ID)
	if err != nil {
		http.Error(w, "Oturum oluşturulamadı.", http.StatusInternalServerError)
		return
	}

	writeAuthJSON(w, http.StatusOK, user, token)
}
