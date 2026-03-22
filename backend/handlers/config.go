package handlers

import (
	"encoding/json"
	"net/http"
	"optiyoo-backend/config"
)

// GetConfigHandler Handles GET /api/config
func GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.AppConfig)
}
