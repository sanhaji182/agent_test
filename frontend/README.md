# GoTest Agent — Frontend Dashboard

Next.js 16.2.12 application built with the App Router, TypeScript, Tailwind CSS 4, and `lucide-react` icons.

## Quick Start

```bash
npm install
npm run dev      # Dev server on :3000
npm run build    # Production build
npm run lint     # ESLint
npm test         # Vitest suite (Testing Library + jsdom)
```

The dashboard expects `NEXT_PUBLIC_API_URL` pointing to the Go backend (default: `http://localhost:8080`). This is baked at build time — set it before `npm run build` for production images.

## Project Structure

| Path | Purpose |
|---|---|
| `src/app/layout.tsx` | Root layout, global CSS, theme init, sidebar, header |
| `src/app/*/page.tsx` | 16 App Router pages (see route map below) |
| `src/lib/api.ts` | Centralized typed REST + SSE client; the single browser API boundary |
| `src/lib/docs.ts` | Bilingual EN/ID embedded product documentation |
| `src/lib/utils.ts` | `cn()` helper (`clsx` + `tailwind-merge`) |
| `src/components/ui/` | Shared primitives: Badge, Card, Chart, Section (skeleton/empty state) |
| `src/components/console/` | Run inspection: Inspector, Timeline, Tabs, ScreenshotStrip |

## Route Map

| Route | Page | Key behavior |
|---|---|---|
| `/` | Control room | Global SSE with 10s polling fallback; onboarding seed |
| `/create` | Guided run creation | 4-step wizard → POST run → redirect to `/runs/[id]` |
| `/projects` | Project → plan → approve workflow | Feature extraction, draft generation, case approval |
| `/tests` | Approved test library | Search, run, refine, proposal review |
| `/runs` | Run inventory | Filterable list with inspector drawer |
| `/runs/[id]` | Run audit console | Live SSE events, video, screenshots, files, timeline, analysis, report, rerun, compare |
| `/runs/[id]/compare` | Run comparison | Structured diff between two runs |
| `/suites` | Test lists and schedules | Create lists, add schedules (daily/weekly/monthly), run-now |
| `/monitoring` | List health dashboard | Aggregated pass/fail per list with trends |
| `/risk` | Risk and recommendations | AI-derived risk scores and actionable recommendations |
| `/releases` | Release list | Display-only; creation is API-driven |
| `/reviews` | Review queue | Approve/reject/request-changes on human reviews |
| `/alerts` | Notification history | Reads the in-memory notification store |
| `/exports` | JSON exports | Risk, run, and confidence report downloads |
| `/docs` | Bilingual documentation | Renders `src/lib/docs.ts` with EN/ID language toggle |
| `/settings` | Provider and execution config | Optimistic saves; provider connection testing |

## Development Notes

- **No authentication is turned on by default.** The backend `API_KEY` is empty, and the dashboard does not send it. If you enable API-key protection on the backend, you must also add authentication to the frontend.
- **SSE connections have different error handling.** The global control-room stream reconnects and has a polling fallback. The per-run stream does not — a transient disconnect on `/runs/[id]` leaves the console stale.
- **Test suite lives in `src/test/`** (Vitest + Testing Library, 20 tests covering the API client, utils, and UI primitives). Run with `npm test`. Page-level and E2E coverage are still open items — see `.ai/TESTING.md`.
- **This Next.js version has breaking changes** from the training data of most AI models. Follow the guidance in `frontend/AGENTS.md` and consult `node_modules/next/dist/docs/` for installed APIs and conventions.
- **`frontend/src/components/sidebar.tsx`** is the canonical navigation inventory — not every route is listed there (e.g., `/projects` is reached from `/tests`).

## Docker

Build: `docker build -f frontend/Dockerfile -t gotest-frontend .`  
The image runs as non-root (UID 1001) on port 3001. `NEXT_PUBLIC_API_URL` is baked at build time — set it via `--build-arg` or in the Dockerfile `ENV` before `npm run build`.
