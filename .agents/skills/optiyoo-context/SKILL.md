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
  - `main.go`: Route registration, CORS (`OPTYOO_CORS_ORIGIN` virgülle çoklu köken; varsayılan `http://localhost:5173`; `OPTYOO_JWT_SECRET` yokken `https://*.trycloudflare.com` kökenleri de izinli), güvenlik başlıkları, JWT korumalı mutasyonlar, port `:8080`.
  - `middleware/auth.go`: `Authorization: Bearer` JWT (HS256, `OPTYOO_JWT_SECRET`; geliştirmede yerleşik zayıf varsayılan), `RequireAuth`, `ParseBearerUserID`.
  - `middleware/security.go`: `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`.
  - `handlers/password.go`: bcrypt ile şifre hash; eski düz metin satırları ilk başarılı girişte hash’e çevrilir.
  - `db/db.go`: PostgreSQL connection, schema DDL, optional inline seed rows.
  - `handlers/`: HTTP logic — `auth.go`, `config.go`, `survey.go`, `media.go` (multipart upload + binary GET).
  - `storage/storage.go`: `BlobStore` interface + `DiskStore` (files under `config.UploadDir()` / env `OPTYOO_UPLOAD_DIR`, default `data/uploads/`); swap implementation later for S3-compatible storage using the same logical `storage_key`.
  - `imagemin/imagemin.go`: `POST /api/media` sırasında raster görseller yeniden kodlanır (JPEG/WebP → kalite 85 JPEG; PNG/statik GIF → zlib sıkıştırmalı PNG); uzun kenar `MaxEdgePixels` (1920) üstü oran korunarak küçültülür; çok kareli animasyonlu GIF aynen saklanır. Çıktı orijinalden büyükse ham dosya tutulur.
  - `models/models.go`: JSON/DB entity structs (`User`, `Survey`, `Question`, `Option`, `Answer`). `Question` / `Option` expose optional `image_url` in JSON when a row exists in `survey_media`.
  - `config/config.go`: `AllowOpenEndedQuestions`, `UploadDir()` / `OPTYOO_UPLOAD_DIR`, theme constants (`ThemeRoot`, `ThemeDark`, `ThemeWada1`–`3`), and `AppConfig` exposed as JSON via `/api/config`.
- `/frontend/`: Vue app.
  - `src/router/index.ts`: Tüm route view bileşenleri dinamik `import()` ile yüklenir. Anket kartlarında `SurveyUserHeader` / `SurveyQuestionBlock` görselleri `img loading="lazy"` kullanır; `CreateSurveyModal` ve `AvatarCropModal` sırasıyla `DashboardView` / `ProfileSettingsView` içinde `defineAsyncComponent` ile ilk açılışta parça yüklenir.
  - `src/stores/auth.ts`: Session user in `localStorage` key `optiyoo_user`.
  - `src/views/DashboardView.vue`: Main poll feed; embeds `CreateSurveyModal`, uses `SurveyCard`-style patterns (poll list, instant vote).
  - `src/components/CreateSurveyModal.vue`: Poll creation (replaces the removed dedicated create page); after `POST /api/surveys` uploads optional per-question / per-option images via `POST /api/media` (multipart), matching array order to returned IDs.
  - `src/components/survey/SurveyQuestionBlock.vue`: Renders `question.image_url` and `option.image_url` when present (`apiBase` prop for absolute URLs).
  - `src/components/SurveyCard.vue`: Reusable poll card UI.
  - `src/views/SurveyView.vue`: Deep link for a single survey (`/s/:id`).
  - `src/views/HomeView.vue`: Auth entry (route `/auth`).
  - `src/views/ProfileSettingsView.vue`: Profil ayarları (`/settings`); `GET` + `PATCH /api/users/{id}`; profil resmi `POST /api/user-media` + `AvatarCropModal.vue` (daire kırpma), `avatar_color` paleti.
- `/scripts/seed_test_data.sh`: Optional curl-based bulk seed against a running API (not required for core flow).

## 3. Frontend Routes (`frontend/src/router/index.ts`)
| Path | View | Role |
|------|------|------|
| `/auth` | `HomeView` | Login / register |
| `/` | `DashboardView` | Feed + create modal |
| `/search` | `SearchView` | `?q=` ile API araması; ana sayfadaki arama buraya yönlendirir |
| `/settings` | `ProfileSettingsView` | Profil: kullanıcı adı, e-posta, şifre (kenar çubuğu Profil veya sağ panel kullanıcı kartı) |
| `/s/:id` | `SurveyView` | Shareable survey page |

## 4. API Surface (high level)
- **Kimlik doğrulama:** `POST /api/register` ve `POST /api/login` yanıtında kullanıcı alanları + `token` (JWT). Korunan uçlarda `Authorization: Bearer <token>` zorunlu. Şifreler bcrypt ile saklanır; DB’de düz metin kalan eski kayıtlar ilk doğru girişte otomatik hash’lenir.
- `POST /api/register` — kayıt; şifre yanıtta dönmez. Kayıtta şifre en az 6 karakter.
- `POST /api/login` — e-posta + şifre; başarıda `token` + kullanıcı JSON.
- `GET /api/users/{id}` — **Bearer gerekli**; path `id` JWT `sub` ile aynı olmalı; şifresiz kullanıcı JSON (`avatar_url`, `avatar_color` dahil).
- `PATCH /api/users/{id}` — **Bearer gerekli**; path `id` JWT ile aynı olmalı. Gövde: isteğe bağlı `name` (≤255), `username`, `email`, `avatar_color` (`#RRGGBB` veya boş string = sıfırla), `remove_avatar` (bool, profil fotoğrafını kaldırır), `new_password` (≥6; yalnızca bu alan doluysa `current_password` zorunlu ve bcrypt / geçiş dönemi düz metin ile doğrulanır). Yeni şifre her zaman bcrypt yazılır.
- `GET /api/config` — full `config.AppConfig` (open-ended flag + theme fields).
- `GET /api/surveys` — **optional** query `user_id`: dolu ise aynı kullanıcıya ait `Authorization: Bearer` zorunlu (`sub` = `user_id`); aksi **403**. `user_id` boşsa herkese açık liste. `user_id` set iken `user_answers` / `user_answer` senkronu; her kayıtta `creator_name`, `creator_username`, isteğe bağlı `creator_avatar_url`, `creator_avatar_color`.
- `GET /api/search` — query `q` (zorunlu, boşsa `[]`); aktif anketlerde oluşturucu `username` / `name`, soru metni, seçenek metni ve `answers.value` (oy / metin cevabı) üzerinde büyük/küçük harf duyarsız alt dize araması. İsteğe bağlı `user_id` + Bearer kuralları `GET /api/surveys` ile aynı. Yanıt gövdesi anket listesi ile aynı şekilde tam iç içe anket dizisi. Performans: Postgres `pg_trgm` + GIN indeksleri (`db.go` başlangıcında, uzantı yoksa log ile atlanır).
- `GET /api/surveys/{id}` — yalnızca `is_active = TRUE` anketler (**pasif veya yok → 404**). `user_id` query varsa Bearer ile `sub` eşleşmesi zorunlu (yukarıdaki gibi).
- `POST /api/surveys` — **Bearer gerekli**; oluşturan kimlik yalnızca JWT’den alınır (gövdedeki `creator_id` yok sayılır). Body: `Survey` + iç içe soru/seçenekler.
- `POST /api/surveys/{id}/answers` — **Bearer gerekli**; oy veren kullanıcı JWT `sub` (gövdede `user_id` yok). Body: `{ answers: [...] }`. `single_choice` / `choice` için `value` ilgili sorunun geçerli `options.id` olmalı; `text` için uzunluk sınırı 4000. Pasif anket **410**. Aynı `survey_id` + `user_id` + `question_id` için ikinci kayıt **403**.
- `POST /api/media` — **Bearer gerekli**; yükleyen = JWT `sub` ve anketin `creator_id` ile eşleşmeli. `multipart/form-data`: `survey_id`, `kind` (`question` | `option`), `ref_id`, `file` (JPEG/PNG/WebP/GIF, max 5MB). Yanıt: `id`, `image_url` `/api/media/{id}`.
- `POST /api/user-media` — **Bearer gerekli**; yükleyen = JWT `sub` (kendi profil resmi). `multipart/form-data`: `file` (JPEG/PNG/WebP/GIF, max 5MB), isteğe bağlı `avatar_color` (`#RRGGBB`). Yanıt: `id`, `image_url` `/api/user-media/{id}`; kullanıcı başına tek kayıt (üstüne yazar).
- `GET /api/media/{id}` — yalnızca `surveys.is_active = TRUE` olan anketlere bağlı medya için dosya akışı; aksi 404.
- `GET /api/user-media/{id}` — profil resmi dosyası; herkese açık okuma (akış).
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
- `optiyoo_user` — JSON user `{ id, name, email, username, avatar_url?, avatar_color?, ... }` (Pinia `auth` store).
- `optiyoo_token` — JWT access token; korunan API isteklerinde `Authorization` başlığında kullanılır. Kullanıcı kaydı token olmadan kalırsa (eski oturum) istemci tutarsız sayılır ve `optiyoo_user` temizlenebilir.
- `completed_polls_<userId>` — JSON array of survey IDs the client treats as **tamamen** cevaplanmış; kısmi çok sorulu anketler listede olmayabilir, kilit durumu `user_answers` ile senkronlanır.

## 7. Database Schema (from `db/db.go`)
- İsteğe bağlı: `CREATE EXTENSION pg_trgm` ve GIN trigram indeksleri — `users.name`, `users.username`, `questions.text`, `options.text`, `answers.value` (arama sorguları için).
- `users(id, email UNIQUE, password, name, username UNIQUE, can_create_multi_question_surveys, avatar_color VARCHAR(7) NULL, created_at, updated_at)`
- `user_media(id, user_id UNIQUE, content_type, storage_key, created_at)` — kullanıcı profil resmi (disk: `users/{user_id}/{id}{ext}`).
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
