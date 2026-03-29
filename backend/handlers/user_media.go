package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"optiyoo-backend/db"
	"optiyoo-backend/imagemin"
	"optiyoo-backend/middleware"
	"optiyoo-backend/storage"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var hexColor7 = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// UploadUserAvatarHandler POST /api/user-media — multipart: file; isteğe bağlı avatar_color (#RRGGBB)
func UploadUserAvatarHandler(w http.ResponseWriter, r *http.Request) {
	if BlobStore == nil {
		http.Error(w, "Depolama yapılandırılmamış.", http.StatusInternalServerError)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Form okunamadı.", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Oturum gerekli.", http.StatusUnauthorized)
		return
	}

	ac := strings.TrimSpace(r.FormValue("avatar_color"))
	if ac != "" && !hexColor7.MatchString(ac) {
		http.Error(w, "avatar_color #RRGGBB formatında olmalıdır.", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file alanı zorunludur.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	payload, err := io.ReadAll(io.LimitReader(file, maxImageBytes+1))
	if err != nil {
		http.Error(w, "Dosya okunamadı.", http.StatusInternalServerError)
		return
	}
	if len(payload) == 0 {
		http.Error(w, "Boş dosya.", http.StatusBadRequest)
		return
	}
	if len(payload) > maxImageBytes {
		http.Error(w, "Dosya çok büyük (en fazla 5 MB).", http.StatusBadRequest)
		return
	}

	rawMIME := http.DetectContentType(payload)
	ext, ok := extForDetected(rawMIME)
	if !ok {
		if sniffed := sniffImageMIME(payload); sniffed != "" {
			rawMIME = sniffed
			ext, ok = extForDetected(sniffed)
		}
	}
	if !ok {
		http.Error(w, "İzin verilen türler: JPEG, PNG, WebP, GIF.", http.StatusBadRequest)
		return
	}
	contentType := strings.TrimSpace(strings.Split(rawMIME, ";")[0])

	origLen := len(payload)
	payload, ext, contentType = imagemin.Compress(payload, contentType)
	if len(payload) < origLen {
		logMediaf("kullanıcı avatar sıkıştırma %d → %d bayt (%s)", origLen, len(payload), contentType)
	}

	ctx := r.Context()
	var mediaID, oldKey string
	err = db.DB.QueryRow(`SELECT id, storage_key FROM user_media WHERE user_id = $1`, uid).Scan(&mediaID, &oldKey)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Medya kaydı okunamadı.", http.StatusInternalServerError)
		return
	}
	if err == sql.ErrNoRows {
		mediaID = generateID()
	} else if oldKey != "" {
		_ = BlobStore.Remove(ctx, oldKey)
	}

	storageKey := "users/" + uid + "/" + mediaID + ext
	if putErr := BlobStore.Put(ctx, storageKey, bytes.NewReader(payload), maxImageBytes); putErr != nil {
		http.Error(w, "Dosya kaydedilemedi.", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	if err == sql.ErrNoRows {
		_, err = db.DB.Exec(
			`INSERT INTO user_media (id, user_id, content_type, storage_key, created_at) VALUES ($1, $2, $3, $4, $5)`,
			mediaID, uid, contentType, storageKey, now,
		)
	} else {
		_, err = db.DB.Exec(
			`UPDATE user_media SET content_type = $1, storage_key = $2, created_at = $3 WHERE id = $4`,
			contentType, storageKey, now, mediaID,
		)
	}
	if err != nil {
		_ = BlobStore.Remove(ctx, storageKey)
		http.Error(w, "Veritabanı güncellenemedi.", http.StatusInternalServerError)
		return
	}

	if ac != "" {
		_, _ = db.DB.Exec(`UPDATE users SET avatar_color = $1, updated_at = $2 WHERE id = $3`, ac, now, uid)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ID       string `json:"id"`
		ImageURL string `json:"image_url"`
	}{ID: mediaID, ImageURL: "/api/user-media/" + mediaID})
}

// GetUserMediaHandler GET /api/user-media/{id} — herkese açık (profil resmi)
func GetUserMediaHandler(w http.ResponseWriter, r *http.Request) {
	if BlobStore == nil {
		http.Error(w, "Depolama yapılandırılmamış.", http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Geçersiz kimlik.", http.StatusBadRequest)
		return
	}

	var storageKey, contentType string
	err := db.DB.QueryRow(
		`SELECT storage_key, content_type FROM user_media WHERE id = $1`,
		id,
	).Scan(&storageKey, &contentType)
	if err == sql.ErrNoRows {
		http.Error(w, "Bulunamadı.", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Okunamadı.", http.StatusInternalServerError)
		return
	}

	rc, err := BlobStore.Open(context.Background(), storageKey)
	if err != nil {
		http.Error(w, "Dosya açılamadı.", http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if f, ok := rc.(*os.File); ok {
		if st, err := f.Stat(); err == nil && st.Size() >= 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(st.Size(), 10))
		}
	}
	_, _ = io.Copy(w, rc)
}

// deleteUserAvatar removes stored avatar file and DB row for the user.
func deleteUserAvatar(ctx context.Context, store storage.BlobStore, userID string) error {
	var key string
	err := db.DB.QueryRow(`SELECT storage_key FROM user_media WHERE user_id = $1`, userID).Scan(&key)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if key != "" {
		_ = store.Remove(ctx, key)
	}
	_, err = db.DB.Exec(`DELETE FROM user_media WHERE user_id = $1`, userID)
	return err
}
