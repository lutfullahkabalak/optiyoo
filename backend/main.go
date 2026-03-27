package main

import (
	"log"
	"net/http"
	"optiyoo-backend/config"
	"optiyoo-backend/db"
	"optiyoo-backend/handlers"
	"optiyoo-backend/middleware"
	"optiyoo-backend/storage"
)

// corsMiddleware allows cross-origin requests from the Vue frontend
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

		// Preflight OPTIONS requests 
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("[optiyoo] veritabanı başlatılıyor…")
	db.InitDB()
	uploadRoot := config.UploadDir()
	ds, err := storage.NewDiskStore(uploadRoot)
	if err != nil {
		log.Fatalf("[optiyoo] yükleme dizini (%s): %v", uploadRoot, err)
	}
	handlers.BlobStore = ds
	log.Printf("[optiyoo] medya dizini: %s\n", uploadRoot)

	// Modern Go 1.22+ ServeMux Routing
	mux := http.NewServeMux()

	// Authenticated Routes
	mux.HandleFunc("POST /api/register", handlers.RegisterHandler)
	mux.HandleFunc("POST /api/login", handlers.LoginHandler)
	mux.HandleFunc("GET /api/users/{id}", handlers.GetUserHandler)
	mux.HandleFunc("PATCH /api/users/{id}", handlers.PatchUserHandler)

	// Survey Endpoints
	mux.HandleFunc("GET /api/surveys", handlers.GetSurveysHandler)
	mux.HandleFunc("GET /api/surveys/{id}", handlers.GetSurveyHandler)
	mux.HandleFunc("POST /api/surveys", handlers.CreateSurveyHandler)
	mux.HandleFunc("GET /api/config", handlers.GetConfigHandler)
	mux.HandleFunc("POST /api/surveys/{id}/answers", handlers.SubmitAnswersHandler)
	mux.HandleFunc("POST /api/media", handlers.UploadMediaHandler)
	mux.HandleFunc("GET /api/media/{id}", handlers.GetMediaHandler)

	// API Health Check Route
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "message": "Optiyoo Backend is running smoothly!"}`))
	})

	chain := middleware.RequestLog(corsMiddleware(mux))
	log.Println("[optiyoo] API http://127.0.0.1:8080 (istek logları [http], medya [media])")
	if err := http.ListenAndServe(":8080", chain); err != nil {
		log.Fatalf("[optiyoo] sunucu: %v", err)
	}
}
