package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/models"
	"strings"
	"time"
)

type patchUserBody struct {
	UserID          string `json:"user_id"`
	CurrentPassword string `json:"current_password"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	NewPassword     string `json:"new_password"`
}

// GetUserHandler returns the current user (no password) when user_id query matches path id.
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	q := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if id == "" || q == "" || q != id {
		http.Error(w, "Yetkisiz.", http.StatusForbidden)
		return
	}

	var user models.User
	err := db.DB.QueryRow(
		`SELECT id, email, name, username, can_create_multi_question_surveys, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Username, &user.CanCreateMultiQuestionSurveys, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Kullanıcı bulunamadı.", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
		return
	}
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// PatchUserHandler updates display name, username, email, and/or password after verifying current password.
func PatchUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		http.Error(w, "Geçersiz kullanıcı.", http.StatusBadRequest)
		return
	}

	var body patchUserBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.UserID) != id {
		http.Error(w, "Yetkisiz işlem.", http.StatusForbidden)
		return
	}
	if body.CurrentPassword == "" {
		http.Error(w, "Mevcut şifre gerekli.", http.StatusBadRequest)
		return
	}

	var curName, curEmail, curUsername, curPassword string
	err := db.DB.QueryRow(`SELECT name, email, username, password FROM users WHERE id = $1`, id).Scan(&curName, &curEmail, &curUsername, &curPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "Kullanıcı bulunamadı.", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
		return
	}
	if curPassword != body.CurrentPassword {
		http.Error(w, "Mevcut şifre hatalı.", http.StatusUnauthorized)
		return
	}

	newN := curName
	if strings.TrimSpace(body.Name) != "" {
		newN = strings.TrimSpace(body.Name)
		if len(newN) > 255 {
			http.Error(w, "Görünen ad en fazla 255 karakter olabilir.", http.StatusBadRequest)
			return
		}
	}
	newU := curUsername
	if strings.TrimSpace(body.Username) != "" {
		newU = strings.TrimSpace(body.Username)
	}
	newE := curEmail
	if strings.TrimSpace(body.Email) != "" {
		newE = strings.TrimSpace(body.Email)
	}
	newP := curPassword
	if body.NewPassword != "" {
		if len(body.NewPassword) < 6 {
			http.Error(w, "Yeni şifre en az 6 karakter olmalı.", http.StatusBadRequest)
			return
		}
		newP = body.NewPassword
	}

	if newN == curName && newU == curUsername && newE == curEmail && newP == curPassword {
		http.Error(w, "Değişiklik yok.", http.StatusBadRequest)
		return
	}

	if newU != curUsername {
		var other string
		err = db.DB.QueryRow(`SELECT id FROM users WHERE username = $1 AND id <> $2`, newU, id).Scan(&other)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
			return
		}
		if err == nil {
			http.Error(w, "Bu kullanıcı adı zaten kullanılıyor.", http.StatusConflict)
			return
		}
	}
	if newE != curEmail {
		var other string
		err = db.DB.QueryRow(`SELECT id FROM users WHERE email = $1 AND id <> $2`, newE, id).Scan(&other)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
			return
		}
		if err == nil {
			http.Error(w, "Bu e-posta zaten kayıtlı.", http.StatusConflict)
			return
		}
	}

	now := time.Now()
	_, err = db.DB.Exec(
		`UPDATE users SET name = $1, username = $2, email = $3, password = $4, updated_at = $5 WHERE id = $6`,
		newN, newU, newE, newP, now, id,
	)
	if err != nil {
		http.Error(w, "Güncelleme başarısız; e-posta veya kullanıcı adı çakışıyor olabilir.", http.StatusInternalServerError)
		return
	}

	var user models.User
	err = db.DB.QueryRow(
		`SELECT id, email, name, username, can_create_multi_question_surveys, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Username, &user.CanCreateMultiQuestionSurveys, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
		return
	}
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
