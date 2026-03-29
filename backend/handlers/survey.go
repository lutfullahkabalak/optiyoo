package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"optiyoo-backend/config"
	"optiyoo-backend/db"
	"optiyoo-backend/middleware"
	"optiyoo-backend/models"
	"strings"
	"time"

	"github.com/lib/pq"
)

// attachUserAnswers loads all of the user's answers for this survey into UserAnswers
// and sets UserAnswer to the first question's selection when present.
func attachUserAnswers(s *models.Survey, userID string) {
	if userID == "" {
		return
	}
	ansRows, err := db.DB.Query("SELECT question_id, value FROM answers WHERE survey_id = $1 AND user_id = $2", s.ID, userID)
	if err != nil {
		return
	}
	defer ansRows.Close()
	m := make(map[string]string)
	for ansRows.Next() {
		var qid, val string
		if ansRows.Scan(&qid, &val) != nil || qid == "" || val == "" {
			continue
		}
		m[qid] = val
	}
	if len(m) == 0 {
		return
	}
	s.UserAnswers = m
	if len(s.Questions) > 0 {
		if v, ok := m[s.Questions[0].ID]; ok {
			vCopy := v
			s.UserAnswer = &vCopy
		}
	}
}

// attachSurveyImages sets question/option ImageURL from survey_media rows.
func attachSurveyImages(s *models.Survey) {
	if s == nil || s.ID == "" {
		return
	}
	rows, err := db.DB.Query(`SELECT kind, ref_id, id FROM survey_media WHERE survey_id = $1`, s.ID)
	if err != nil {
		return
	}
	defer rows.Close()
	qmap := make(map[string]string)
	omap := make(map[string]string)
	for rows.Next() {
		var kind, refID, mid string
		if rows.Scan(&kind, &refID, &mid) != nil || refID == "" || mid == "" {
			continue
		}
		u := "/api/media/" + mid
		switch kind {
		case "question":
			qmap[refID] = u
		case "option":
			omap[refID] = u
		}
	}
	for i := range s.Questions {
		if u, ok := qmap[s.Questions[i].ID]; ok {
			s.Questions[i].ImageURL = u
		}
		for j := range s.Questions[i].Options {
			if u, ok := omap[s.Questions[i].Options[j].ID]; ok {
				s.Questions[i].Options[j].ImageURL = u
			}
		}
	}
}

// GetSurveysHandler lists all active surveys GET /api/surveys
func GetSurveysHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID != "" {
		bearer, err := middleware.ParseBearerUserID(r)
		if err != nil || bearer != userID {
			http.Error(w, "Yetkisiz.", http.StatusForbidden)
			return
		}
	}
	rows, err := db.DB.Query(`SELECT s.id, s.creator_id, COALESCE(u.name, ''), COALESCE(u.username, ''), u.avatar_color, um.id, s.created_at
		FROM surveys s
		LEFT JOIN users u ON u.id = s.creator_id
		LEFT JOIN user_media um ON um.user_id = s.creator_id
		WHERE s.is_active = TRUE ORDER BY s.created_at DESC`)
	if err != nil {
		http.Error(w, "Anketler yüklenemedi.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var surveys []models.Survey
	for rows.Next() {
		var s models.Survey
		var creatorColor, creatorMediaID sql.NullString
		if err := rows.Scan(&s.ID, &s.CreatorID, &s.CreatorName, &s.CreatorUsername, &creatorColor, &creatorMediaID, &s.CreatedAt); err == nil {
			if creatorColor.Valid && strings.TrimSpace(creatorColor.String) != "" {
				s.CreatorAvatarColor = strings.TrimSpace(creatorColor.String)
			}
			if creatorMediaID.Valid && creatorMediaID.String != "" {
				s.CreatorAvatarURL = "/api/user-media/" + creatorMediaID.String
			}
			fillSurveyListPayload(&s, userID)
			surveys = append(surveys, s)
		}
	}

	if surveys == nil {
		surveys = []models.Survey{} // Prevent null inside JSON
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(surveys)
}

// GetSurveyHandler returns a specific survey with its nested questions GET /api/surveys/{id}
func GetSurveyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID != "" {
		bearer, err := middleware.ParseBearerUserID(r)
		if err != nil || bearer != userID {
			http.Error(w, "Yetkisiz.", http.StatusForbidden)
			return
		}
	}

	var s models.Survey
	var creatorColor, creatorMediaID sql.NullString
	err := db.DB.QueryRow(`SELECT s.id, s.creator_id, COALESCE(u.name, ''), COALESCE(u.username, ''), u.avatar_color, um.id, s.created_at
		FROM surveys s
		LEFT JOIN users u ON u.id = s.creator_id
		LEFT JOIN user_media um ON um.user_id = s.creator_id
		WHERE s.id = $1 AND s.is_active = TRUE`, id).
		Scan(&s.ID, &s.CreatorID, &s.CreatorName, &s.CreatorUsername, &creatorColor, &creatorMediaID, &s.CreatedAt)
	if err == nil {
		if creatorColor.Valid && strings.TrimSpace(creatorColor.String) != "" {
			s.CreatorAvatarColor = strings.TrimSpace(creatorColor.String)
		}
		if creatorMediaID.Valid && creatorMediaID.String != "" {
			s.CreatorAvatarURL = "/api/user-media/" + creatorMediaID.String
		}
	}

	if err != nil {
		http.Error(w, "Anket bulunamadı", http.StatusNotFound)
		return
	}

	fillSurveyListPayload(&s, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// fillSurveyListPayload loads questions, options, vote counts, media URLs, and optional user answers for list/detail-shaped survey rows.
func fillSurveyListPayload(s *models.Survey, userID string) {
	qRows, _ := db.DB.Query("SELECT id, type, text, q_order FROM questions WHERE survey_id = $1 ORDER BY q_order", s.ID)
	defer qRows.Close()
	s.Questions = nil
	for qRows.Next() {
		var q models.Question
		qRows.Scan(&q.ID, &q.Type, &q.Text, &q.Order)

		if q.Type != "text" {
			oRows, _ := db.DB.Query("SELECT o.id, o.text, (SELECT COUNT(*) FROM answers a WHERE a.question_id = $1 AND a.value = o.id) as vote_count FROM options o WHERE o.question_id = $1", q.ID)
			for oRows.Next() {
				var opt models.Option
				oRows.Scan(&opt.ID, &opt.Text, &opt.VoteCount)
				q.Options = append(q.Options, opt)
			}
			oRows.Close()
		}
		s.Questions = append(s.Questions, q)
	}

	attachSurveyImages(s)
	if userID != "" {
		attachUserAnswers(s, userID)
	}
}

// loadActiveSurveysOrdered returns active surveys in the same order as ids (skips missing/inactive ids).
func loadActiveSurveysOrdered(ids []string, userID string) ([]models.Survey, error) {
	if len(ids) == 0 {
		return []models.Survey{}, nil
	}
	rows, err := db.DB.Query(`SELECT s.id, s.creator_id, COALESCE(u.name, ''), COALESCE(u.username, ''), u.avatar_color, um.id, s.created_at
		FROM surveys s
		LEFT JOIN users u ON u.id = s.creator_id
		LEFT JOIN user_media um ON um.user_id = s.creator_id
		WHERE s.is_active = TRUE AND s.id = ANY($1)
		ORDER BY array_position($2::text[], s.id::text)`,
		pq.Array(ids), pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Survey
	for rows.Next() {
		var s models.Survey
		var creatorColor, creatorMediaID sql.NullString
		if err := rows.Scan(&s.ID, &s.CreatorID, &s.CreatorName, &s.CreatorUsername, &creatorColor, &creatorMediaID, &s.CreatedAt); err != nil {
			continue
		}
		if creatorColor.Valid && strings.TrimSpace(creatorColor.String) != "" {
			s.CreatorAvatarColor = strings.TrimSpace(creatorColor.String)
		}
		if creatorMediaID.Valid && creatorMediaID.String != "" {
			s.CreatorAvatarURL = "/api/user-media/" + creatorMediaID.String
		}
		fillSurveyListPayload(&s, userID)
		out = append(out, s)
	}
	return out, nil
}

// SubmitAnswersHandler processes user answers POST /api/surveys/{id}/answers
// Her istekte bir veya birden fazla soru cevabı kabul edilir; aynı (anket, kullanıcı, soru) için tekrar 403 döner.
func SubmitAnswersHandler(w http.ResponseWriter, r *http.Request) {
	surveyID := r.PathValue("id")

	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Oturum gerekli.", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Answers []models.Answer `json:"answers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Geçersiz istek gövdesi.", http.StatusBadRequest)
		return
	}

	if len(payload.Answers) == 0 {
		http.Error(w, "En az bir cevap göndermeniz gerekiyor.", http.StatusBadRequest)
		return
	}

	var surveyActive bool
	if err := db.DB.QueryRow(`SELECT is_active FROM surveys WHERE id = $1`, surveyID).Scan(&surveyActive); err == sql.ErrNoRows {
		http.Error(w, "Anket bulunamadı.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Anket doğrulanamadı.", http.StatusInternalServerError)
		return
	}
	if !surveyActive {
		http.Error(w, "Bu anket artık aktif değil.", http.StatusGone)
		return
	}

	questionRows, err := db.DB.Query("SELECT id, type FROM questions WHERE survey_id = $1", surveyID)
	if err != nil {
		http.Error(w, "Anket soruları doğrulanamadı.", http.StatusInternalServerError)
		return
	}
	defer questionRows.Close()

	expectedQuestions := make(map[string]struct{})
	qTypes := make(map[string]string)
	for questionRows.Next() {
		var questionID, qType string
		if scanErr := questionRows.Scan(&questionID, &qType); scanErr != nil {
			http.Error(w, "Anket soruları okunamadı.", http.StatusInternalServerError)
			return
		}
		expectedQuestions[questionID] = struct{}{}
		qTypes[questionID] = qType
	}

	if len(expectedQuestions) == 0 {
		http.Error(w, "Bu ankette cevaplanacak soru bulunamadı.", http.StatusBadRequest)
		return
	}

	seenInPayload := make(map[string]struct{})
	for _, ans := range payload.Answers {
		if ans.QuestionID == "" || ans.Value == "" {
			http.Error(w, "Her cevap için soru kimliği ve değer zorunludur.", http.StatusBadRequest)
			return
		}
		if _, ok := expectedQuestions[ans.QuestionID]; !ok {
			http.Error(w, "Bu ankete ait olmayan bir soruya cevap gönderildi.", http.StatusBadRequest)
			return
		}
		if _, dup := seenInPayload[ans.QuestionID]; dup {
			http.Error(w, "Aynı soruya tek istekte birden fazla cevap gönderilemez.", http.StatusBadRequest)
			return
		}
		seenInPayload[ans.QuestionID] = struct{}{}

		switch qTypes[ans.QuestionID] {
		case "single_choice", "choice":
			var n int
			if err := db.DB.QueryRow(
				`SELECT COUNT(*) FROM options WHERE id = $1 AND question_id = $2`,
				ans.Value, ans.QuestionID,
			).Scan(&n); err != nil || n != 1 {
				http.Error(w, "Geçersiz seçenek.", http.StatusBadRequest)
				return
			}
		case "text":
			if len(ans.Value) > 4000 {
				http.Error(w, "Metin cevabı çok uzun.", http.StatusBadRequest)
				return
			}
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "İşlem başlatılamadı.", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, ans := range payload.Answers {
		var existing int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM answers WHERE survey_id = $1 AND user_id = $2 AND question_id = $3",
			surveyID, uid, ans.QuestionID,
		).Scan(&existing); err != nil {
			http.Error(w, "Mükerrer yanıt kontrolü başarısız.", http.StatusInternalServerError)
			return
		}
		if existing > 0 {
			http.Error(w, "Bu soru için zaten yanıt verdiniz.", http.StatusForbidden)
			return
		}

		ans.ID = generateID()
		ans.CreatedAt = time.Now()
		if _, err := tx.Exec(
			"INSERT INTO answers (id, survey_id, question_id, user_id, value, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			ans.ID, surveyID, ans.QuestionID, uid, ans.Value, ans.CreatedAt,
		); err != nil {
			http.Error(w, "Cevap kaydedilemedi.", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Cevaplar tamamlanamadı.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Cevap kaydedildi."}`))
}

// CreateSurveyHandler Handles POST /api/surveys
func CreateSurveyHandler(w http.ResponseWriter, r *http.Request) {
	var s models.Survey
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Geçersiz istek gövdesi.", http.StatusBadRequest)
		return
	}

	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Oturum gerekli.", http.StatusUnauthorized)
		return
	}
	s.CreatorID = uid

	var canCreateMultiQuestion bool
	err := db.DB.QueryRow("SELECT can_create_multi_question_surveys FROM users WHERE id = $1", s.CreatorID).Scan(&canCreateMultiQuestion)
	if err == sql.ErrNoRows {
		http.Error(w, "Anket oluşturacak kullanıcı bulunamadı.", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Kullanıcı yetkisi doğrulanamadı.", http.StatusInternalServerError)
		return
	}

	if len(s.Questions) == 0 {
		http.Error(w, "Anket en az bir soru içermelidir.", http.StatusBadRequest)
		return
	}
	if !canCreateMultiQuestion && len(s.Questions) != 1 {
		http.Error(w, "Bu hesap için çok sorulu anket özelliği aktif değil.", http.StatusForbidden)
		return
	}

	for _, q := range s.Questions {
		if q.Type == "text" && !config.AppConfig.AllowOpenEndedQuestions {
			http.Error(w, "Sistem yapılandırması gereği açık uçlu (kısa metin) sorulara şu an izin verilmemektedir.", http.StatusForbidden)
			return
		}
	}

	s.ID = generateID()
	s.CreatedAt = time.Now()
	s.IsActive = true

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction başlatılamadı", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO surveys (id, creator_id, created_at, is_active) VALUES ($1, $2, $3, $4)",
		s.ID, s.CreatorID, s.CreatedAt, s.IsActive)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Anket kaydedilemedi", http.StatusInternalServerError)
		return
	}

	for i := range s.Questions {
		q := &s.Questions[i]
		q.ID = generateID()
		q.SurveyID = s.ID
		_, err = tx.Exec("INSERT INTO questions (id, survey_id, type, text, q_order) VALUES ($1, $2, $3, $4, $5)",
			q.ID, s.ID, q.Type, q.Text, q.Order)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Soru kaydedilemedi", http.StatusInternalServerError)
			return
		}

		if q.Type != "text" {
			for j := range q.Options {
				o := &q.Options[j]
				o.ID = generateID()
				o.QuestionID = q.ID
				_, err = tx.Exec("INSERT INTO options (id, question_id, text) VALUES ($1, $2, $3)",
					o.ID, q.ID, o.Text)
				if err != nil {
					tx.Rollback()
					http.Error(w, "Seçenek kaydedilemedi.", http.StatusInternalServerError)
					return
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Transaction bitirilemedi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}
