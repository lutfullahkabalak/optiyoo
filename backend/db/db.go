package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error

	// Varsayılan PostgreSQL bağlantı dizesi
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=optiyoo sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
	}

	// Bağlantıyı test et
	if err = DB.Ping(); err != nil {
		log.Printf("[db] UYARI: PostgreSQL ping başarısız (Docker çalışıyor mu?): %v\n", err)
		return // Uygulamayı kırma, belki container gecikmeli kalkıyordur
	}

	// Uygulama tablolarını oluştur (PostgreSQL formatında)
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE,
		password TEXT,
		name VARCHAR(255) NOT NULL,
		username VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS surveys (
		id TEXT PRIMARY KEY,
		creator_id TEXT,
		is_active BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS questions (
		id TEXT PRIMARY KEY,
		survey_id TEXT,
		type TEXT,
		text TEXT,
		q_order INTEGER
	);

	CREATE TABLE IF NOT EXISTS options (
		id TEXT PRIMARY KEY,
		question_id TEXT,
		text TEXT
	);

	CREATE TABLE IF NOT EXISTS answers (
		id TEXT PRIMARY KEY,
		survey_id TEXT,
		question_id TEXT,
		user_id TEXT,
		value TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS survey_media (
		id TEXT PRIMARY KEY,
		survey_id TEXT NOT NULL,
		kind TEXT NOT NULL,
		ref_id TEXT NOT NULL,
		content_type TEXT NOT NULL,
		storage_key TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (survey_id, kind, ref_id)
	);
	CREATE INDEX IF NOT EXISTS survey_media_survey_id_idx ON survey_media (survey_id);
	`
	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatalf("Tablolar oluşturulamadı: %v", err)
	}

	// Backward-compatible user feature flag migration.
	_, err = DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS can_create_multi_question_surveys BOOLEAN NOT NULL DEFAULT FALSE;`)
	if err != nil {
		log.Fatalf("users tablosu feature flag kolonu eklenemedi: %v", err)
	}

	// Username, updated_at; legacy points kolonunu kaldır.
	_, err = DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(255);`)
	if err != nil {
		log.Fatalf("users.username eklenemedi: %v", err)
	}
	_, err = DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;`)
	if err != nil {
		log.Fatalf("users.updated_at eklenemedi: %v", err)
	}
	_, err = DB.Exec(`UPDATE users SET updated_at = COALESCE(created_at, CURRENT_TIMESTAMP) WHERE updated_at IS NULL;`)
	if err != nil {
		log.Fatalf("users.updated_at doldurulamadı: %v", err)
	}
	_, err = DB.Exec(`UPDATE users SET username = 'user_' || id WHERE username IS NULL OR trim(username) = '';`)
	if err != nil {
		log.Fatalf("users.username doldurulamadı: %v", err)
	}
	_, err = DB.Exec(`ALTER TABLE users ALTER COLUMN username SET NOT NULL;`)
	if err != nil {
		log.Fatalf("users.username NOT NULL yapılamadı: %v", err)
	}
	_, err = DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_username_uq ON users (username);`)
	if err != nil {
		log.Fatalf("users.username benzersiz indeks oluşturulamadı: %v", err)
	}
	_, err = DB.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS points;`)
	if err != nil {
		log.Fatalf("users.points kaldırılamadı: %v", err)
	}

	// Seed Dummy Survey for MVP Evaluation
	DB.Exec(`
		INSERT INTO surveys (id, creator_id, is_active) SELECT 's1', 'u1', TRUE WHERE NOT EXISTS (SELECT 1 FROM surveys WHERE id = 's1');
		INSERT INTO questions (id, survey_id, type, text, q_order) SELECT 'q1', 's1', 'choice', 'Genel olarak bu uygulamadan ne kadar memnunsunuz?', 1 WHERE NOT EXISTS (SELECT 1 FROM questions WHERE id = 'q1');
		INSERT INTO questions (id, survey_id, type, text, q_order) SELECT 'q2', 's1', 'text', 'Geliştirilmesini istediğiniz en önemli özellik nedir?', 2 WHERE NOT EXISTS (SELECT 1 FROM questions WHERE id = 'q2');
		INSERT INTO options (id, question_id, text) SELECT 'o1', 'q1', 'Harika🥳' WHERE NOT EXISTS (SELECT 1 FROM options WHERE id = 'o1');
		INSERT INTO options (id, question_id, text) SELECT 'o2', 'q1', 'İdare Eder😐' WHERE NOT EXISTS (SELECT 1 FROM options WHERE id = 'o2');
		INSERT INTO options (id, question_id, text) SELECT 'o3', 'q1', 'Kötü👎' WHERE NOT EXISTS (SELECT 1 FROM options WHERE id = 'o3');
	`)

	log.Println("[db] PostgreSQL hazır (şema doğrulandı).")
}
