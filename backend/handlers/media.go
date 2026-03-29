package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"optiyoo-backend/db"
	"optiyoo-backend/imagemin"
	"optiyoo-backend/middleware"
	"optiyoo-backend/storage"
	"strconv"
	"strings"
	"time"
)

// BlobStore is set from main (disk today; swap for S3-backed implementation later).
var BlobStore storage.BlobStore

const (
	maxImageBytes = 5 * 1024 * 1024
)

var allowedImageMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func extForDetected(mime string) (string, bool) {
	// http.DetectContentType may return charset suffix; normalize.
	base := strings.Split(mime, ";")[0]
	base = strings.TrimSpace(base)
	ext, ok := allowedImageMIME[base]
	return ext, ok
}

// sniffImageMIME handles cases where DetectContentType returns application/octet-stream (common for some clients).
func sniffImageMIME(b []byte) string {
	if len(b) < 12 {
		return ""
	}
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return "image/jpeg"
	}
	if len(b) >= 8 && b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 && b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A {
		return "image/png"
	}
	if len(b) >= 6 && b[0] == 'G' && b[1] == 'I' && b[2] == 'F' && b[3] == '8' && (b[4] == '7' || b[4] == '9') && b[5] == 'a' {
		return "image/gif"
	}
	if len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return "image/webp"
	}
	return ""
}

func logMediaf(format string, args ...any) {
	log.Printf("[media] "+format, args...)
}

// UploadMediaHandler POST /api/media — multipart: user_id, survey_id, kind (question|option), ref_id, file
func UploadMediaHandler(w http.ResponseWriter, r *http.Request) {
	if BlobStore == nil {
		logMediaf("yükleme reddedildi: BlobStore nil")
		http.Error(w, "Depolama yapılandırılmamış.", http.StatusInternalServerError)
		return
	}
	// maxMemory: büyük parçalar diske yazılır; sınırı düşük tutmayın (multipart sınırı + dosya).
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		logMediaf("multipart ayrıştırma hatası: %v", err)
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

	surveyID := strings.TrimSpace(r.FormValue("survey_id"))
	kind := strings.TrimSpace(strings.ToLower(r.FormValue("kind")))
	refID := strings.TrimSpace(r.FormValue("ref_id"))

	if surveyID == "" || kind == "" || refID == "" {
		logMediaf("eksik alan: survey_id=%q kind=%q", surveyID, kind)
		http.Error(w, "survey_id, kind ve ref_id zorunludur.", http.StatusBadRequest)
		return
	}
	if kind != "question" && kind != "option" {
		logMediaf("geçersiz kind=%q", kind)
		http.Error(w, "kind değeri question veya option olmalıdır.", http.StatusBadRequest)
		return
	}

	var creatorID string
	err := db.DB.QueryRow("SELECT creator_id FROM surveys WHERE id = $1", surveyID).Scan(&creatorID)
	if err == sql.ErrNoRows {
		logMediaf("anket yok: survey_id=%s", surveyID)
		http.Error(w, "Anket bulunamadı.", http.StatusNotFound)
		return
	}
	if err != nil {
		logMediaf("anket sorgu hatası: %v", err)
		http.Error(w, "Anket doğrulanamadı.", http.StatusInternalServerError)
		return
	}
	if creatorID != uid {
		logMediaf("yetkisiz: survey_id=%s creator=%s uid=%s", surveyID, creatorID, uid)
		http.Error(w, "Bu anket için görsel yükleme yetkiniz yok.", http.StatusForbidden)
		return
	}

	if !refBelongsToSurvey(surveyID, kind, refID) {
		logMediaf("ref uyuşmuyor: survey_id=%s kind=%s ref_id=%s", surveyID, kind, refID)
		http.Error(w, "Seçilen soru veya seçenek bu ankete ait değil.", http.StatusBadRequest)
		return
	}

	file, hdr, err := r.FormFile("file")
	if err != nil {
		logMediaf("file alanı yok veya okunamadı: %v", err)
		http.Error(w, "file alanı zorunludur.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	payload, err := io.ReadAll(io.LimitReader(file, maxImageBytes+1))
	if err != nil {
		logMediaf("dosya okuma: %v", err)
		http.Error(w, "Dosya okunamadı.", http.StatusInternalServerError)
		return
	}
	if len(payload) == 0 {
		logMediaf("boş dosya survey_id=%s", surveyID)
		http.Error(w, "Boş dosya.", http.StatusBadRequest)
		return
	}
	if len(payload) > maxImageBytes {
		logMediaf("dosya çok büyük: %d bayt survey_id=%s", len(payload), surveyID)
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
		base := strings.TrimSpace(strings.Split(rawMIME, ";")[0])
		logMediaf("desteklenmeyen tür: algılanan=%q (ilk 16 bayt hex: %x) survey_id=%s", base, payload[:min(16, len(payload))], surveyID)
		http.Error(w, "İzin verilen türler: JPEG, PNG, WebP, GIF.", http.StatusBadRequest)
		return
	}
	contentType := strings.TrimSpace(strings.Split(rawMIME, ";")[0])
	_ = hdr

	origLen := len(payload)
	payload, ext, contentType = imagemin.Compress(payload, contentType)
	if len(payload) < origLen {
		logMediaf("sıkıştırma %d → %d bayt (%s)", origLen, len(payload), contentType)
	}

	ctx := r.Context()
	var mediaID, oldKey string
	err = db.DB.QueryRow(
		`SELECT id, storage_key FROM survey_media WHERE survey_id = $1 AND kind = $2 AND ref_id = $3`,
		surveyID, kind, refID,
	).Scan(&mediaID, &oldKey)
	if err != nil && err != sql.ErrNoRows {
		logMediaf("survey_media sorgu: %v", err)
		http.Error(w, "Medya kaydı okunamadı.", http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		mediaID = generateID()
	} else {
		if oldKey != "" {
			_ = BlobStore.Remove(ctx, oldKey)
		}
	}

	storageKey := fmt.Sprintf("surveys/%s/%s%s", surveyID, mediaID, ext)
	if putErr := BlobStore.Put(ctx, storageKey, bytes.NewReader(payload), maxImageBytes); putErr != nil {
		logMediaf("disk yazılamadı key=%s: %v", storageKey, putErr)
		http.Error(w, "Dosya kaydedilemedi.", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	if err == sql.ErrNoRows {
		_, err = db.DB.Exec(
			`INSERT INTO survey_media (id, survey_id, kind, ref_id, content_type, storage_key, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			mediaID, surveyID, kind, refID, contentType, storageKey, now,
		)
	} else {
		_, err = db.DB.Exec(
			`UPDATE survey_media SET content_type = $1, storage_key = $2, created_at = $3
			 WHERE id = $4`,
			contentType, storageKey, now, mediaID,
		)
	}
	if err != nil {
		logMediaf("DB yazılamadı survey_id=%s: %v", surveyID, err)
		_ = BlobStore.Remove(ctx, storageKey)
		http.Error(w, "Veritabanı güncellenemedi.", http.StatusInternalServerError)
		return
	}

	logMediaf("yüklendi survey_id=%s kind=%s ref_id=%s media_id=%s %d bayt %s ext=%s", surveyID, kind, refID, mediaID, len(payload), contentType, ext)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ID       string `json:"id"`
		ImageURL string `json:"image_url"`
	}{ID: mediaID, ImageURL: "/api/media/" + mediaID})
}

func refBelongsToSurvey(surveyID, kind, refID string) bool {
	if kind == "question" {
		var sid string
		err := db.DB.QueryRow(`SELECT survey_id FROM questions WHERE id = $1`, refID).Scan(&sid)
		return err == nil && sid == surveyID
	}
	var qSurvey string
	err := db.DB.QueryRow(
		`SELECT q.survey_id FROM options o JOIN questions q ON q.id = o.question_id WHERE o.id = $1`,
		refID,
	).Scan(&qSurvey)
	return err == nil && qSurvey == surveyID
}

// GetMediaHandler GET /api/media/{id}
func GetMediaHandler(w http.ResponseWriter, r *http.Request) {
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
	var active bool
	err := db.DB.QueryRow(`
		SELECT m.storage_key, m.content_type, COALESCE(s.is_active, FALSE)
		FROM survey_media m
		JOIN surveys s ON s.id = m.survey_id
		WHERE m.id = $1
	`, id).Scan(&storageKey, &contentType, &active)
	if err == sql.ErrNoRows {
		http.Error(w, "Bulunamadı.", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Okunamadı.", http.StatusInternalServerError)
		return
	}
	if !active {
		http.Error(w, "Bulunamadı.", http.StatusNotFound)
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
