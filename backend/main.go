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

// corsMiddleware allows the configured browser origin (OPTYOO_CORS_ORIGIN) to call the API with Authorization.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowOrigin := config.ResolveCORSAllowOrigin(r.Header.Get("Origin"))
		if allowOrigin == "" {
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Vary", "Origin")

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
	if config.UsingDevJWTSecret() {
		log.Println("[optiyoo] UYARI: OPTYOO_JWT_SECRET tanımlı değil; geliştirme anahtarı kullanılıyor. Üretimde güçlü bir gizli anahtar ayarlayın.")
	}
	uploadRoot := config.UploadDir()
	ds, err := storage.NewDiskStore(uploadRoot)
	if err != nil {
		log.Fatalf("[optiyoo] yükleme dizini (%s): %v", uploadRoot, err)
	}
	handlers.BlobStore = ds
	log.Printf("[optiyoo] medya dizini: %s\n", uploadRoot)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", handlers.RegisterHandler)
	mux.HandleFunc("POST /api/login", handlers.LoginHandler)

	mux.Handle("GET /api/users/{id}", middleware.RequireAuth(http.HandlerFunc(handlers.GetUserHandler)))
	mux.Handle("PATCH /api/users/{id}", middleware.RequireAuth(http.HandlerFunc(handlers.PatchUserHandler)))

	mux.HandleFunc("GET /api/surveys", handlers.GetSurveysHandler)
	mux.HandleFunc("GET /api/search", handlers.SearchSurveysHandler)
	mux.HandleFunc("GET /api/surveys/{id}", handlers.GetSurveyHandler)
	mux.Handle("POST /api/surveys", middleware.RequireAuth(http.HandlerFunc(handlers.CreateSurveyHandler)))
	mux.HandleFunc("GET /api/config", handlers.GetConfigHandler)
	mux.Handle("POST /api/surveys/{id}/answers", middleware.RequireAuth(http.HandlerFunc(handlers.SubmitAnswersHandler)))
	mux.Handle("POST /api/media", middleware.RequireAuth(http.HandlerFunc(handlers.UploadMediaHandler)))
	mux.HandleFunc("GET /api/media/{id}", handlers.GetMediaHandler)
	mux.Handle("POST /api/user-media", middleware.RequireAuth(http.HandlerFunc(handlers.UploadUserAvatarHandler)))
	mux.HandleFunc("GET /api/user-media/{id}", handlers.GetUserMediaHandler)

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "message": "Optiyoo Backend is running smoothly!"}`))
	})

	apiChain := middleware.SecurityHeaders(corsMiddleware(mux))
	chain := middleware.RequestLog(apiChain)
	log.Println("[optiyoo] API http://127.0.0.1:8080 (istek logları [http], medya [media])")
	if err := http.ListenAndServe(":8080", chain); err != nil {
		log.Fatalf("[optiyoo] sunucu: %v", err)
	}
}
