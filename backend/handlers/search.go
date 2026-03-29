package handlers

import (
	"encoding/json"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/middleware"
	"optiyoo-backend/models"
	"strings"
	"unicode/utf8"
)

const searchQueryMaxRunes = 200

// escapeILIKEPattern escapes \, %, _ for use in ILIKE ... ESCAPE '\'.
func escapeILIKEPattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// SearchSurveysHandler GET /api/search?q=... — aktif anketlerde oluşturucu adı/kullanıcı adı, soru metni, seçenek metni ve kayıtlı cevap değerlerinde arar.
func SearchSurveysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Yöntem desteklenmiyor.", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID != "" {
		bearer, err := middleware.ParseBearerUserID(r)
		if err != nil || bearer != userID {
			http.Error(w, "Yetkisiz.", http.StatusForbidden)
			return
		}
	}

	raw := strings.TrimSpace(r.URL.Query().Get("q"))
	if raw == "" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
		return
	}
	if utf8.RuneCountInString(raw) > searchQueryMaxRunes {
		http.Error(w, "Arama metni çok uzun.", http.StatusBadRequest)
		return
	}

	pat := "%" + escapeILIKEPattern(raw) + "%"

	rows, err := db.DB.Query(`
		SELECT s.id
		FROM surveys s
		WHERE s.is_active = TRUE
		AND (
			EXISTS (
				SELECT 1 FROM users u
				WHERE u.id = s.creator_id
				AND (u.username ILIKE $1 ESCAPE '\' OR u.name ILIKE $1 ESCAPE '\')
			)
			OR EXISTS (
				SELECT 1 FROM questions q
				WHERE q.survey_id = s.id AND q.text ILIKE $1 ESCAPE '\'
			)
			OR EXISTS (
				SELECT 1 FROM questions q
				INNER JOIN options o ON o.question_id = q.id
				WHERE q.survey_id = s.id AND o.text ILIKE $1 ESCAPE '\'
			)
			OR EXISTS (
				SELECT 1 FROM answers a
				WHERE a.survey_id = s.id AND a.value ILIKE $1 ESCAPE '\'
			)
		)
		ORDER BY s.created_at DESC`,
		pat)
	if err != nil {
		http.Error(w, "Arama yapılamadı.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil && id != "" {
			ids = append(ids, id)
		}
	}

	surveys, err := loadActiveSurveysOrdered(ids, userID)
	if err != nil {
		http.Error(w, "Sonuçlar yüklenemedi.", http.StatusInternalServerError)
		return
	}
	if surveys == nil {
		surveys = []models.Survey{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(surveys)
}
