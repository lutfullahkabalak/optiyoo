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
  - `handlers/`: HTTP logic — `auth.go`, `config.go`, `survey.go`.
  - `models/models.go`: JSON/DB entity structs (`User`, `Survey`, `Question`, `Option`, `Answer`).
  - `config/config.go`: `AllowOpenEndedQuestions`, theme constants (`ThemeRoot`, `ThemeDark`, `ThemeWada1`–`3`), and `AppConfig` exposed as JSON via `/api/config`.
- `/frontend/`: Vue app.
  - `src/stores/auth.ts`: Session user in `localStorage` key `optiyoo_user`.
  - `src/views/DashboardView.vue`: Main poll feed; embeds `CreateSurveyModal`, uses `SurveyCard`-style patterns (poll list, instant vote).
  - `src/components/CreateSurveyModal.vue`: Poll creation (replaces the removed dedicated create page).
  - `src/components/SurveyCard.vue`: Reusable poll card UI.
  - `src/views/SurveyView.vue`: Deep link for a single survey (`/s/:id`).
  - `src/views/HomeView.vue`: Auth entry (route `/auth`).
- `/scripts/seed_test_data.sh`: Optional curl-based bulk seed against a running API (not required for core flow).

## 3. Frontend Routes (`frontend/src/router/index.ts`)
| Path | View | Role |
|------|------|------|
| `/auth` | `HomeView` | Login / register |
| `/` | `DashboardView` | Feed + create modal |
| `/s/:id` | `SurveyView` | Shareable survey page |

## 4. API Surface (high level)
- `POST /api/register`, `POST /api/login` — user JSON; passwords cleared on register response.
- `GET /api/config` — full `config.AppConfig` (open-ended flag + theme fields).
- `GET /api/surveys` — **optional** query `user_id`: when set, each survey may include `user_answer` if that user already voted (used to sync UI + completed set).
- `GET /api/surveys/{id}` — same `user_id` query semantics.
- `POST /api/surveys` — create survey (body: `Survey` with nested question/options); creator must send `creator_id` matching logged-in user.
- `POST /api/surveys/{id}/answers` — body `{ user_id, answers: [...] }`; duplicate vote for same `survey_id` + `user_id` returns **403** (checked in handler before inserts).
- `GET /api/health` — health JSON.

## 5. Core Business & Architecture Rules
- **Single-question polls:** `CreateSurveyHandler` rejects `len(s.Questions) != 1` with **400**.
- **Open-ended locking:** If any question has `type == "text"` and `AllowOpenEndedQuestions` is false, create returns **403**. Frontend should still read `/api/config` to disable text mode in the builder.
- **Poll type string:** The interactive UI treats the first question as a poll when `type === "single_choice"`. Legacy rows may use other type strings (e.g. old seeds); those will not render as the Twitter-style poll until aligned.
- **Unique IDs:** No UUIDs. `generateID` in `handlers/auth.go` uses `crypto/rand` (8 bytes → hex).
- **Interactive voting:** Feed/detail views use radio + immediate `POST` (no separate submit for the poll card flow); `403` from API locks further voting; UI also tracks completion in `localStorage` (see below).
- **Vote tallies:** `vote_count` on each `Option` is computed in SQL (subquery counting `answers` matching option id) in `GetSurveysHandler` / `GetSurveyHandler`; frontend draws percentage bars from these counts.
- **Active surveys only:** `GET /api/surveys` lists rows where `surveys.is_active = TRUE`. New creates set `is_active` true.
- **Rewards (+5 OPT):** Documented in product docs (e.g. README / `.cursorrules`) as a goal. **Current schema and `SubmitAnswersHandler` do not persist a points balance** — do not assume a `points` column or automatic rewards unless code and migrations add them.

## 6. localStorage (frontend)
- `optiyoo_user` — JSON user `{ id, name, email }` (Pinia `auth` store).
- `completed_polls_<userId>` — JSON array of survey IDs the client treats as already voted (must stay consistent with server; server is authoritative via `user_answer` / 403).

## 7. Database Schema (from `db/db.go`)
- `users(id, email UNIQUE, password, name, created_at)`
- `surveys(id, creator_id, is_active, created_at)`
- `questions(id, survey_id, type, text, q_order)`
- `options(id, question_id, text)`
- `answers(id, survey_id, question_id, user_id, value, created_at)` — duplicate prevention is **application-level** in `SubmitAnswersHandler` (not a DB UNIQUE constraint in current DDL).

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
