# Auth Service (for TODO frontend)

This small Go service acts as an authentication gateway for a frontend TODO application. It talks to a Hasura GraphQL backend to look up users and returns JWTs that include Hasura claims so the frontend can call Hasura directly.

Core responsibilities
- Accept login requests (email + password).
- Verify credentials against users stored in Hasura.
- Issue HS256 JWTs containing Hasura claims (roles and x-hasura-user-id).
- Provide a protected `/v1/me` endpoint that validates the Bearer token and returns the authenticated user's info.

Why this exists
- Keeps the secret verification and password checks on a small backend service while letting the frontend make authorized GraphQL requests to Hasura using the returned JWT.

Quick architecture
- Frontend (TODO app) -> POST /v1/users (email/password) -> this service -> Hasura (verifies user) -> service returns JWT
- Frontend stores JWT and calls Hasura directly (with authorization header) or calls protected endpoints (e.g. `/v1/me`) on this service.

Setup

1. Prereqs

- Go 1.18+ (or your project's configured Go version).
- Access to a Hasura GraphQL instance with a `users` table containing at least `id`, `email`, `name`, and `password` (bcrypt hashed).

2. Environment variables

Create a `.env` or export these variables in your environment:

- PORT – port the service listens on (e.g. 8000)
- HASURA_URL – full URL of your Hasura GraphQL endpoint
- HASURA_GRAPHQL_ADMIN_SECRET – Hasura admin secret (if your Hasura is protected)
- HASURA_JWT_SECRET – secret used to sign the HS256 tokens served to the frontend

API

- POST /v1/users
	- Body: { "email": "user@example.com", "password": "plaintext" }
	- Success: 200
		{
			"user": { "id": 123, "name": "Alice", "email": "alice@example.com" },
			"token": "<JWT>"
		}
	- Errors: 400 (bad request), 401 (invalid credentials), 500 (server error)

- GET /v1/me
	- Header: Authorization: Bearer <JWT>
	- Success: 200 { "id": 123, "name": "Alice", "email": "alice@example.com" }
	- Errors: 401 (missing/invalid token), 500 (server error)

Notes & gotchas

- Passwords in Hasura must be bcrypt-hashed. The service uses `golang.org/x/crypto/bcrypt` to compare hashes.
- JWT lifetime in code may be short by default (check `helpers.GenerateHasuraJWT`); adjust `exp` as needed.
- The JWT must be HS256 and the frontend must attach it as `Authorization: Bearer <token>` to any requests.
- The `sub` claim is used as the user id in the token. If your Hasura `id` type is integer, the token `sub` is still treated as a string — the service converts it appropriately when querying Hasura by primary key.

Further improvements (ideas)

- Add refresh tokens for longer sessions.
- Rate-limit login endpoint and add brute-force protections.
- Add unit tests around middleware and token parsing.

Contact

If you run into build or runtime errors, run `go build` locally and paste the error output; I can help fix any type or import errors.
