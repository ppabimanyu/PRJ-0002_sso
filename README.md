# SSO Service

Identity provider untuk semua project pribadi: satu akun untuk semua aplikasi, baik user internal maupun publik. Service ini berbicara OIDC/OAuth 2.0 standar sehingga Application tidak terkunci ke implementasi ini.

Fondasi runtime berasal dari `PRJ-0001_boilerplate`. Kontrak desain ada di [docs/architecture.md](docs/architecture.md), [docs/spec.md](docs/spec.md), [GLOSSARY.md](GLOSSARY.md), dan [docs/adr/](docs/adr/).

## Stack

| Bagian | Teknologi |
| --- | --- |
| Backend | Go, chi, pgx, sqlc, goose |
| Database | PostgreSQL |
| Cache dan rate limit | Redis |
| Email development | Mailpit |
| Observability | OpenTelemetry, OTLP, slog |
| Frontend | React 19, Vite, TypeScript |
| UI | Tailwind CSS v4, shadcn/ui, lucide-react |
| Routing dan data | React Router v7, TanStack Query v5, openapi-fetch |
| Testing | testify, testcontainers, Vitest, React Testing Library, MSW |

## Struktur repo

```text
backend/
  cmd/api/                 Entry point
  internal/config/         Environment ke struct, validasi fail-fast
  internal/domain/         Entity, konstanta, dan port
  internal/httphandler/    Router dan handler
  internal/infra/          Database, Redis, logger, OpenTelemetry
  internal/middleware/     Request ID, access log, recovery, CORS, rate limit
  internal/pkg/            problem, slogx, validate, pagination
  internal/repository/     Implementasi port dan kode sqlc
  db/migrations/           Migrasi goose
  db/queries/              Query sqlc
  docs/                    OpenAPI hasil generate
frontend/
  src/app/                 Router, providers, layout, error boundary
  src/components/ui/       Komponen shadcn/ui
  src/components/shared/   Komponen lintas fitur
  src/features/            Fitur per project
  src/hooks/               Hook bersama
  src/lib/                 API client, config, util
  src/types/               Tipe hasil generate
deploy/                    Compose, nginx, konfigurasi observability
docs/architecture.md       Arsitektur lengkap
```

## Prasyarat

- Go 1.27.4
- Node.js 22 atau lebih baru
- Docker dan Docker Compose
- air, sqlc, swag, dan golangci-lint untuk development lokal

Pasang `air` dengan `go install github.com/air-verse/air@latest`. Pasang sqlc, swag, dan golangci-lint sesuai dokumentasi masing-masing.

Goose tidak perlu dipasang lokal jika memakai Docker Compose. Perintah `make migrate-up` menjalankan goose di dalam image backend.

## Mulai cepat

1. Salin konfigurasi.

```bash
cp .env.example .env
```

2. Jalankan stack.

```bash
make dev
```

3. Buka http://localhost:3000.

Compose menjalankan PostgreSQL, Redis, Mailpit, migrasi goose, backend, dan frontend. Frontend di port 3000 mem-proxy `/api`, `/healthz`, dan `/readyz` ke backend.

Endpoint development:

- `GET http://localhost:8080/healthz` untuk liveness.
- `GET http://localhost:8080/readyz` untuk readiness, termasuk cek PostgreSQL dan Redis.
- http://localhost:8025 untuk Mailpit.

## Perintah

| Perintah | Fungsi |
| --- | --- |
| `make dev` | Menjalankan seluruh stack dengan Docker Compose |
| `make down` | Menghentikan stack |
| `make logs` | Mengikuti log stack |
| `make observability` | Menjalankan stack dengan Jaeger dan Prometheus |
| `make migrate-up` | Menjalankan migrasi goose |
| `make migrate-down` | Membatalkan satu migrasi |
| `make backend-dev` | Menjalankan backend dengan air |
| `make frontend-dev` | Menjalankan Vite dev server |
| `make generate` | Generate sqlc, OpenAPI, dan tipe frontend |
| `make test` | Unit test backend dan frontend |
| `make test-integration` | Integration test backend dengan testcontainers |
| `make lint` | Lint backend dan frontend |
| `make build` | Build binary backend dan bundle frontend |
| `make smoke` | Cek stack yang sedang berjalan end to end |

## Development lokal

Backend dengan hot reload:

```bash
docker compose -f deploy/docker-compose.yml --profile dev up -d db redis mailpit
make migrate-up
make backend-dev
```

Frontend dengan Vite:

```bash
cd frontend && npm install
cd ..
make frontend-dev
```

Vite berjalan di http://localhost:5173 dan mem-proxy `/api`, `/healthz`, dan `/readyz` ke http://localhost:8080.

## Generate kode

`make generate` menjalankan tiga langkah:

1. `sqlc generate` dari query di `backend/db/queries`.
2. `swag init` dari anotasi Go ke `backend/docs/swagger.json`.
3. Konversi Swagger ke OpenAPI 3 dan generate `frontend/src/types/api.ts`.

CI gagal jika hasil generate berbeda dari file yang di-commit. Jalankan `make generate` setiap kali mengubah query, migrasi, atau anotasi handler.

## Menambah fitur

1. Tambah migrasi goose di `backend/db/migrations`.
2. Tambah query di `backend/db/queries` lalu jalankan `make generate`.
3. Tambah entity dan port di `backend/internal/domain`.
4. Implementasikan port di `backend/internal/repository`.
5. Tambah handler di `backend/internal/httphandler` dan mount di `/api/v1`.
6. Tambah fitur frontend di `frontend/src/features/<nama>`.
7. Jalankan `make generate` untuk memperbarui tipe dan client.

Handler logic hanya mengimpor `domain`, `pkg`, dan modul `auth`; router wiring juga mengimpor `middleware`. Akses data lewat port di `domain`, dan implementasinya ada di `repository`. Modul auth domain adalah pengecualian yang disengaja; lihat ADR 0007.

## Observability

Jalankan stack dengan profil observability.

```bash
make observability
```

- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090
- OTLP HTTP: http://localhost:4318
- Metrics collector: http://localhost:8889/metrics

Backend mengirim trace dan metrics ke OTel Collector. Collector meneruskan trace ke Jaeger dan mengekspos metrics untuk Prometheus. Log JSON menyertakan `trace_id` dan `span_id`. Response menyertakan `X-Trace-Id` dan `X-Request-Id`.

## Testing

```bash
make test
make test-integration
```

`make test` menjalankan unit test backend dan frontend. `make test-integration` menjalankan test repository dengan PostgreSQL dari testcontainers dan membutuhkan Docker.

## Konfigurasi

Semua variabel ada di `.env.example`. Docker Compose dan Vite membaca `.env` yang sama dari root repo. Variabel utama:

| Variabel | Default | Fungsi |
| --- | --- | --- |
| `APP_ENV` | `development` | Nama environment |
| `PORT` | `8080` | Port HTTP backend |
| `DATABASE_URL` | wajib | Koneksi PostgreSQL |
| `REDIS_URL` | wajib | Koneksi Redis |
| `CORS_ALLOWED_ORIGINS` | kosong | Origin browser yang diizinkan, dipisah koma |
| `TRUSTED_PROXY_CIDRS` | kosong | CIDR reverse proxy yang boleh mengirim `X-Forwarded-For` |
| `RATE_LIMIT_REQUESTS` | `60` | Jumlah request per window |
| `RATE_LIMIT_WINDOW` | `1m` | Panjang window rate limit |
| `LOG_LEVEL` | `info` | Level log |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | kosong | Endpoint OTLP, kosong berarti export dimatikan |
| `OTEL_SERVICE_NAME` | `sso-api` | Nama service di trace |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | Rasio sampling trace |
| `VITE_API_URL` | `/api/v1` | Base URL API untuk frontend |

Rate limit memakai Redis dan fail-open. Jika Redis tidak tersedia, request tetap dilayani dan error dicatat di log.

## Deployment

Backend dan frontend punya image sendiri. Backend memakai binary statis di distroless. Frontend memakai nginx dengan SPA fallback dan proxy `/api` ke backend. Untuk production, jalankan migrasi sebelum backend baru menerima traffic, dan letakkan reverse proxy TLS di depan kedua service. `deploy/nginx.conf` adalah contoh reverse proxy same-origin.

## Dokumentasi

- [docs/architecture.md](docs/architecture.md)
- [docs/spec.md](docs/spec.md)
- [GLOSSARY.md](GLOSSARY.md)
- [docs/adr/](docs/adr/)

## Lisensi

MIT. Lihat [LICENSE](LICENSE).
