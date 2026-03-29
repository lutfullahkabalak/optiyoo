package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/middleware"
	"optiyoo-backend/models"
	"strings"
	"time"
)

type patchUserBody struct {
	CurrentPassword string  `json:"current_password"`
	Name            string  `json:"name"`
	Username        string  `json:"username"`
	Email           string  `json:"email"`
	NewPassword     string  `json:"new_password"`
	AvatarColor     *string `json:"avatar_color"`
	RemoveAvatar    bool    `json:"remove_avatar"`
}

func fillUserAvatarFields(u *models.User) {
	if u == nil || u.ID == "" {
		return
	}
	var mediaID sql.NullString
	var color sql.NullString
	err := db.DB.QueryRow(`
		SELECT NULLIF(TRIM(u.avatar_color), ''), um.id
		FROM users u
		LEFT JOIN user_media um ON um.user_id = u.id
		WHERE u.id = $1`, u.ID).Scan(&color, &mediaID)
	if err != nil {
		return
	}
	if color.Valid && color.String != "" {
		u.AvatarColor = color.String
	}
	if mediaID.Valid && mediaID.String != "" {
		u.AvatarURL = "/api/user-media/" + mediaID.String
	}
}

// GetUserHandler returns the current user (no password) for the authenticated account.
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pathID := strings.TrimSpace(r.PathValue("id"))
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || pathID == "" || pathID != uid {
		http.Error(w, "Yetkisiz.", http.StatusForbidden)
		return
	}

	var user models.User
	err := db.DB.QueryRow(
		`SELECT id, email, name, username, can_create_multi_question_surveys, created_at, updated_at FROM users WHERE id = $1`,
		uid,
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
	fillUserAvatarFields(&user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// PatchUserHandler updates display name, username, email, avatar fields, and/or password.
// JWT ownership is sufficient for non-password fields; current_password is required only when setting new_password.
func PatchUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pathID := strings.TrimSpace(r.PathValue("id"))
	if pathID == "" {
		http.Error(w, "Geçersiz kullanıcı.", http.StatusBadRequest)
		return
	}
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || pathID != uid {
		http.Error(w, "Yetkisiz işlem.", http.StatusForbidden)
		return
	}

	var body patchUserBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Geçersiz istek gövdesi.", http.StatusBadRequest)
		return
	}
	if body.NewPassword != "" && body.CurrentPassword == "" {
		http.Error(w, "Mevcut şifre gerekli.", http.StatusBadRequest)
		return
	}

	var curName, curEmail, curUsername, curPassword string
	err := db.DB.QueryRow(`SELECT name, email, username, password FROM users WHERE id = $1`, uid).Scan(&curName, &curEmail, &curUsername, &curPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "Kullanıcı bulunamadı.", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
		return
	}
	if body.NewPassword != "" && !passwordsMatch(curPassword, body.CurrentPassword) {
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
		h, hErr := hashPassword(body.NewPassword)
		if hErr != nil {
			http.Error(w, "Şifre güncellenemedi.", http.StatusInternalServerError)
			return
		}
		newP = h
	}

	var curAvatarColor sql.NullString
	var hadAvatarMedia bool
	_ = db.DB.QueryRow(`SELECT avatar_color FROM users WHERE id = $1`, uid).Scan(&curAvatarColor)
	_ = db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM user_media WHERE user_id = $1)`, uid).Scan(&hadAvatarMedia)
	curColorStr := ""
	if curAvatarColor.Valid {
		curColorStr = strings.TrimSpace(curAvatarColor.String)
	}

	avatarColorTouched := false
	if body.AvatarColor != nil {
		avatarColorTouched = true
		nv := strings.TrimSpace(*body.AvatarColor)
		if nv != "" && !hexColor7.MatchString(nv) {
			http.Error(w, "avatar_color #RRGGBB formatında olmalı veya boş olmalıdır.", http.StatusBadRequest)
			return
		}
	}
	avatarRemoved := body.RemoveAvatar && hadAvatarMedia
	colorChanged := false
	if body.AvatarColor != nil {
		nv := strings.TrimSpace(*body.AvatarColor)
		if nv != curColorStr {
			colorChanged = true
		}
	}

	profileChanged := newN != curName || newU != curUsername || newE != curEmail || newP != curPassword
	if !profileChanged && !avatarRemoved && !colorChanged {
		http.Error(w, "Değişiklik yok.", http.StatusBadRequest)
		return
	}

	if newU != curUsername {
		var other string
		err = db.DB.QueryRow(`SELECT id FROM users WHERE username = $1 AND id <> $2`, newU, uid).Scan(&other)
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
		err = db.DB.QueryRow(`SELECT id FROM users WHERE email = $1 AND id <> $2`, newE, uid).Scan(&other)
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
	if body.RemoveAvatar && hadAvatarMedia {
		if err := deleteUserAvatar(r.Context(), BlobStore, uid); err != nil {
			http.Error(w, "Profil resmi kaldırılamadı.", http.StatusInternalServerError)
			return
		}
	}

	if profileChanged {
		_, err = db.DB.Exec(
			`UPDATE users SET name = $1, username = $2, email = $3, password = $4, updated_at = $5 WHERE id = $6`,
			newN, newU, newE, newP, now, uid,
		)
		if err != nil {
			http.Error(w, "Güncelleme başarısız; e-posta veya kullanıcı adı çakışıyor olabilir.", http.StatusInternalServerError)
			return
		}
	} else if avatarColorTouched || avatarRemoved {
		_, err = db.DB.Exec(`UPDATE users SET updated_at = $1 WHERE id = $2`, now, uid)
		if err != nil {
			http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
			return
		}
	}

	if avatarColorTouched {
		nv := strings.TrimSpace(*body.AvatarColor)
		var colorArg interface{}
		if nv == "" {
			colorArg = nil
		} else {
			colorArg = nv
		}
		_, err = db.DB.Exec(`UPDATE users SET avatar_color = $1, updated_at = $2 WHERE id = $3`, colorArg, now, uid)
		if err != nil {
			http.Error(w, "Renk güncellenemedi.", http.StatusInternalServerError)
			return
		}
	}

	var user models.User
	err = db.DB.QueryRow(
		`SELECT id, email, name, username, can_create_multi_question_surveys, created_at, updated_at FROM users WHERE id = $1`,
		uid,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Username, &user.CanCreateMultiQuestionSurveys, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		http.Error(w, "Sunucu hatası.", http.StatusInternalServerError)
		return
	}
	user.Password = ""
	fillUserAvatarFields(&user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
