# Optiyoo

**Optiyoo**, kullanıcıların hızlıca "Twitter/X benzeri" anketler (poll) oluşturabildiği ve aktif akış (feed) üzerinden anında oy kullanabildiği interaktif bir geri bildirim platformudur. 

Kullanıcılar katıldıkları anketler üzerinden **OPT Puanı** (+5 OPT) kazanır ve bu puanlar profillerindeki cüzdanda anlık olarak listelenir. Oylama deneyimi, sonuçların ektileşimli yüzdelik (`%`) çubuklarıyla anında görselleştirildiği sosyal medya odaklı bir yapıya sahiptir.

## 🚀 Teknolojik Altyapı (Tech Stack)

### Backend (Golang)
- **Dil:** Go 1.26+ 
- **Mimari:** Dışa bağımlılığı (framework) en aza indirilmiş `net/http` tabanlı güncel `ServeMux` (Go 1.22+) routing yapısı.
- **Veritabanı:** PostgreSQL (`lib/pq` Go sürücüsü ile)
- **Öne Çıkan Özellikler:** Güvenli Cripto Random ID (`crypto/rand`) kullanımı, kompleks Vote aggregation için dinamik SQL join'leri.

### Frontend (Vue.js)
- **Çatı:** Vue 3 (Composition API, `<script setup lang="ts">`) & Vite
- **Durum Yönetimi (State):** Pinia
- **Yönlendirme:** Vue Router
- **Tasarım:** Sadece Vanilla CSS (Tailwind veya başka kütüphane kullanılmamıştır). Renk paleti Sanzo Wada Renk Kombinasyonları (#109) temel alınarak oluşturulmuş özgün, cam/modern ögeler barındıran bir stildir.

## 📁 Proje Yapısı
Monorepo mantığıyla tasarlanmıştır:
- `/backend`: Uygulamanın sunucu tarafını tutar. İş kuralları, HTTP Handlers, veritabanı şemaları burada bulunur.
- `/frontend`: Uygulamanın istemci (client) tarafını oluşturur. Anında oylama akışı (Dashboard), anket yaratma sayfası gibi etkileşimli ekranlar bu klasördedir.

## ⚡ Geliştirme Ortamını Kurma ve Çalıştırma

Projeyi yerel bilgisayarınızda (local) çalıştırmak için aşağıdaki araçların yüklü olması gerekir:
- Go (1.26 önerilir)
- Node.js & npm (18+ önerilir)
- Docker & Docker Compose (PostgreSQL veritabanı için)

### Adım Adım Kurulum

**1. Veritabanını Ayağa Kaldırın:**
Projenin kök dizininde veya `/backend` içinde bulunan docker-compose dosyası aracılığıyla PostgreSQL'i başlatın.
```bash
docker-compose up -d
```

**2. Backend'i Başlatın:**
```bash
cd backend
go run .
```
Backend `http://localhost:8080` adresinde çalışmaya başlayacak ve veritabanı bağlantısı ile tabloları (users, surveys, questions, options, answers) otomatik olarak yapılandıracaktır.

**3. Frontend'i Başlatın:**
```bash
cd frontend
npm install
npm run dev
```
Frontend Vue CLI/Vite aracılığıyla ayağa kalkacak ve ekranda size yerel çalışma adresini (`http://localhost:5173` vb.) verecektir.

## 🔐 Temel Kurallar ve Mimari Notlar
- **Tek Soru Kuralı (Single Question Rule):** Sistem genelinde her bir anket, hızlı oylama yapılabilmesi adına sadece (1) tek bir soru barındırabilir.
- **Açık Uçlu Soruların Kapatılması:** Optiyoo prototip (MVP) aşamasında ağırlıklı olarak "Çoktan Seçmeli (Single Choice)" yapıyı ön plana çıkarır. Açık uçlu sorular konfigürasyon tarafından şimdilik kısıtlanmıştır.
- **Oylama ve Doğrulama:** Backend, kullanıcıların bir ankete sadece bir defa katılabilmesini SQL tabanlı güvence altına alırken; Frontend tarafında da lokal takibi (`completed_polls`) sürdürerek UI performansını artırır.

---
*Bu proje, geliştirici ve sistem ajanlarının tam uyumla ilerlemesi adına .agents ve Cursor kuralları (cursorrules) referans alınarak ölçeklendirilmektedir.*
