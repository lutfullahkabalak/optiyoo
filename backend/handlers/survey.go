package handlers

import (
	"encoding/json"
	"net/http"
	"optiyoo-backend/config"
	"optiyoo-backend/db"
	"optiyoo-backend/models"
	"time"
)

// GetSurveysHandler lists all active surveys GET /api/surveys
func GetSurveysHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	rows, err := db.DB.Query("SELECT s.id, s.creator_id, COALESCE(u.name, ''), s.created_at FROM surveys s LEFT JOIN users u ON u.id = s.creator_id WHERE s.is_active = TRUE ORDER BY s.created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var surveys []models.Survey
	for rows.Next() {
		var s models.Survey
		if err := rows.Scan(&s.ID, &s.CreatorID, &s.CreatorName, &s.CreatedAt); err == nil {
			// Yükleme (Join mantığı)
			qRows, _ := db.DB.Query("SELECT id, type, text, q_order FROM questions WHERE survey_id = $1 ORDER BY q_order", s.ID)
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
			qRows.Close()

			if userID != "" {
				var userAnswer string
				err := db.DB.QueryRow("SELECT value FROM answers WHERE survey_id = $1 AND user_id = $2", s.ID, userID).Scan(&userAnswer)
				if err == nil && userAnswer != "" {
					s.UserAnswer = &userAnswer
				}
			}

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
	userID := r.URL.Query().Get("user_id")

	var s models.Survey
	err := db.DB.QueryRow("SELECT s.id, s.creator_id, COALESCE(u.name, ''), s.created_at FROM surveys s LEFT JOIN users u ON u.id = s.creator_id WHERE s.id = $1", id).
		Scan(&s.ID, &s.CreatorID, &s.CreatorName, &s.CreatedAt)

	if err != nil {
		http.Error(w, "Anket bulunamadı", http.StatusNotFound)
		return
	}

	qRows, _ := db.DB.Query("SELECT id, type, text, q_order FROM questions WHERE survey_id = $1 ORDER BY q_order", s.ID)
	defer qRows.Close()

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

	if userID != "" {
		var userAnswer string
		err := db.DB.QueryRow("SELECT value FROM answers WHERE survey_id = $1 AND user_id = $2", s.ID, userID).Scan(&userAnswer)
		if err == nil && userAnswer != "" {
			s.UserAnswer = &userAnswer
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// SubmitAnswersHandler processes user answers and adds point balance POST /api/surveys/{id}/answers
func SubmitAnswersHandler(w http.ResponseWriter, r *http.Request) {
	surveyID := r.PathValue("id")
	
	var payload struct {
		UserID  string          `json:"user_id"`
		Answers []models.Answer `json:"answers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if payload.UserID != "" {
		var existing int
		db.DB.QueryRow("SELECT COUNT(*) FROM answers WHERE survey_id = $1 AND user_id = $2", surveyID, payload.UserID).Scan(&existing)
		if existing > 0 {
			http.Error(w, "Bu oylamada zaten oy kullandınız.", http.StatusForbidden)
			return
		}
	}

	for _, ans := range payload.Answers {
		ans.ID = generateID()
		ans.CreatedAt = time.Now()
		db.DB.Exec("INSERT INTO answers (id, survey_id, question_id, user_id, value, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			ans.ID, surveyID, ans.QuestionID, payload.UserID, ans.Value, ans.CreatedAt)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Anket başarıyla gönderildi!"}`))
}

// CreateSurveyHandler Handles POST /api/surveys
func CreateSurveyHandler(w http.ResponseWriter, r *http.Request) {
	var s models.Survey
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(s.Questions) != 1 {
		http.Error(w, "Sistem kuralları gereği her anket yalnızca tam olarak 1 (tek) sorudan oluşabilir.", http.StatusBadRequest)
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

	for _, q := range s.Questions {
		q.ID = generateID()
		_, err = tx.Exec("INSERT INTO questions (id, survey_id, type, text, q_order) VALUES ($1, $2, $3, $4, $5)",
			q.ID, s.ID, q.Type, q.Text, q.Order)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Soru kaydedilemedi", http.StatusInternalServerError)
			return
		}

		if q.Type != "text" {
			for _, o := range q.Options {
				o.ID = generateID()
				_, err = tx.Exec("INSERT INTO options (id, question_id, text) VALUES ($1, $2, $3)",
					o.ID, q.ID, o.Text)
				if err != nil {
					tx.Rollback()
					http.Error(w, "Seçenek kaydedilemedi: "+err.Error(), http.StatusInternalServerError)
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

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
