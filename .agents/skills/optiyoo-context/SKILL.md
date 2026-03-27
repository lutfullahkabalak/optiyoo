---
name: optiyoo-context
description: MUST BE LOADED in new sessions. Contains the core architecture, tech stack, database schema, and exact business rules for the Optiyoo platform. Prevents unnecessary codebase scanning. When behavior changes in backend/models, router, or config, update this file in the same PR or immediately after.
---

# Optiyoo Platform Context & Architecture

This file acts as the ultimate reference point for AI Agents to understand the Optiyoo application without needing to deep-scan source code files during new sessions.

## 1. Tech Stack
- **Backend:** Go (Golang) 1.22+. Uses strictly standard library `net/http` for routing (`http.ServeMux` with Go 1.22+ path patterns). No Gin/Fiber.
- **Frontend:** Vue 3 via Vite. Uses `<script setup lang="ts">`, Vue Router, and Pinia for state management.
- **Database:** PostgreSQL (via `lib/pq` Go driver).
- **Styling:** Vanilla CSS without any framework (e.g., Tailwind). Design tokens are custom-built based on Sanzo Wada's Color Combination #109; **themes** are driven by `GET /api/config` and applied on `document.body` (see `frontend/src/App.vue`).

## 2. Directory Structure (Monorepo)
- `/backend/`: Contains the Go application.
  - `main.go`: Route registration, CORS middleware, port `:8080`.
  - `db/db.go`: PostgreSQL connection, schema DDL, optional inline seed rows.
  - `handlers/`: HTTP logic — `auth.go`, `config.go`, `survey.go`, `media.go` (multipart upload + binary GET).
  - `storage/storage.go`: `BlobStore` interface + `DiskStore` (files under `config.UploadDir()` / env `OPTYOO_UPLOAD_DIR`, default `data/uploads/`); swap implementation later for S3-compatible storage using the same logical `storage_key`.
  - `imagemin/imagemin.go`: `POST /api/media` sırasında raster görseller yeniden kodlanır (JPEG/WebP → kalite 85 JPEG; PNG/statik GIF → zlib sıkıştırmalı PNG); uzun kenar `MaxEdgePixels` (1920) üstü oran korunarak küçültülür; çok kareli animasyonlu GIF aynen saklanır. Çıktı orijinalden büyükse ham dosya tutulur.
  - `models/models.go`: JSON/DB entity structs (`User`, `Survey`, `Question`, `Option`, `Answer`). `Question` / `Option` expose optional `image_url` in JSON when a row exists in `survey_media`.
  - `config/config.go`: `AllowOpenEndedQuestions`, `UploadDir()` / `OPTYOO_UPLOAD_DIR`, theme constants (`ThemeRoot`, `ThemeDark`, `ThemeWada1`–`3`), and `AppConfig` exposed as JSON via `/api/config`.
- `/frontend/`: Vue app.
  - `src/stores/auth.ts`: Session user in `localStorage` key `optiyoo_user`.
  - `src/views/DashboardView.vue`: Main poll feed; embeds `CreateSurveyModal`, uses `SurveyCard`-style patterns (poll list, instant vote).
  - `src/components/CreateSurveyModal.vue`: Poll creation (replaces the removed dedicated create page); after `POST /api/surveys` uploads optional per-question / per-option images via `POST /api/media` (multipart), matching array order to returned IDs.
  - `src/components/survey/SurveyQuestionBlock.vue`: Renders `question.image_url` and `option.image_url` when present (`apiBase` prop for absolute URLs).
  - `src/components/SurveyCard.vue`: Reusable poll card UI.
  - `src/views/SurveyView.vue`: Deep link for a single survey (`/s/:id`).
  - `src/views/HomeView.vue`: Auth entry (route `/auth`).
  - `src/views/ProfileSettingsView.vue`: Profil ayarları (`/settings`); `GET` + `PATCH /api/users/{id}`.
- `/scripts/seed_test_data.sh`: Optional curl-based bulk seed against a running API (not required for core flow).

## 3. Frontend Routes (`frontend/src/router/index.ts`)
| Path | View | Role |
|------|------|------|
| `/auth` | `HomeView` | Login / register |
| `/` | `DashboardView` | Feed + create modal |
| `/settings` | `ProfileSettingsView` | Profil: kullanıcı adı, e-posta, şifre (kenar çubuğu Profil veya sağ panel kullanıcı kartı) |
| `/s/:id` | `SurveyView` | Shareable survey page |

## 4. API Surface (high level)
- `POST /api/register`, `POST /api/login` — user JSON; passwords cleared on register response.
- `GET /api/users/{id}?user_id={id}` — oturumdaki kullanıcıyı doğrular (`user_id` path ile aynı olmalı); şifresiz kullanıcı JSON.
- `PATCH /api/users/{id}` — gövde: `user_id` (path ile aynı), `current_password` (zorunlu), isteğe bağlı `name` (görünen ad, en fazla 255 karakter), `username`, `email`, `new_password` (en az 6 karakter). En az bir alan gerçekten değişmeli; mevcut şifre yanlışsa **401**, çakışma **409** metinleri; yanıtta güncel kullanıcı (şifresiz).
- `GET /api/config` — full `config.AppConfig` (open-ended flag + theme fields).
- `GET /api/surveys` — **optional** query `user_id`: when set, each survey may include `user_answers` (`question_id` → seçilen `option_id` map) if that user already voted; `user_answer` remains the first question’s value for backward compatibility (used to sync UI + completed set). Her kayıtta `creator_name` ve `creator_username` (oluşturan `users` join) döner.
- `GET /api/surveys/{id}` — same `user_id` query semantics; `creator_username` dahil.
- `POST /api/surveys` — create survey (body: `Survey` with nested question/options); creator must send `creator_id` matching logged-in user.
- `POST /api/surveys/{id}/answers` — body `{ user_id, answers: [...] }`; aynı `survey_id` + `user_id` + `question_id` için ikinci kayıt **403** (handler, insert öncesi). İstekte tek soru veya birden fazla soru olabilir; tüm soruların tek seferde dolması zorunlu değildir.
- `POST /api/media` — `multipart/form-data`: `user_id` (anket `creator_id` ile eşleşmeli), `survey_id`, `kind` (`question` | `option`), `ref_id` (soru veya seçenek id), `file` (JPEG/PNG/WebP/GIF, max 5MB). Sunucu kabul sonrası görseli `imagemin` ile sıkıştırır/yeniden boyutlandırır (API sözleşmesi değişmez). Yanıt JSON: `id`, `image_url` path `/api/media/{id}`. Aynı `(survey_id, kind, ref_id)` için tekrar yükleme eski dosyayı değiştirir.
- `GET /api/media/{id}` — yalnızca `surveys.is_active = TRUE` olan anketlere bağlı medya için dosya akışı; aksi 404.
- `GET /api/health` — health JSON.

## 5. Core Business & Architecture Rules
- **Multi-question create permission (per-user):** `users.can_create_multi_question_surveys` controls whether creator can submit 1+ questions. Default is **false**. If false, create is restricted to a single question.
- **Open-ended locking:** If any question has `type == "text"` and `AllowOpenEndedQuestions` is false, create returns **403**. Frontend should still read `/api/config` to disable text mode in the builder.
- **Poll type string:** The interactive UI treats the first question as a poll when `type === "single_choice"`. Legacy rows may use other type strings (e.g. old seeds); those will not render as the Twitter-style poll until aligned.
- **Unique IDs:** No UUIDs. `generateID` in `handlers/auth.go` uses `crypto/rand` (8 bytes → hex).
- **Interactive voting:** Feed/detail views keep answers per question, allow question navigation with left/right arrows for multi-question surveys, and **each choice triggers an immediate `POST`** with a single (or batched) answer payload—çok soruda tüm anketin bitmesi beklenmez. **Per-question lock:** kullanıcı bir soruda seçim yaptıktan sonra o soru yenilenene kadar değiştirilemez (çubuklar + devre dışı girişler). `403` aynı soruya tekrar yanıt denemesinde; `completed_polls_*` yalnızca anketteki tüm sorular yanıtlandığında güncellenir (kısmi ilerlemede sunucu `user_answers` otoriter).
- **Vote tallies:** `vote_count` on each `Option` is computed in SQL (subquery counting `answers` matching option id) in `GetSurveysHandler` / `GetSurveyHandler`; frontend draws percentage bars from these counts.
- **Active surveys only:** `GET /api/surveys` lists rows where `surveys.is_active = TRUE`. New creates set `is_active` true.
- **Survey images:** Binary files live on disk (or future object store) keyed by `survey_media.storage_key`. `GET /api/surveys` and `GET /api/surveys/{id}` attach `image_url` on each `Question` / `Option` when a `survey_media` row exists (`kind` + `ref_id`). Public image fetch requires the parent survey to be active.
- **Rewards (+5 OPT):** Documented in product docs (e.g. README / `.cursorrules`) as a goal. **Current schema and `SubmitAnswersHandler` do not persist a points balance** — do not assume a `points` column or automatic rewards unless code and migrations add them.

## 6. localStorage (frontend)
- `optiyoo_user` — JSON user `{ id, name, email, username, ... }` (Pinia `auth` store).
- `completed_polls_<userId>` — JSON array of survey IDs the client treats as **tamamen** cevaplanmış; kısmi çok sorulu anketler listede olmayabilir, kilit durumu `user_answers` ile senkronlanır.

## 7. Database Schema (from `db/db.go`)
- `users(id, email UNIQUE, password, name, username UNIQUE, can_create_multi_question_surveys, created_at, updated_at)`
- `surveys(id, creator_id, is_active, created_at)`
- `questions(id, survey_id, type, text, q_order)`
- `options(id, question_id, text)`
- `answers(id, survey_id, question_id, user_id, value, created_at)` — duplicate prevention is **application-level** in `SubmitAnswersHandler` (not a DB UNIQUE constraint in current DDL).
- `survey_media(id, survey_id, kind, ref_id, content_type, storage_key, created_at)` — `UNIQUE (survey_id, kind, ref_id)`; `kind` is `question` (ref = `questions.id`) or `option` (ref = `options.id`). Index on `survey_id`.

## 8. Common Developer Commands
- Database (Docker): `docker-compose up -d` in project root.
- Backend: `cd backend && go run .`
- Frontend: `cd frontend && npm run dev`
- Optional API seed: `./scripts/seed_test_data.sh` (with backend reachable; see script env `BASE_URL`).

---

## 9. Bu skill’in güncel kalması — zorunlu süreç

Aşağıdaki değişikliklerden **herhangi biri** yapıldığında, **aynı değişiklik kümesinde** veya hemen ardından bu `SKILL.md` dosyası güncellenmelidir (aksi halde agent’lar eski varsayımlarla kod üretir).

1. **Yeni veya kaldırılan API route** → Bölüm 4 ve gerekirse `backend/main.go` özeti.
2. **`config.AppConfig` veya tema sabitleri** → Bölüm 1–2 ve iş kuralları (açık uçlu / tema).
3. **`models` alanları veya JSON etiketleri** → İlgili struct davranışı ve API örnekleri.
4. **Şema veya benzersizlik kuralları** (`db.go`) → Bölüm 7; handler’da duplicate/validasyon mantığı değiştiyse Bölüm 5.
5. **Yeni route veya `localStorage` anahtarı** → Bölüm 3 ve 6.
6. **Anket oluşturma / liste UI’sinin taşınması** (ör. modal vs sayfa, yeni bileşen) → Bölüm 2 dosya listesi.
7. **Ürün kuralı kodla uyumlu hale geldiğinde** (ör. gerçekten puan tablosu ve oy sonrası güncelleme eklendiğinde) → Bölüm 5’teki “Rewards” maddesini koddaki gerçekle eşitleyin; README / `.cursorrules` ile çelişi varsa onları da aynı PR’da düzeltin.

**PR / commit disiplini:** Mimari veya sözleşme (API, şema, depolama anahtarları) değişen her PR’da kontrol listesi: “`optiyoo-context/SKILL.md` güncellendi mi?” — Hayırsa ya skill güncellenir ya değişiklik kapsamı skill’e dokunmuyor diye bilinçli olarak sınırlanır.
