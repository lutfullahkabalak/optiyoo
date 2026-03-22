package main

import (
	"log"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/handlers"
)

// corsMiddleware allows cross-origin requests from the Vue frontend
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Preflight OPTIONS requests 
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Database başlatılıyor...")
	db.InitDB()

	// Modern Go 1.22+ ServeMux Routing
	mux := http.NewServeMux()

	// Authenticated Routes
	mux.HandleFunc("POST /api/register", handlers.RegisterHandler)
	mux.HandleFunc("POST /api/login", handlers.LoginHandler)

	// Survey Endpoints
	mux.HandleFunc("GET /api/surveys", handlers.GetSurveysHandler)
	mux.HandleFunc("GET /api/surveys/{id}", handlers.GetSurveyHandler)
	mux.HandleFunc("POST /api/surveys", handlers.CreateSurveyHandler)
	mux.HandleFunc("GET /api/config", handlers.GetConfigHandler)
	mux.HandleFunc("POST /api/surveys/{id}/answers", handlers.SubmitAnswersHandler)

	// API Health Check Route
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "message": "Optiyoo Backend is running smoothly!"}`))
	})

	log.Println("CORS eklendi. Sunucu port :8080 üzerinde çalışıyor...")
	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		log.Fatalf("Sunucu hatası: %v", err)
	}
}
