# Phase 8: production deployment

Whisper's target layout is a Next.js frontend on Vercel, the Go API on Render, PostgreSQL on Neon, and source code in a private GitHub repository.

## Configuration map

| Service | Setting | Value |
| --- | --- | --- |
| Vercel | Root directory | `frontend` |
| Vercel | `NEXT_PUBLIC_API_URL` | `https://<render-service>.onrender.com/api/v1` |
| Render | Blueprint | Root `render.yaml` |
| Render | `DATABASE_URL` | Neon pooled connection string with `sslmode=require` |
| Render | `FRONTEND_ORIGIN` | Exact Vercel production origin, for example `https://<project>.vercel.app` |
| Render | `SESSION_COOKIE_SECURE` | `true` |
| Render | Health check | `/api/v1/health` |

The Go server uses Render's `PORT` value when present and falls back to `BACKEND_PORT` for local development. Secure production cookies use `SameSite=None` so credentialed API requests can cross from Vercel to Render; they remain `HttpOnly` and `Secure`. Use only exact origins, never `*`, with credentialed CORS.

In Neon, copy the pooled connection string from the connection details panel. Keep it in Render's secret environment variable field. Do not add it to GitHub, the frontend, or a committed `.env` file. Render's web service automatically supplies `PORT`.

## Deployment order

1. Create a Neon project and database. Keep a local copy of its pooled connection string in a password manager.
2. Push this repository to a private GitHub repository. Do not commit `.env` or production credentials.
3. Create the Render Web Service from the repository using the root `render.yaml` Blueprint. Set `DATABASE_URL` and the exact production `FRONTEND_ORIGIN` in Render's environment settings. Wait for `/api/v1/health` to report a connected database.
4. Run `go run ./cmd/create-admin` locally with the production `DATABASE_URL` loaded into that terminal to create the first production admin. The command prompts for the password without echoing it. Do not add a public registration route.
5. Import the same GitHub repository into Vercel and set its root directory to `frontend`. Add `NEXT_PUBLIC_API_URL` using the Render service URL, then deploy.
6. Update Render's `FRONTEND_ORIGIN` with the exact Vercel production URL and redeploy the API. Update `NEXT_PUBLIC_API_URL` and redeploy Vercel if the Render URL changed.
7. Verify health, admin login/logout, public conversation creation, message send/reply, WebSocket reconnect, and HTTPS cookie behavior on the deployed URLs.

The frontend and API use different provider domains (`vercel.app` and `onrender.com`). The production session cookie is therefore cross-site. Some browsers or privacy settings can block third-party cookies even with `SameSite=None; Secure`. If admin sessions fail in a browser, use custom subdomains on the same domain (for example `app.example.com` and `api.example.com`) or put the API behind a same-site reverse proxy.

## Free-tier limits

Free plans are appropriate for learning and previews, not a 24/7 service guarantee. Render free web services spin down after 15 minutes without inbound traffic and may take about a minute to wake; free web services also have monthly instance-hour and bandwidth/build limits. Neon Free currently includes 100 CU-hours and 1 GB database storage per project; compute scales to zero when idle, so the first connection can be slower. Vercel Hobby has monthly usage/build limits and is intended for personal projects. Check each provider's current limits before deployment because plans change.

References: [Render free services](https://render.com/docs/free), [Render native Go runtime](https://render.com/docs/language-support), [Vercel monorepos](https://vercel.com/docs/monorepos), [Vercel Hobby limits](https://vercel.com/docs/plans/hobby), and [Neon Free plan](https://neon.com/blog/neon-free-plan-1-gb-per-project).
