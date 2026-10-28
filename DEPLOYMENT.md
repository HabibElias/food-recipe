# Deploy the API on Render

Render supports building a Dockerfile from a Git repository. The included
`render.yaml` declares two separate Docker web services: Hasura and the Go API.
It uses your existing hosted PostgreSQL database and does not provision one.
Render does not run this repository's Compose stack as one web service. The
frontend belongs in the separate `food-recipe-app` Vercel project.

## Provide the hosted database URL

For deployment managed from the cloud task, add `VERCEL_TOKEN` and
`RENDER_API_KEY` in the task environment's secret settings, scoped to
`api.vercel.com` and `api.render.com` respectively. Add
`HASURA_GRAPHQL_DATABASE_URL` as a runtime environment variable so the deployment
process can forward the actual URL to Render. A proxy-backed HTTPS token
placeholder cannot be used as a PostgreSQL connection string. Never send these
values through chat. No Neon API key is needed when the database already exists.

Enter the PostgreSQL connection string directly in Render's secret environment
field `HASURA_GRAPHQL_DATABASE_URL` for **food-recipe-hasura**. Do not paste it
into chat, commit it, put it in a Docker build argument, or set it in Vercel's
public variables. Keep the provider's TLS parameters. Use a direct PostgreSQL
endpoint or a provider-supported session pooler rather than a transaction pooler
unless the provider specifically documents Hasura compatibility.

Only the database provider and hostname need to be shared for connectivity
planning; keep the username, password, and full URL private. The Go API uses
`HASURA_URL`, an HTTPS GraphQL endpoint, and does not need the database password.

Do not apply `init.sql` automatically to an existing application database. Check
its schema and available Hasura metadata first; the supplied SQL is incomplete
for this frontend.

## Create the API service

1. Push the deployment changes to `HabibElias/food-recipe` on `master`.
2. In Render, create a **Blueprint** from that repository and branch. Render reads
   `render.yaml` and asks for the Hasura and API settings below. Alternatively,
   create two Docker **Web Services** using `./deploy/hasura/Dockerfile` for
   Hasura and `./Dockerfile` for the API, with repository root as build context.
3. Supply these variables in Render's environment settings:

   | Variable | Value |
   | --- | --- |
   | `HASURA_URL` | The deployed Hasura HTTPS URL ending in `/v1/graphql` |
   | `HASURA_GRAPHQL_ADMIN_SECRET` | The same private admin secret configured on Hasura |
   | `HASURA_JWT_SECRET` | The same signing key configured in Hasura's JWT JSON; use a strong key of at least 32 bytes |
   | `HASURA_ACTION_SECRET` | A separate random webhook secret matching Hasura; required for action/event authentication |
   | `ALLOWED_ORIGINS` | The exact frontend origin, e.g. `https://your-project.vercel.app`; comma-separate additional trusted origins |
   | `CALLBACK_URL` | `https://<your-api>.onrender.com/v1/payment/verify` |

   Render supplies `PORT`. The API already listens on all interfaces and reads
   that variable. Its health-check path is `/v1/healthz`.
4. Add `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`,
   `CLOUDINARY_API_SECRET`, and `CHAPA_SECRET_KEY` when enabling real image
   uploads and payments. Store credentials only in provider environment settings.
5. Deploy and request `https://<your-api>.onrender.com/v1/healthz`; expect
   `{"status":"ok"}`. This checks the API process, so also test signup/login
   through Hasura before calling the application functional.

The Blueprint chooses free instances to avoid requesting paid resources. Free
services can sleep and interrupt subscriptions or delay callbacks. Select
appropriate always-on plans before relying on production subscriptions,
payments, or notification delivery.

## Hasura and the database

The selected deployment uses Hasura GraphQL Engine **v2.44.0** with hosted
PostgreSQL. Render Postgres or another compatible managed PostgreSQL provider
can supply the database URL.

For a PostgreSQL deployment:

1. Configure the hosted PostgreSQL URL on Hasura, then use the initializer below
   for the **new, empty** database. It creates the base schema, the additional
   frontend tables, and matching Hasura metadata through Hasura's admin API.
2. Configure the Blueprint's **food-recipe-hasura** service, which uses the
   existing image `hasura/graphql-engine:v2.44.0` through a minimal Dockerfile:

   | Variable | Value |
   | --- | --- |
   | `HASURA_GRAPHQL_DATABASE_URL` | Your PostgreSQL connection URL |
   | `HASURA_GRAPHQL_SERVER_PORT` | `8080`; set Render's `PORT` to `8080` too |
   | `HASURA_GRAPHQL_ADMIN_SECRET` | A private random admin secret matching the Go API |
   | `HASURA_GRAPHQL_JWT_SECRET` | JSON: `{"type":"HS256","key":"<the same signing key as the Go API>"}` |
   | `HASURA_ACTION_SECRET` | The same private webhook secret configured on the Go API |
   | `API_BASE_URL` | The public HTTPS origin of the Go API, without a trailing slash |
   | `HASURA_GRAPHQL_UNAUTHORIZED_ROLE` | `anon` |
   | `HASURA_GRAPHQL_CORS_DOMAIN` | Your exact Vercel frontend origin |
   | `HASURA_GRAPHQL_ENABLE_CONSOLE` | `false` for production |
   | `HASURA_GRAPHQL_DEV_MODE` | `false` |

   Use `/healthz` as Hasura's health check. Its public GraphQL endpoint is
   `https://<hasura-service>.onrender.com/v1/graphql`, and subscriptions use
   `wss://<hasura-service>.onrender.com/v1/graphql`. Vercel hosts the Nuxt
   frontend; Hasura hosts the long-lived GraphQL WebSocket connections.
3. Run the versioned initializer from this repository root, with `HASURA_URL` and
   `HASURA_GRAPHQL_ADMIN_SECRET` securely available in the process environment:

   ```sh
   python3 deploy/hasura/bootstrap.py --initialize-empty-database
   ```

   The endpoint must be HTTPS and end in `/v1/graphql`. The initializer applies
   `init.sql` and `deploy/database/schema.sql`, then installs metadata generated
   by `deploy/hasura/metadata.py`. It rejects an unexpected existing database or
   unrelated metadata instead of overwriting it, and recognizes its own completed
   initialization on repeat runs. Hasura resolves webhook URLs from `API_BASE_URL`
   and sends `X-Hasura-Action-Secret` from its `HASURA_ACTION_SECRET` variable.

   The schema retains numeric recipe prices; the frontend mutations use Hasura's
   `numeric` scalar. Metadata includes the recipe relationships, signup/login/Me
   actions, role permissions, and the recipe-like notification trigger. Account
   email/password columns are not available through public profile queries.
   Users can change only their own recipes and notifications; paid instructions
   require ownership or a successful purchase.

Image-upload and payment actions are currently **admin-only**. Supplying
Cloudinary or Chapa credentials alone does not grant public action access.
Complete their ownership/price validation and integration tests before enabling
user permissions; normal signup, recipe data, and notifications do not require
those external integrations.

For an existing local database created from the old `init.sql`, review and apply
the additive `deploy/database/schema.sql` migration with a trusted SQL client
before using the updated API, which reads the new profile fields. Do not point
the empty-database initializer at an unrelated existing production database.

## Docker validation

The isolated schema/auth/permission/subscription smoke suite creates and removes
its own local Docker resources. With Go 1.26 and Node 22+ available, run it from
this repository root:

```sh
python3 deploy/hasura/smoke_test.py
```

If `go` on your PATH is not the required toolchain, set `GO_BINARY` to the Go
executable path. This suite does not use your hosted database or cloud API tokens.

The image uses Go 1.26.10, checked-in `vendor/` dependencies, a static binary, and
a non-root runtime with CA certificates. To validate it locally:

```sh
docker build -t food-recipe-api .
docker run --rm -p 8000:8000 --env-file .env food-recipe-api
```

Use a separate terminal to request `/v1/healthz`. An API container needs a
reachable Hasura endpoint; `localhost` inside it refers to that container.

References: [Render Docker](https://render.com/docs/docker),
[Render Blueprint specification](https://render.com/docs/blueprint-spec),
[Hasura v2 documentation](https://hasura.io/docs/2.0/).
