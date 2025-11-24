# Food Recipe — Backend

This repository contains the backend services for the Food Recipe application. It's primarily a small Go-based gateway that:

- Integrates with Hasura GraphQL for data storage and permissions.
- Provides authentication (login/signup) and issues Hasura-compatible JWTs.
- Handles image uploads (Cloudinary) for recipes and user avatars.
- Integrates with Chapa for payment initialization and verification.
- Processes Hasura event webhooks (notifications on likes, etc.) and writes notification rows to the database.

Project responsibilities
- Authenticate users, compare bcrypt hashed passwords, and produce HS256 JWTs containing Hasura claims.
- Expose simple REST endpoints used by the frontend and Hasura Actions (image upload, payment init).
- Provide webhook endpoints for payment verification and Hasura event triggers (likes -> notifications).
- Upload images to Cloudinary and persist metadata/URLs in Hasura.

Architecture overview

Frontend (Nuxt) ↔ this Go gateway ↔ Hasura GraphQL ↔ Postgres

- The gateway signs JWTs that include Hasura claims so the frontend can call Hasura directly when appropriate.
- Some operations are implemented as Hasura Actions that call into this service (image upload, user image upload, payment initialize).
- Hasura event triggers (e.g. insert on `recipe_like`) call webhook endpoints in this service to create notifications.

Quick start

1. Prerequisites

- Go 1.26
- A Hasura instance (connected to Postgres) with the required tables tracked
- Cloudinary account (for image uploads)
- Chapa account (for payments)

2. Environment variables

Set these in your environment or in a `.env` file:

- `PORT` — port to run the service (e.g. `8000`)
- `HASURA_URL` — Hasura GraphQL endpoint
- `HASURA_GRAPHQL_ADMIN_SECRET` — Hasura admin secret (if applicable)
- `HASURA_JWT_SECRET` — secret to sign HS256 tokens for Hasura
- `CLOUDINARY_URL` / Cloudinary-specific envs as used in `clients/cloudinary` (cloud name, api key/secret)
- `CHAPA_SECRET_KEY` — Chapa secret for payment API

3. Build & run

```bash
go build && ./food-recipe
```

HTTP Endpoints (mounted under `/v1`)

- `POST /v1/login` — login (email + password) — returns JWT in body and Authorization header
- `POST /v1/signup` — create a new user (Hasura insert via the gateway)
- `POST /v1/me` — return current user (expects Hasura-compatible JWT in Authorization header)
- `POST /v1/payment` — initialize Chapa payment (Hasura Action)
- `GET  /v1/payment/verify` — verify Chapa payment (webhook/callback)
- `POST /v1/upload-recipe-images` — Hasura Action to upload recipe images to Cloudinary
- `POST /v1/upload-user-image` — Hasura Action to upload a user's avatar and update their `avatar_url`
- `POST /v1/like-notification` — Hasura event webhook for `recipe_like` inserts (creates a notification record)

Hasura integration points

- The repo includes `clients/hasura` with GraphQL query and mutation structs used by the go-graphql-client.
- Add Hasura Actions pointing to the relevant endpoints (image upload, user image upload, payment init).
- Configure Hasura Event Triggers (e.g. `recipe_like` insert) to POST to `/v1/like-notification`.

Frontend (Nuxt) integration suggestions

- Use the JWT returned by `POST /v1/login` for subsequent Hasura requests (set `Authorization: Bearer <token>`).
- For image uploads: call the Hasura Action that returns pre-signed or handled upload; the action returns the uploaded URL(s).
- For real-time notifications: use Hasura Subscriptions on the `notifications` table filtered by `user_id`.

Contact

If you want me to update this README with more project-specific details (exact Hasura migrations, sample GraphQL queries, or a step-by-step deployment guide), tell me which part to expand.
