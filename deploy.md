# Panduan Deployment Lokal - Nara PRD Generator App

Dokumen ini memuat panduan lengkap untuk melakukan *deployment* dan menjalankan aplikasi **Nara PRD Generator** di komputer lokal (*localhost*).

Tersedia dua metode yang dapat dipilih:
1. **Metode 1: Menggunakan Docker Compose (Paling Mudah, Cepat & Otomatis)** — *Rekomendasi untuk demonstrasi, pengujian menyeluruh, atau menjalankan full stack tanpa instalasi runtime manual.*
2. **Metode 2: Menjalankan Secara Standalone / Manual (Live Development)** — *Cocok untuk pengembangan aktif yang membutuhkan fitur hot-reload pada backend atau frontend.*

---

## 1. Arsitektur Port & Layanan

Secara default, aplikasi berjalan pada port-port berikut:

| Layanan | Teknologi | Port | URL Akses |
| :--- | :--- | :---: | :--- |
| **Frontend Web** | Next.js 14, React 18, Tailwind CSS | `3000` | [http://localhost:3000](http://localhost:3000) |
| **Backend API** | Go 1.25, Gin Web Framework | `8080` | [http://localhost:8080](http://localhost:8080) |
| **Database** | PostgreSQL 16 | `5432` | `localhost:5432` |

---

## 2. Metode 1: Menggunakan Docker Compose (Rekomendasi)

### Prasyarat:
- Pastikan **Docker Desktop** (atau Docker Engine + Docker Compose) sudah terpasang dan aplikasinya sedang berjalan (*running*).

### Langkah-langkah:

#### Langkah 1: Siapkan File Environment `.env`
Buka terminal di root direktori proyek (`PRD-Generator-App`), lalu salin template konfigurasi:

**Windows (PowerShell):**
```powershell
Copy-Item .env.example .env
```

**Linux / macOS:**
```bash
cp .env.example .env
```

Buka file `.env` yang baru dibuat dan sesuaikan nilainya (kunci minimal 32 karakter):
```env
JWT_SECRET=rahasia-kunci-jwt-minimal-32-karakter-acak-123456
ENCRYPTION_KEY=rahasia-kunci-enkripsi-aes-minimal-32-karakter
FRONTEND_ORIGIN=http://localhost:3000
PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_URL=http://localhost:8080
AI_BASE_URL=https://api.openai.com/v1
AI_API_KEY=   # Opsional: dapat dikosongkan dan diisi nanti lewat UI Settings
AI_MODEL=gpt-4o-mini
TRUSTED_PROXIES=
```
> **Catatan Kunci AI:** Jika `AI_API_KEY` dikosongkan di sini, server tetap berjalan normal. Anda dapat mengonfigurasi dan menyimpan API key (OpenAI, Gemini, Groq, DeepSeek, Claude) atau mengarahkan ke LLM lokal (Ollama) secara langsung melalui menu **Settings** di antarmuka web.

#### Langkah 2: Bangun & Jalankan Seluruh Stack
Jalankan perintah berikut di terminal:
```bash
docker compose up --build
```

**Apa yang terjadi secara otomatis:**
1. Container `postgres` menyala dan diverifikasi kesehatannya (*healthcheck*).
2. Container `migrate` mengeksekusi migrasi DDL database (`000001` sampai `000006`) ke PostgreSQL.
3. Container `api` mengompilasi binary Go dan menyalakan REST API di port `8080`.
4. Container `frontend` membangun production bundle Next.js mandiri (*standalone*) dan melayani web di port `3000`.

#### Langkah 3: Akses Aplikasi
Buka browser Anda dan navigasi ke:
👉 **[http://localhost:3000](http://localhost:3000)**

Health check API dapat diperiksa di:
👉 **[http://localhost:8080/readyz](http://localhost:8080/readyz)**

#### Perintah Pengelolaan Docker:
- Menjalankan di latar belakang (*detached mode*):
  ```bash
  docker compose up -d --build
  ```
- Melihat log output:
  ```bash
  docker compose logs -f
  ```
- Menghentikan aplikasi:
  ```bash
  docker compose down
  ```
- Menghentikan aplikasi dan mereset total database (*clean state*):
  ```bash
  docker compose down -v
  ```

---

## 3. Metode 2: Menjalankan Secara Standalone (Development Mode)

Gunakan metode ini jika Anda ingin melakukan koding (*development*) dengan fitur *live-reloading*.

### Prasyarat:
- **Go** (versi $\ge 1.25$)
- **Node.js** (versi $\ge 20$) & **npm**
- **PostgreSQL 16**

---

### Langkah 2.1: Menyiapkan Database PostgreSQL

Jika Anda belum memiliki PostgreSQL lokal, Anda dapat menyalakan instance PostgreSQL ringan melalui Docker:
```powershell
docker run -d --name pg-nara -p 5432:5432 -e POSTGRES_USER=prd -e POSTGRES_PASSWORD=prd -e POSTGRES_DB=prd_generator postgres:16-alpine
```

---

### Langkah 2.2: Eksekusi Migrasi Database

Jalankan seluruh skrip SQL migrasi yang berada di folder `backend/migrations/` ke database `prd_generator`.

Jika menggunakan tool CLI `golang-migrate`:
```powershell
migrate -path backend/migrations -database "postgres://prd:prd@localhost:5432/prd_generator?sslmode=disable" up
```

*Alternatif tanpa CLI migrate:* Anda dapat mengeksekusi isi file `.up.sql` secara berurutan (`000001` s/d `000006`) menggunakan GUI database pilihan Anda (DBeaver, TablePlus, pgAdmin) atau melalui `psql`.

---

### Langkah 2.3: Menjalankan Backend Go API

1. Buka terminal baru dan masuk ke direktori `backend`:
   ```powershell
   cd backend
   ```
2. Set variabel lingkungan:

   **PowerShell (Windows):**
   ```powershell
   $env:APP_ENV="development"
   $env:HTTP_ADDR=":8080"
   $env:DATABASE_URL="postgres://prd:prd@localhost:5432/prd_generator?sslmode=disable"
   $env:JWT_SECRET="rahasia-kunci-jwt-minimal-32-karakter-acak-123456"
   $env:ENCRYPTION_KEY="rahasia-kunci-enkripsi-aes-minimal-32-karakter"
   $env:FRONTEND_ORIGIN="http://localhost:3000"
   $env:PUBLIC_API_URL="http://localhost:8080"
   $env:COOKIE_SECURE="false"
   $env:AI_BASE_URL="https://api.openai.com/v1"
   $env:AI_MODEL="gpt-4o-mini"
   ```

   **Bash (Linux / macOS):**
   ```bash
   export APP_ENV=development
   export HTTP_ADDR=:8080
   export DATABASE_URL="postgres://prd:prd@localhost:5432/prd_generator?sslmode=disable"
   export JWT_SECRET="rahasia-kunci-jwt-minimal-32-karakter-acak-123456"
   export ENCRYPTION_KEY="rahasia-kunci-enkripsi-aes-minimal-32-karakter"
   export FRONTEND_ORIGIN="http://localhost:3000"
   export PUBLIC_API_URL="http://localhost:8080"
   export COOKIE_SECURE="false"
   export AI_BASE_URL="https://api.openai.com/v1"
   export AI_MODEL="gpt-4o-mini"
   ```

3. Unduh modul dependensi & jalankan backend:
   ```bash
   go mod download
   go run ./cmd/api
   ```
   *Backend siap dan mendengarkan request di `http://localhost:8080`.*

---

### Langkah 2.4: Menjalankan Frontend Next.js

1. Buka terminal baru dan masuk ke direktori `frontend`:
   ```powershell
   cd frontend
   ```
2. Instal dependensi Node.js:
   ```bash
   npm install
   ```
3. Buat file `.env.local` di dalam folder `frontend`:
   ```env
   NEXT_PUBLIC_API_URL=http://localhost:8080
   ```
4. Jalankan server development:
   ```bash
   npm run dev
   ```
   *Frontend siap di `http://localhost:3000` dengan dukungan hot-reloading.*

---

## 4. Langkah Pertama Menggunakan Aplikasi

1. **Registrasi Akun Baru**:
   - Buka browser di [http://localhost:3000/register](http://localhost:3000/register).
   - Masukkan Nama, Email, dan Password (minimal 10 karakter).
   - Klik **Daftar** untuk langsung diarahkan ke Dashboard.

2. **Pengaturan AI Provider**:
   - Masuk ke menu **Settings** ([http://localhost:3000/settings](http://localhost:3000/settings)).
   - Di bagian **Konfigurasi AI Generator**, pilih salah satu provider:
     - **OpenAI**: Masukkan API key OpenAI Anda (`sk-...`).
     - **Google Gemini**: Masukkan Gemini API key Anda (`AIzaSy...`).
     - **Groq Cloud**: Masukkan Groq API key (`gsk_...`) untuk kecepatan inferensi tertinggi.
     - **DeepSeek**: Masukkan API key DeepSeek (`sk-...`).
     - **Anthropic Claude**: Masukkan Anthropic API key (`sk-ant-...`).
     - **Ollama (Lokal)**: Tanpa API key! Pastikan Ollama berjalan di komputer Anda pada `http://localhost:11434/v1` dengan model seperti `llama3.2` atau `deepseek-r1:8b`.
   - Klik **Uji Koneksi** untuk memverifikasi handshake ke provider AI.
   - Klik **Simpan Pengaturan**.

3. **Membuat Dokumen PRD Pertama**:
   - Klik menu **Projects** $\rightarrow$ **Buat Proyek Baru**.
   - Masukkan ide produk Anda (contoh: *Platform SaaS manajemen inventaris warung UMKM berbasis mobile*).
   - Jawab pertanyaan klarifikasi cerdas dari AI.
   - AI akan menyusun PRD lengkap secara *streaming real-time* dengan visualisasi diagram alur arsitektur Mermaid otomatis.

---

## 5. Pemecahan Masalah (Troubleshooting)

### A. Port Bentrok (Port Already in Use)
- **Gejala:** Error `listen tcp :3000: bind: address already in use` atau `:8080` atau `:5432`.
- **Solusi:**
  - Cari proses yang menggunakan port tersebut:
    ```powershell
    # Windows
    Get-NetTCPConnection -LocalPort 3000, 8080, 5432 | Select-Object LocalPort, OwningProcess
    ```
  - Hentikan proses tersebut atau ubah pemetaan port host di file `docker-compose.yml` (misal `"3001:3000"`).

### B. Status Sesi Login Hilang / Gagal Login
- **Gejala:** Muncul error `unauthorized` atau loop kembali ke halaman login.
- **Solusi:** Pastikan `FRONTEND_ORIGIN` di backend bernilai sama persis dengan URL yang Anda buka di browser (`http://localhost:3000`). Backend menggunakan cookie dengan atribut `SameSite: Lax` yang terikat pada origin frontend.

### C. Gagal Koneksi ke AI Provider
- **Gejala:** Muncul pesan error koneksi pada tombol Uji Koneksi di Settings.
- **Solusi:**
  - Pastikan koneksi internet aktif jika menggunakan OpenAI, Gemini, Groq, atau DeepSeek.
  - Jika menggunakan Ollama di dalam Docker, ubah base URL dari `http://localhost:11434/v1` menjadi `http://host.docker.internal:11434/v1` agar container dapat menjangkau host machine.

### D. Menjalankan Unit Test Mandiri
Untuk memastikan seluruh modul backend berfungsi normal di komputer Anda:
```powershell
cd backend
go test -v ./...
```
*(Seluruh pengujian `ai`, `postgres`, `security`, dan `service` harus berstatus `PASS`).*
