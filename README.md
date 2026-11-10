# Whisper

Whisper adalah platform pesan anonim dan percakapan privat real-time. Repository ini menggunakan monorepo dengan frontend Next.js, backend Go, dan PostgreSQL.

## Prasyarat

- Node.js dan npm
- Go
- Docker Engine dan Docker Compose
- PostgreSQL client (`psql`)

## Menjalankan layanan lokal

1. Siapkan konfigurasi lokal:

   ```bash
   cp .env.example .env
   ```

   Nilai di `.env` hanya untuk development lokal. Jangan gunakan nilai tersebut di production.

2. Jalankan PostgreSQL:

   ```bash
   docker compose up -d db
   docker compose ps
   ```

   PostgreSQL hanya dipublikasikan pada `127.0.0.1:5432`. Compose menunggu health check `pg_isready` sebelum menandai database sehat.

3. Jalankan backend di terminal terpisah:

   ```bash
   cd backend
   set -a
   source ../.env
   set +a
   go run ./cmd/api
   ```

   Saat aplikasi mulai, backend membuka pool pgx dan menerapkan migration yang belum dijalankan. Health endpoint: <http://localhost:8080/api/v1/health>

4. Buat admin pertama dari terminal backend yang lain. Perintah ini meminta username dan password secara interaktif; karakter password tidak ditampilkan:

   ```bash
   cd backend
   set -a
   source ../.env
   set +a
   go run ./cmd/create-admin
   ```

   Gunakan password minimal 12 karakter. Tidak ada endpoint pendaftaran admin publik.

5. Jalankan frontend di terminal lain:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   Buka <http://localhost:3000>.

## Memeriksa PostgreSQL

Dengan nilai development bawaan di `.env.example`:

```bash
psql -h 127.0.0.1 -p 5432 -U whisper -d whisper -c 'SELECT 1;'
```

Saat diminta, masukkan nilai `POSTGRES_PASSWORD` dari `.env`. Perintah ini harus mengembalikan satu baris dengan nilai `1`.

## Menghentikan PostgreSQL

```bash
docker compose down
```

Data database disimpan di volume Docker `postgres_data` dan tetap ada setelah container dihentikan. Untuk menghapus database lokal beserta datanya, gunakan `docker compose down -v` setelah memastikan data tersebut memang boleh dihapus.

## Status implementasi

Phase 1–2 menyediakan scaffold Next.js, backend Gin dengan pool `pgx`, PostgreSQL melalui Compose, health check database, dan migration awal. Phase 3 menyediakan login admin berbasis bcrypt dengan sesi cookie `HttpOnly`, logout, endpoint identitas admin, bootstrap admin lokal, serta pembuatan percakapan anonim dengan token acak yang hanya disimpan dalam bentuk hash. Token pengunjung dikirim satu kali saat percakapan dibuat dan diletakkan di fragment link (`#k=...`), lalu dipakai sebagai Bearer token untuk membaca percakapan.

Endpoint Phase 3:

- `POST /api/v1/admin/auth/login`
- `POST /api/v1/admin/auth/logout`
- `GET /api/v1/admin/me`
- `POST /api/v1/public/{slug}/conversations`
- `GET /api/v1/conversations/{id}`

Phase 4 menambahkan pengiriman dan riwayat pesan. Pengunjung memakai `Authorization: Bearer <access_token>`; admin memakai cookie sesi dari login. Riwayat dikembalikan urut lama ke baru, dengan `limit` (maksimum 100) dan `before` memakai ID pesan tertua halaman sebagai cursor. Kirim `client_message_id` UUID yang sama saat retry agar pesan tidak tergandakan.

- `POST /api/v1/conversations/{id}/messages` (pengunjung)
- `GET /api/v1/conversations/{id}/messages?limit=50&before={message_id}` (pengunjung)
- `POST /api/v1/conversations/{id}/messages/read` (pengunjung menandai balasan admin terbaca)
- `POST /api/v1/admin/conversations/{id}/messages` (admin membalas)
- `GET /api/v1/admin/conversations/{id}/messages?limit=50&before={message_id}` (admin)
- `POST /api/v1/admin/conversations/{id}/messages/read` (admin menandai pesan pengunjung terbaca)

Phase 5 menyediakan WebSocket di `GET /api/v1/ws?conversation_id={id}`. Setelah upgrade, kirim event `authenticate`; pengunjung menyertakan token di payload JSON (tidak di URL), sedangkan admin memakai cookie sesi. Event yang didukung: `message.send`, `message.read`, `typing.start`, `typing.stop`, dan `ping`. Server mengirim `auth.success`, `message.ack`, `message.new`, `message.read`, `typing.*`, `presence.update`, `message.sync`, `pong`, dan `error`. Pesan tetap disimpan lewat service database sebelum dibroadcast. Setelah autentikasi, server mengirim 100 pesan terbaru untuk sinkronisasi awal atau reconnect; REST API tetap menyediakan pagination untuk mengambil riwayat lebih lanjut.

Untuk deployment HTTPS, set `SESSION_COOKIE_SECURE=true` dan sesuaikan `FRONTEND_ORIGIN`. Cookie sesi berlaku di `/api/v1` agar dapat dipakai pada API admin dan handshake WebSocket.

## Menjalankan pemeriksaan

Tes unit backend dan build dapat dijalankan dari folder `backend` dengan `go test ./...` dan `go build ./...`. Untuk tes idempotency yang memakai PostgreSQL lokal, muat `.env` lalu jalankan:

```bash
cd backend
set -a
source ../.env
set +a
WHISPER_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/repository -run TestMessageCreateIsIdempotent -count=1
```

Tes integrasi membuat admin, percakapan, dan pesan uji dengan nama acak, lalu menghapus admin uji setelah selesai.
