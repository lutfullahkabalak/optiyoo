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
		log.Println("UYARI: PostgreSQL bağlantısı başarısız. Lütfen Docker üzerinden veritabanının çalıştığından emin olun.")
		log.Printf("Hata detayı: %v\n", err)
		return // Uygulamayı kırma, belki container gecikmeli kalkıyordur
	}

	// Uygulama tablolarını oluştur (PostgreSQL formatında)
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE,
		password TEXT,
		name TEXT,
		points INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS surveys (
		id TEXT PRIMARY KEY,
		title TEXT,
		description TEXT,
		creator_id TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT TRUE
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
	`
	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatalf("Tablolar oluşturulamadı: %v", err)
	}
	
	// Seed Dummy Survey for MVP Evaluation
	DB.Exec(`
		INSERT INTO surveys (id, title, description, creator_id) SELECT 'survey_1', 'Genel Memnuniyet Testi', 'Hizmetlerimizi nasıl buluyorsunuz? Fikirleriniz ve eleştrileriniz bizim için çok değerli!', 'system' WHERE NOT EXISTS (SELECT 1 FROM surveys WHERE id = 'survey_1');
		INSERT INTO questions (id, survey_id, type, text, q_order) SELECT 'q1', 'survey_1', 'choice', 'Genel olarak bu uygulamadan ne kadar memnunsunuz?', 1 WHERE NOT EXISTS (SELECT 1 FROM questions WHERE id = 'q1');
		INSERT INTO questions (id, survey_id, type, text, q_order) SELECT 'q2', 'survey_1', 'text', 'Geliştirilmesini istediğiniz en önemli özellik nedir?', 2 WHERE NOT EXISTS (SELECT 1 FROM questions WHERE id = 'q2');
		INSERT INTO options (id, question_id, text) SELECT 'o1', 'q1', 'Harika🥳' WHERE NOT EXISTS (SELECT 1 FROM options WHERE id = 'o1');
		INSERT INTO options (id, question_id, text) SELECT 'o2', 'q1', 'İdare Eder😐' WHERE NOT EXISTS (SELECT 1 FROM options WHERE id = 'o2');
		INSERT INTO options (id, question_id, text) SELECT 'o3', 'q1', 'Kötü👎' WHERE NOT EXISTS (SELECT 1 FROM options WHERE id = 'o3');
	`)
	
	log.Println("PostgreSQL veritabanı başarıyla bağlandı ve tablolar doğrulandı.")
}
