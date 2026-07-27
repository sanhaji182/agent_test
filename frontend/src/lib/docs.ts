// Bilingual documentation content
export type Lang = "en" | "id";

export interface DocPage {
  slug: string;
  title: { en: string; id: string };
  content: { en: string; id: string };
  category: string;
}

export const categories = [
  { key: "start", label: { en: "Getting Started", id: "Memulai" } },
  { key: "features", label: { en: "Features", id: "Fitur" } },
  { key: "ops", label: { en: "Operations", id: "Operasional" } },
];

export const docs: DocPage[] = [
  {
    slug: "introduction",
    category: "start",
    title: { en: "Introduction", id: "Pengantar" },
    content: {
      en: `# Introduction

GoTest Agent is a self-hosted AI testing platform that reads your codebase, generates tests, executes them in a sandboxed browser, and auto-fixes failures.

## What it does
1. **Analyzes** your project — detects language, framework, routes, and endpoints.
2. **Generates** a test plan with prioritized scenarios.
3. **Writes** Playwright test files automatically.
4. **Executes** tests in Steel Browser with video recording.
5. **Auto-fixes** failing tests (up to 3 attempts).
6. **Reports** results with screenshots, video replay, and live streaming.

## Who is it for?
- Developers who want automated testing without manual setup.
- QA teams who need continuous monitoring and failure alerts.
- Tech leads who want release confidence scores.
- Teams who prefer self-hosted over SaaS testing tools.

## Key differentiators
- **AI-powered**: generates tests from your code, not templates.
- **Self-hosted**: your data stays on your infrastructure.
- **Video replay**: watch exactly what happened during test execution.
- **Risk intelligence**: prioritizes tests by failure risk.
- **Real-time**: live execution streaming to the control room.`,
      id: `# Pengantar

GoTest Agent adalah platform testing AI self-hosted yang membaca kode Anda, menghasilkan test, menjalankannya di browser sandbox, dan memperbaiki kegagalan secara otomatis.

## Apa yang dilakukan
1. **Menganalisis** project Anda — mendeteksi bahasa, framework, route, dan endpoint.
2. **Membuat** rencana test dengan skenario yang diprioritaskan.
3. **Menulis** file test Playwright secara otomatis.
4. **Menjalankan** test di Steel Browser dengan rekaman video.
5. **Memperbaiki** test yang gagal secara otomatis (maks 3 percobaan).
6. **Melaporkan** hasil dengan screenshot, video replay, dan streaming langsung.

## Untuk siapa?
- Developer yang ingin testing otomatis tanpa setup manual.
- Tim QA yang butuh monitoring berkelanjutan dan alert kegagalan.
- Tech lead yang ingin skor kepercayaan release.
- Tim yang lebih memilih self-hosted daripada SaaS testing.

## Keunggulan utama
- **AI-powered**: menghasilkan test dari kode Anda, bukan template.
- **Self-hosted**: data Anda tetap di infrastruktur Anda.
- **Video replay**: tonton persis apa yang terjadi saat eksekusi test.
- **Risk intelligence**: memprioritaskan test berdasarkan risiko kegagalan.
- **Real-time**: streaming eksekusi langsung ke control room.`,
    },
  },
  {
    slug: "getting-started",
    category: "start",
    title: { en: "Getting Started", id: "Memulai" },
    content: {
      en: `# Getting Started

## Quick Start (Docker)

\`\`\`bash
git clone https://github.com/sanhaji182/agent_test.git
cd agent_test
cp .env.example .env
# Add your ANTHROPIC_API_KEY to .env
make up
\`\`\`

Open **http://localhost:3001** for the dashboard.

## First Steps
1. Click **"Seed Demo Data"** on the Overview page to see sample runs.
2. Browse the **Runs** page to inspect test executions.
3. Open a run to see the execution console with video, steps, and failures.
4. Visit **Monitoring** to create a recurring schedule.
5. Check **Risk** for intelligent test recommendations.

## Connect Your Project

\`\`\`bash
curl -X POST http://localhost:8080/api/v1/runs \\
  -H "Content-Type: application/json" \\
  -d '{"project_path": "/path/to/project", "requirements": "test login flow"}'
\`\`\`

Or use the MCP tool in Cursor/VS Code: \`run_tests(project_path="/path/to/project")\`

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| ANTHROPIC_API_KEY | Yes | AI model access |
| API_KEY | No | API authentication (empty = no auth) |
| DATABASE_URL | No | PostgreSQL (falls back to in-memory) |
| STEEL_API_URL | No | Steel Browser for test execution |`,
      id: `# Memulai

## Quick Start (Docker)

\`\`\`bash
git clone https://github.com/sanhaji182/agent_test.git
cd agent_test
cp .env.example .env
# Tambahkan ANTHROPIC_API_KEY ke .env
make up
\`\`\`

Buka **http://localhost:3001** untuk dashboard.

## Langkah Pertama
1. Klik **"Seed Demo Data"** di halaman Overview untuk melihat contoh run.
2. Buka halaman **Runs** untuk memeriksa eksekusi test.
3. Buka sebuah run untuk melihat console eksekusi dengan video, langkah, dan kegagalan.
4. Kunjungi **Monitoring** untuk membuat jadwal berulang.
5. Cek **Risk** untuk rekomendasi test yang cerdas.

## Hubungkan Project Anda

\`\`\`bash
curl -X POST http://localhost:8080/api/v1/runs \\
  -H "Content-Type: application/json" \\
  -d '{"project_path": "/path/to/project", "requirements": "test login flow"}'
\`\`\`

Atau gunakan MCP tool di Cursor/VS Code: \`run_tests(project_path="/path/to/project")\`

## Environment Variables

| Variable | Wajib | Deskripsi |
|----------|-------|-----------|
| ANTHROPIC_API_KEY | Ya | Akses model AI |
| API_KEY | Tidak | Autentikasi API (kosong = tanpa auth) |
| DATABASE_URL | Tidak | PostgreSQL (fallback ke in-memory) |
| STEEL_API_URL | Tidak | Steel Browser untuk eksekusi test |`,
    },
  },
  {
    slug: "concepts",
    category: "start",
    title: { en: "Core Concepts", id: "Konsep Utama" },
    content: {
      en: `# Core Concepts

## Run
A test run is one complete execution cycle: analyze → plan → write → execute → fix → report.

## State Machine
Every run goes through these states:
- **idle** → waiting to start
- **analyzing** → reading your codebase
- **plan_generated** → test plan created
- **writing_tests** → generating Playwright files
- **running** → executing tests in browser
- **fixing** → auto-fixing failures (up to 3x)
- **done** → completed successfully
- **failed** → completed with unresolved failures
- **simulated** → step-walk only; no real browser execution ran

## Schedule
A recurring configuration that automatically creates runs at set intervals (daily, weekly, monthly, or cron).

## Risk Score
A 0-100% score indicating how likely a test or schedule is to fail, based on historical patterns.

## Confidence Grade
A letter grade (A-F) for releases, computed from pass rate, risk score, and data freshness.

## Review
A human approval step for AI-generated test plans, scripts, or fix suggestions.

## Suite
A named group of tests organized by tags, project, or environment.`,
      id: `# Konsep Utama

## Run
Sebuah test run adalah satu siklus eksekusi lengkap: analisis → rencana → tulis → jalankan → perbaiki → lapor.

## State Machine
Setiap run melewati state berikut:
- **idle** → menunggu dimulai
- **analyzing** → membaca kode Anda
- **plan_generated** → rencana test dibuat
- **writing_tests** → menghasilkan file Playwright
- **running** → menjalankan test di browser
- **fixing** → memperbaiki kegagalan otomatis (maks 3x)
- **done** → selesai berhasil
- **failed** → selesai dengan kegagalan yang belum terselesaikan
- **simulated** → simulasi langkah saja; tidak ada eksekusi browser nyata yang berjalan

## Schedule
Konfigurasi berulang yang otomatis membuat run pada interval tertentu (harian, mingguan, bulanan, atau cron).

## Risk Score
Skor 0-100% yang menunjukkan seberapa besar kemungkinan test atau schedule gagal, berdasarkan pola historis.

## Confidence Grade
Nilai huruf (A-F) untuk release, dihitung dari pass rate, risk score, dan kesegaran data.

## Review
Langkah persetujuan manusia untuk rencana test, script, atau saran perbaikan yang dihasilkan AI.

## Suite
Grup test bernama yang diorganisir berdasarkan tag, project, atau environment.`,
    },
  },
  {
    slug: "control-room",
    category: "features",
    title: { en: "Control Room", id: "Control Room" },
    content: {
      en: `# Control Room

The Overview page is your real-time test execution control room.

## What it shows
- **Running Now** — active test executions with live progress
- **Failed — Needs Attention** — recent failures with error context
- **Recommended Actions** — AI-suggested next steps
- **Highest Risk** — tests and schedules most likely to fail
- **Trend** — pass rate over time

## Live updates
The control room receives instant updates via Server-Sent Events (SSE). When a test fails or completes, you see it immediately without refreshing.

## Connection status
- 🟢 **Live** — connected to event stream
- ⚪ **Polling** — fallback mode (refreshes every 10s)

## Failure alerts
When a test fails, a toast notification appears automatically showing the test name and error.`,
      id: `# Control Room

Halaman Overview adalah control room eksekusi test real-time Anda.

## Apa yang ditampilkan
- **Running Now** — eksekusi test aktif dengan progress langsung
- **Failed — Needs Attention** — kegagalan terbaru dengan konteks error
- **Recommended Actions** — langkah selanjutnya yang disarankan AI
- **Highest Risk** — test dan schedule yang paling mungkin gagal
- **Trend** — pass rate dari waktu ke waktu

## Update langsung
Control room menerima update instan via Server-Sent Events (SSE). Ketika test gagal atau selesai, Anda langsung melihatnya tanpa refresh.

## Status koneksi
- 🟢 **Live** — terhubung ke event stream
- ⚪ **Polling** — mode fallback (refresh setiap 10 detik)

## Alert kegagalan
Ketika test gagal, notifikasi toast muncul otomatis menampilkan nama test dan error.`,
    },
  },
  {
    slug: "runs",
    category: "features",
    title: { en: "Runs & Audit View", id: "Run & Tampilan Audit" },
    content: {
      en: `# Runs & Audit View

## Run List
Browse all test executions with search, filter by status, and sort by time. Click any run to open the execution console.

## Execution Console
Each run detail page is an audit-style inspection view:
- **Summary bar** — status, timeline, result counts, rerun button
- **Video tab** — browser recording with failure markers and step chips
- **Live Events** — real-time execution log
- **Steps** — test plan breakdown by scenario
- **Files** — generated Playwright test code
- **Recordings** — screenshot metadata
- **Failures** — error details with screenshot links
- **Visual** — baseline vs current comparison

## Actions
- **Rerun** — create a new run with the same configuration
- **Compare** — diff two runs to see what changed
- **Report** — open HTML report in new tab
- **Download** — export run data as JSON`,
      id: `# Run & Tampilan Audit

## Daftar Run
Jelajahi semua eksekusi test dengan pencarian, filter berdasarkan status, dan urutkan berdasarkan waktu. Klik run manapun untuk membuka console eksekusi.

## Console Eksekusi
Setiap halaman detail run adalah tampilan inspeksi bergaya audit:
- **Summary bar** — status, timeline, jumlah hasil, tombol rerun
- **Tab Video** — rekaman browser dengan penanda kegagalan dan chip langkah
- **Live Events** — log eksekusi real-time
- **Steps** — breakdown rencana test per skenario
- **Files** — kode test Playwright yang dihasilkan
- **Recordings** — metadata screenshot
- **Failures** — detail error dengan link screenshot
- **Visual** — perbandingan baseline vs current

## Aksi
- **Rerun** — buat run baru dengan konfigurasi yang sama
- **Compare** — bandingkan dua run untuk melihat perubahan
- **Report** — buka laporan HTML di tab baru
- **Download** — ekspor data run sebagai JSON`,
    },
  },
  {
    slug: "video-replay",
    category: "features",
    title: { en: "Video Replay", id: "Video Replay" },
    content: {
      en: `# Browser Video Replay

Every browser test run is recorded as a video so you can watch exactly what happened.

## How it works
1. Playwright records the browser session (1280×720 WebM).
2. Video is saved and linked to the run.
3. The Video tab shows the player with inspection controls.

## Inspection features
- **Jump to failure** — seeks video to the exact failure moment
- **Step markers** — clickable chips that seek to each test step
- **Timeline scrubber** — visual bar with failure marker (red) and step markers
- **Active step highlight** — current step lights up during playback
- **Download** — save the video file locally

## Precision
- Step timestamps come from Playwright's JSON report (precise when available).
- If precise data is unavailable, steps are distributed approximately and labeled as such.

## When video is unavailable
- Status shows "No video recording" with guidance to run with Steel Browser.
- The run detail still works fully without video.`,
      id: `# Video Replay Browser

Setiap test run browser direkam sebagai video sehingga Anda bisa menonton persis apa yang terjadi.

## Cara kerja
1. Playwright merekam sesi browser (1280×720 WebM).
2. Video disimpan dan dihubungkan ke run.
3. Tab Video menampilkan player dengan kontrol inspeksi.

## Fitur inspeksi
- **Jump to failure** — seek video ke momen kegagalan yang tepat
- **Step markers** — chip yang bisa diklik untuk seek ke setiap langkah test
- **Timeline scrubber** — bar visual dengan penanda kegagalan (merah) dan penanda langkah
- **Active step highlight** — langkah saat ini menyala selama pemutaran
- **Download** — simpan file video secara lokal

## Presisi
- Timestamp langkah berasal dari laporan JSON Playwright (presisi saat tersedia).
- Jika data presisi tidak tersedia, langkah didistribusikan secara perkiraan dan diberi label demikian.

## Ketika video tidak tersedia
- Status menampilkan "No video recording" dengan panduan untuk menjalankan dengan Steel Browser.
- Detail run tetap berfungsi penuh tanpa video.`,
    },
  },
  {
    slug: "monitoring",
    category: "features",
    title: { en: "Monitoring & Schedules", id: "Monitoring & Jadwal" },
    content: {
      en: `# Monitoring & Schedules

## Schedules
Create recurring test runs that execute automatically.

### Frequencies
- **Daily** — runs every 24 hours
- **Weekly** — runs every 7 days
- **Monthly** — runs once per month
- **Cron** — custom cron expression (e.g., \`*/5 * * * *\`)

### Environment targeting
Each schedule can target: local, staging, or production.

### Controls
- **Run Now** — trigger immediately
- **Pause/Resume** — toggle without deleting
- **Webhook** — get notified on failure via Slack/Telegram

## Background scheduler
The server runs a background goroutine that checks for due schedules every 60 seconds and creates runs automatically.`,
      id: `# Monitoring & Jadwal

## Jadwal
Buat test run berulang yang dijalankan secara otomatis.

### Frekuensi
- **Daily** — berjalan setiap 24 jam
- **Weekly** — berjalan setiap 7 hari
- **Monthly** — berjalan sekali per bulan
- **Cron** — ekspresi cron kustom (misal \`*/5 * * * *\`)

### Target environment
Setiap jadwal bisa menargetkan: local, staging, atau production.

### Kontrol
- **Run Now** — jalankan segera
- **Pause/Resume** — toggle tanpa menghapus
- **Webhook** — dapatkan notifikasi saat gagal via Slack/Telegram

## Background scheduler
Server menjalankan goroutine background yang memeriksa jadwal yang jatuh tempo setiap 60 detik dan membuat run secara otomatis.`,
    },
  },
  {
    slug: "risk",
    category: "features",
    title: { en: "Risk & Recommendations", id: "Risiko & Rekomendasi" },
    content: {
      en: `# Risk & Recommendations

## Risk scoring
Every test and schedule gets a risk score (0-100%) based on:
- Recent failure frequency
- Environment criticality (production > staging > local)
- Schedule staleness (no run in 48h+ = higher risk)
- Last run status

## Recommendations
The system suggests actions:
- **Run now** — for stale or high-risk schedules
- **Investigate** — for tests failing frequently
- **Prioritize** — for moderate-risk items

## Suite selection
Choose how to select tests for a run:
- **All** — run everything
- **High risk** — only tests with risk score ≥ 40%
- **Flaky** — only tests that flip between pass/fail
- **Impacted** — only tests affected by recent code changes (git diff)`,
      id: `# Risiko & Rekomendasi

## Skor risiko
Setiap test dan jadwal mendapat skor risiko (0-100%) berdasarkan:
- Frekuensi kegagalan terbaru
- Kritikalitas environment (production > staging > local)
- Keusangan jadwal (tidak ada run 48 jam+ = risiko lebih tinggi)
- Status run terakhir

## Rekomendasi
Sistem menyarankan aksi:
- **Run now** — untuk jadwal yang usang atau berisiko tinggi
- **Investigate** — untuk test yang sering gagal
- **Prioritize** — untuk item berisiko sedang

## Seleksi suite
Pilih cara memilih test untuk sebuah run:
- **All** — jalankan semua
- **High risk** — hanya test dengan skor risiko ≥ 40%
- **Flaky** — hanya test yang berganti-ganti antara pass/fail
- **Impacted** — hanya test yang terpengaruh perubahan kode terbaru (git diff)`,
    },
  },
  {
    slug: "security",
    category: "ops",
    title: { en: "Security", id: "Keamanan" },
    content: {
      en: `# Security

## Authentication
- API endpoints are protected by API key (X-Api-Key header).
- Set \`API_KEY\` environment variable to enable auth.
- Empty API_KEY = no authentication (development only).

## File access
- Video files are served behind API key authentication.
- No public access to recordings without valid credentials.

## Input validation
- All URL parameters are validated (alphanumeric + dash, max 64 chars).
- Request bodies are limited to 1 MiB via MaxBytesReader.
- Error messages: internal errors return plain text messages which may include \`err.Error()\` in some handlers. Future releases will standardize on \`{"error": "..."}\` JSON responses.

## Network security
- Steel Browser should NOT be exposed to public internet.
- Use Docker internal networking for service-to-service communication.
- Add a reverse proxy (nginx/Caddy) with TLS for production.

## Best practices
- Always set API_KEY in production.
- Use strong, unique keys.
- Rotate keys periodically.
- Monitor the /api/v1/notifications endpoint for failure alerts.`,
      id: `# Keamanan

## Autentikasi
- Endpoint API dilindungi oleh API key (header X-Api-Key).
- Set variabel environment \`API_KEY\` untuk mengaktifkan auth.
- API_KEY kosong = tanpa autentikasi (hanya untuk development).

## Akses file
- File video disajikan di balik autentikasi API key.
- Tidak ada akses publik ke rekaman tanpa kredensial yang valid.

## Validasi input
- Semua parameter URL divalidasi (alfanumerik + dash, maks 64 karakter).
- Body request dibatasi 1 MiB via MaxBytesReader.
- Pesan error: error internal mengembalikan pesan teks biasa yang mungkin menyertakan \`err.Error()\` di beberapa handler. Rilis mendatang akan standarisasi ke respons JSON \`{"error": "..."}\`.

## Keamanan jaringan
- Steel Browser TIDAK boleh diekspos ke internet publik.
- Gunakan networking internal Docker untuk komunikasi antar service.
- Tambahkan reverse proxy (nginx/Caddy) dengan TLS untuk produksi.

## Best practices
- Selalu set API_KEY di produksi.
- Gunakan key yang kuat dan unik.
- Rotasi key secara berkala.
- Monitor endpoint /api/v1/notifications untuk alert kegagalan.`,
    },
  },
  {
    slug: "deployment",
    category: "ops",
    title: { en: "Deployment", id: "Deployment" },
    content: {
      en: `# Deployment

## Docker Compose (recommended)

\`\`\`bash
make up      # Start all 6 services
make down    # Stop
make logs    # Tail logs
make ps      # Show status
make rebuild # Rebuild and restart
\`\`\`

## Services
| Service | Port | Purpose |
|---------|------|---------|
| backend | 8080 | Go API server |
| frontend | 3001 | Next.js dashboard |
| postgres | 5432 | Database |
| redis | 6379 | Job queue |
| steel-browser | 3010 | Headless browser |
| langgraph-sidecar | 8000 | Multi-agent orchestrator |

## Health checks
All services have Docker health checks. Use \`make ps\` to verify all are healthy.

## Data persistence
- PostgreSQL data: \`pgdata\` Docker volume
- Videos/screenshots: \`app_data\` Docker volume

## Minimum requirements
- Docker 24+
- 4GB RAM (8GB recommended)
- 2 vCPU (4 recommended)`,
      id: `# Deployment

## Docker Compose (direkomendasikan)

\`\`\`bash
make up      # Mulai semua 6 service
make down    # Stop
make logs    # Lihat log
make ps      # Tampilkan status
make rebuild # Rebuild dan restart
\`\`\`

## Service
| Service | Port | Fungsi |
|---------|------|--------|
| backend | 8080 | Server API Go |
| frontend | 3001 | Dashboard Next.js |
| postgres | 5432 | Database |
| redis | 6379 | Job queue |
| steel-browser | 3010 | Browser headless |
| langgraph-sidecar | 8000 | Orchestrator multi-agent |

## Health check
Semua service memiliki Docker health check. Gunakan \`make ps\` untuk memverifikasi semua sehat.

## Persistensi data
- Data PostgreSQL: volume Docker \`pgdata\`
- Video/screenshot: volume Docker \`app_data\`

## Kebutuhan minimum
- Docker 24+
- 4GB RAM (8GB direkomendasikan)
- 2 vCPU (4 direkomendasikan)`,
    },
  },
  {
    slug: "faq",
    category: "ops",
    title: { en: "FAQ", id: "FAQ" },
    content: {
      en: `# FAQ

## Can I use this without an Anthropic API key?
The dashboard, monitoring, and review features work without it. But test generation and execution require the AI model.

## Does it support PHP/Laravel projects?
PHP framework detection and PHPUnit generation are not yet implemented. The agent generates Playwright browser tests and API-level test scripts, which work against any web-facing project regardless of backend language. Native PHPUnit support is tracked as a future enhancement.

## How much does it cost?
Self-hosted is free. You only pay for your Anthropic API usage.

## Can I run it on ARM64 (Apple Silicon)?
Yes. All Docker images support ARM64.

## What happens if video recording fails?
The test run continues normally. Video status shows "failed" but results are unaffected.

## How do I reset all data?
\`\`\`bash
make down
docker volume rm agent_test_pgdata agent_test_app_data
make up
\`\`\``,
      id: `# FAQ

## Bisa digunakan tanpa API key Anthropic?
Dashboard, monitoring, dan fitur review bisa berjalan tanpanya. Tapi pembuatan dan eksekusi test membutuhkan model AI.

## Apakah mendukung project PHP/Laravel?
Deteksi framework PHP dan pembuatan PHPUnit belum diimplementasikan. Agent menghasilkan test browser Playwright dan script test level API, yang dapat bekerja dengan project web apapun terlepas dari bahasa backend. Dukungan PHPUnit native direncanakan sebagai peningkatan di masa depan.

## Berapa biayanya?
Self-hosted gratis. Anda hanya membayar penggunaan API Anthropic Anda.

## Bisa dijalankan di ARM64 (Apple Silicon)?
Ya. Semua Docker image mendukung ARM64.

## Apa yang terjadi jika rekaman video gagal?
Test run tetap berjalan normal. Status video menampilkan "failed" tapi hasil tidak terpengaruh.

## Bagaimana cara reset semua data?
\`\`\`bash
make down
docker volume rm agent_test_pgdata agent_test_app_data
make up
\`\`\``,
    },
  },
];
