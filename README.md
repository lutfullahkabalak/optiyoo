# Optiyoo

**Optiyoo**, kullanıcıların hızlıca "Twitter/X benzeri" anketler (poll) oluşturabildiği ve aktif akış (feed) üzerinden anında oy kullanabildiği interaktif bir geri bildirim platformudur.

Oylama deneyimi, sonuçların etkileşimli yüzdelik (`%`) çubuklarıyla anında görselleştirildiği sosyal medya odaklı bir yapıya sahiptir. **OPT puanı (+5 oy başına)** ve profil cüzdanında gösterimi ürün yol haritasında yer alır; mevcut veritabanı ve API katmanında henüz uygulanmamıştır.

## Teknolojik Altyapı (Tech Stack)

### Backend (Golang)
- **Dil / toolchain:** `backend/go.mod` içindeki Go sürümü (şu an 1.26.x).
- **Mimari:** Dışa bağımlılığı (framework) en aza indirilmiş `net/http` tabanlı `ServeMux` (Go 1.22+) routing yapısı.
- **Veritabanı:** PostgreSQL (`lib/pq` Go sürücüsü ile).
- **Öne çıkan özellikler:** Güvenli kriptografik rastgele ID (`crypto/rand`), anket sonuçları için `vote_count` alt sorguları ile API JSON çıktısı.

### Frontend (Vue.js)
- **Çatı:** Vue 3 (Composition API, `<script setup lang="ts">`) ve Vite.
- **Durum yönetimi:** Pinia.
- **Yönlendirme:** Vue Router (`/auth`, `/` feed, `/s/:id` paylaşılabilir anket sayfası).
- **Tasarım:** Vanilla CSS (Tailwind veya benzeri yok). Renk paleti Sanzo Wada renk kombinasyonları (#109) temel alınır. Temalar `GET /api/config` ile sunulur ve `App.vue` içinde `body` sınıfı olarak uygulanır.

## Proje yapısı
Monorepo:
- **`/backend`:** HTTP handler’lar, veritabanı bağlantısı ve şema (`db/db.go`).
- **`/frontend`:** Dashboard (feed + anında oylama), `CreateSurveyModal.vue` ile anket oluşturma, `SurveyCard` / `SurveyView` bileşenleri.
- **`/scripts`:** İsteğe bağlı toplu test verisi (`seed_test_data.sh`).

## Geliştirme ortamını kurma ve çalıştırma

Gereksinimler:
- Go (`backend/go.mod` ile uyumlu sürüm)
- Node.js ve npm (18+ önerilir)
- Docker ve Docker Compose (PostgreSQL için)

### Adımlar

**1. Veritabanını başlatın** (proje kökünde):
```bash
docker-compose up -d
```

**2. Backend’i çalıştırın:**
```bash
cd backend
go run .
```
Sunucu `http://localhost:8080` üzerinde dinler; tablolar (`users`, `surveys`, `questions`, `options`, `answers`) uygulama açılışında oluşturulur/doğrulanır.

**3. Frontend’i çalıştırın:**
```bash
cd frontend
npm install
npm run dev
```
Vite yerel adresi (ör. `http://localhost:5173`) terminalde gösterilir.

## Temel kurallar ve mimari notlar
- **Tek soru kuralı:** Her anket tam olarak bir soru içerebilir; aksi oluşturma isteğinde reddedilir.
- **Açık uçlu sorular:** Yapılandırmada kapalıyken hem API oluşturmayı reddeder hem de istemci `/api/config` ile metin sorusunu kilitleyebilir.
- **Çift oy:** Aynı kullanıcı aynı ankete ikinci kez cevap gönderemez; API **403** döner. İstemci tarafında `completed_polls_<kullanıcıId>` ve (varsa) `user_answer` ile arayüz tutarlı tutulur.

---
*Ajan ve geliştirici bağlamı için `.agents/skills/optiyoo-context/SKILL.md` ve `.cursorrules` dosyaları referans alınmalıdır.*
