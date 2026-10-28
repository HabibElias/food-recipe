# Food Recipe — Backend Service

Clean Architecture Go backend service for the Food Recipe platform, integrating **Fiber v2**, **Hasura GraphQL Engine**, **PostgreSQL**, **Cloudinary**, and **Chapa Payment Gateway**.

For Docker deployment on Render and production Hasura configuration, see
[DEPLOYMENT.md](DEPLOYMENT.md).

---

## Architecture Overview

This project follows the **Clean Architecture** pattern (Domain, Application, Infrastructure, Presentation):

```
food-recipe/
├── cmd/
│   └── app/
│       └── main.go              # Composition root & DI container
├── internal/
│   ├── domain/                  # Pure enterprise business rules (zero external deps)
│   │   ├── entities/            # User, Recipe, RecipeImage, Payment, Transaction, Notification
│   │   ├── errors/              # Domain error definitions
│   │   ├── ports/               # Outbound interfaces (PaymentGateway, FileStorage)
│   │   └── repositories/        # Repository interfaces
│   ├── application/             # Application business logic & orchestration
│   │   ├── dto/                 # Input/output data transfer objects
│   │   └── services/            # AuthService, PaymentService, MediaService, NotificationService
│   ├── infrastructure/          # Adapters to external services & databases
│   │   ├── config/              # Environment-based configuration
│   │   ├── payment/chapa/       # Chapa payment gateway adapter
│   │   ├── persistence/hasura/  # Hasura GraphQL persistence adapter & repositories
│   │   ├── security/            # JWT HS256 tokens & bcrypt password hashing
│   │   └── storage/cloudinary/  # Cloudinary image upload adapter
│   └── presentation/            # Inbound HTTP adapters
│       ├── http/
│       │   ├── handlers/        # Fiber HTTP route handlers
│       │   ├── middleware/      # Auth, CORS, Logger, Recover
│       │   └── router/          # Route registration
│       └── response/            # Standardized JSON envelopes & error mapping
├── Dockerfile                   # Multi-stage production build
├── docker-compose.yml           # Postgres + Hasura + App stack
├── init.sql                     # Postgres database schema & indexes
├── Makefile                     # Developer automation commands
└── .env.example                 # Sample environment variables
```

---

## Quick Start

### 1. Prerequisites
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed & running.
- [Go 1.24+](https://go.dev/) (only required if running the Go app locally outside Docker).
- [Make](https://www.gnu.org/software/make/) (available on Linux/macOS, and via `winget install ezwinports.make` on Windows).

### 2. Configure Environment Variables
Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```
Update your credentials (`CLOUDINARY_*`, `CHAPA_SECRET_KEY`) in `.env`.

---

## How to Run Hasura with Docker

You have two main development workflows:

### Workflow A: Run Everything in Docker (Recommended for quick testing)
Starts PostgreSQL 15, Hasura GraphQL Engine, and the Go App container together:

```bash
make docker-up
```
- **Hasura Console:** [http://localhost:8080/console](http://localhost:8080/console) (Admin Secret: `myadminsecretkey`)
- **Go API Gateway:** [http://localhost:8000](http://localhost:8000)
- **PostgreSQL:** `localhost:5432` (User: `postgres`, Password: `postgres`, DB: `food_recipe`)

To stop all containers:
```bash
make docker-down
```

---

### Workflow B: Run Hasura in Docker + App Locally (Recommended for Go development)
Runs PostgreSQL and Hasura in Docker, while running the Go application on your host machine for fast development, debugging, and live reload:

1. **Start Postgres and Hasura:**
   ```bash
   make docker-hasura
   ```

2. **Run the Go application locally:**
   ```bash
   make run
   ```
   The application connects to Hasura at `http://localhost:8080/v1/graphql`.

---

## Setting Up Hasura Console (Step-by-Step)

Once Hasura is running:

### Step 1: Open Hasura Console
1. Open your browser to [http://localhost:8080/console](http://localhost:8080/console).
2. Enter the admin secret: `myadminsecretkey` (or your custom `HASURA_GRAPHQL_ADMIN_SECRET`).

### Step 2: Track Database Tables
When Hasura starts up with Docker, `init.sql` automatically creates all required tables and indexes (`user`, `recipe`, `recipe_image`, `recipe_like`, `transaction`, `payment`, `notifications`).

1. In the top navigation bar, click on **Data**.
2. Click on **default** database -> **public** schema.
3. You will see **Untracked tables or views**. Click **Track All**.

### Step 3: Track Relationships
1. Below the tables list, under **Untracked foreign-key relationships**, click **Track All**.
2. Hasura will automatically establish:
   - `recipe_like.recipe` & `recipe_like.user`
   - `recipe.images` & `recipe.likes`
   - `user.recipes`

### Step 4: Configure Hasura Actions (Optional / Direct Integration)
Hasura Actions route GraphQL mutations to your Go backend endpoints.

> **Webhook URLs:**
> - If running Hasura in Docker and Go App on host machine: use `http://host.docker.internal:8000/v1/...`
> - If running Hasura and Go App inside Docker Compose: use `http://app:8000/v1/...`

| Action Name | GraphQL Type | Webhook URL | Description |
|---|---|---|---|
| `login` | Mutation | `/v1/login` | Authenticate user and return Hasura JWT |
| `signup` | Mutation | `/v1/signup` | Register user with hashed password |
| `createPayment` | Mutation | `/v1/payment` | Initialize Chapa checkout session |
| `uploadRecipeImages` | Mutation | `/v1/upload-recipe-images` | Upload recipe photos to Cloudinary |
| `uploadUserImage` | Mutation | `/v1/upload-user-image` | Upload user avatar to Cloudinary |

### Step 5: Configure Hasura Event Triggers
To send notifications when a recipe is liked:
1. In Hasura Console, go to **Events** -> **Create Trigger**.
2. **Trigger Name:** `recipe_like_notification`
3. **Table:** `recipe_like`
4. **Operations:** Check `Insert`.
5. **Webhook URL:** `http://host.docker.internal:8000/v1/like-notification` (or `http://app:8000/v1/like-notification`).
6. Click **Create Event Trigger**.

---

## Makefile Commands Reference

Run `make` or `make help` to see all available commands:

| Command | Description |
|---|---|
| `make run` / `make dev` | Run the Go application locally (`go run ./cmd/app`) |
| `make build` | Build compiled binary into `bin/food-recipe` |
| `make test` | Run all unit and integration tests |
| `make test-cover` | Run tests with coverage summary |
| `make fmt` | Format all Go source files (`gofmt -s -w .`) |
| `make tidy` | Run `go mod tidy` and update `vendor/` directory |
| `make clean` | Clean up build binaries and temporary files |
| `make docker-up` | Start full Docker stack (Postgres + Hasura + App) in background |
| `make docker-hasura` | Start only Postgres + Hasura (for local app development) |
| `make docker-down` | Stop and remove all containers |
| `make docker-build` | Rebuild the Go app Docker container image |
| `make docker-logs` | Stream logs across all containers |
| `make docker-logs-app` | Stream logs from Go app container |
| `make docker-logs-hasura` | Stream logs from Hasura container |
| `make docker-clean` | Stop containers and remove Postgres database volume (fresh start) |
| `make docker-restart` | Restart all running containers |

---

## HTTP Endpoints Reference

All application endpoints are mounted under `/v1`:

| Method | Endpoint | Auth | Purpose |
|---|---|---|---|
| `POST` | `/v1/login` | None | Verify password and issue Hasura HS256 JWT |
| `POST` | `/v1/signup` | None | Register a new user |
| `POST` | `/v1/me` | Bearer JWT | Fetch current user from JWT session claims |
| `POST` | `/v1/payment` | Bearer JWT / Hasura Action | Initialize Chapa checkout payment |
| `GET`  | `/v1/payment/verify` | None | Chapa payment verification callback |
| `POST` | `/v1/upload-recipe-images` | Bearer JWT / Hasura Action | Upload multipart recipe photos to Cloudinary |
| `POST` | `/v1/upload-user-image` | Bearer JWT / Hasura Action | Upload user avatar and update profile |
| `POST` | `/v1/like-notification` | None (Hasura Webhook) | Process `recipe_like` insert event and create notification |
