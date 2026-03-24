package models

import "time"

// User represents an account in the system
type User struct {
	ID        string    `json:"id,omitzero"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitzero"` // Yalnızca istekte alınır, DB'ye yazılır, yanıtta boşaltılarak dışarı sızdırılmaz
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitzero"`
}

// Survey represents a single survey form
type Survey struct {
	ID          string     `json:"id,omitzero"`
	CreatorID   string     `json:"creator_id"` // Anket oluşturan kullanıcının ID'si
	CreatorName string     `json:"creator_name,omitempty"`
	Questions   []Question `json:"questions,omitzero"`
	UserAnswer  *string    `json:"user_answer,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitzero"`
	IsActive    bool       `json:"is_active"`
}

// Question represents a single question within a survey
type Question struct {
	ID       string   `json:"id,omitzero"`
	SurveyID string   `json:"survey_id"`
	Type     string   `json:"type"` // "single_choice", "text", "rating"
	Text     string   `json:"text"`
	Options  []Option `json:"options,omitzero"` // Sadece çoktan seçmeliyse dolar
	Order    int      `json:"order"`
}

// Option represents a choice in a multiple-choice question
type Option struct {
	ID         string `json:"id,omitzero"`
	QuestionID string `json:"question_id"`
	Text       string `json:"text"`
	VoteCount  int    `json:"vote_count"`
}

// Answer represents a user's submitted response to a specific question
type Answer struct {
	ID         string    `json:"id,omitzero"`
	SurveyID   string    `json:"survey_id"`
	QuestionID string    `json:"question_id"`
	UserID     string    `json:"user_id"`
	Value      string    `json:"value"` // Cevap içeriği (OptionID, Açık uçlu metin veya 1-5 puan)
	CreatedAt  time.Time `json:"created_at,omitzero"`
}
