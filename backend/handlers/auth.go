package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/models"
	"time"
)

// generateID creates a safe unique ID using crypto/rand
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RegisterHandler Handles POST /api/register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.ID = generateID()
	user.CreatedAt = time.Now()

	// Save to database
	_, err := db.DB.Exec("INSERT INTO users (id, email, password, name) VALUES ($1, $2, $3, $4)",
		user.ID, user.Email, user.Password, user.Name)

	if err != nil {
		http.Error(w, "E-posta kullanımda veya sistem hatası oluştu.", http.StatusInternalServerError)
		return
	}

	// Don't leak the password field even if it's stored
	user.Password = ""
	
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// LoginHandler Handles POST /api/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, email, name FROM users WHERE email = $1 AND password = $2",
		creds.Email, creds.Password).Scan(&user.ID, &user.Email, &user.Name)

	if err != nil {
		http.Error(w, "Hatalı şifre veya e-posta", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
