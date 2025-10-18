# DataCat

DataCat is an observability workspace for metrics, logs, traces, dashboards, and alerts.
The server is written in Go 1.21. The web application uses Angular 18 and PrimeNG.

## Features

- Namespace access with Admin, Editor, and Viewer roles.
- Keycloak-compatible OIDC login with PKCE.
- Dashboards, panels, query variables, and tags.
- Prometheus queries, Elasticsearch logs, and Jaeger traces.
- Scheduled alerts with pending, firing, muted, and error states.
- Telegram, SMTP, webhook, and HTTP plugin notifications.
- PostgreSQL persistence, encrypted credentials, and optional Vault KV v2.
- Local or Redis query caching.
- gRPC metric ingestion and a Go OpenTelemetry exporter.

## Local development

Install Go 1.21.13, Node.js 20.19, Docker, and Docker Compose.

```sh
make dev
```

Open http://localhost:4200 and sign in as `developer` with password
`local-development-only`. This account belongs to the local Keycloak realm only.
The script creates a private `.env`, starts PostgreSQL and Keycloak, installs the
web dependencies, and runs both applications. Ports 4200, 8080, 8081, 4317, and
54329 must be available. Use `docker compose down` to stop the dependencies.

For separate terminals:

```sh
./scripts/init-env.sh
set -a; source .env; set +a

docker compose up -d --wait postgres
docker compose up -d keycloak
# Wait for Keycloak to finish importing the local realm.
make run
# In another terminal:
make ui
```

## Checks

```sh
make vet
make test-race
make ui-check
TEST_DATABASE_URL='postgres://postgres:password@localhost:5432/datacat_test?sslmode=disable' make test-integration
```

Integration tests require a dedicated PostgreSQL database ending in `_test` and
a database user with permission to create temporary test databases.

## Repository

| Path | Purpose |
| --- | --- |
| `cmd/datacat` | Server entry point |
| `cmd/import-legacy` | Import the previous PostgreSQL schema |
| `internal/api` | HTTP, gRPC, and authentication |
| `internal/usecases` | Resource operations and alert evaluation |
| `internal/repository` | PostgreSQL access and migrations |
| `internal/extapi` | Telemetry, notifications, and Vault clients |
| `pkg/exporter` | OpenTelemetry metric exporter |
| `api` | HTTP and protobuf contracts |
| `frontend/datacat-ui` | Angular application |
| `.ops` | Images, local identity provider, and Kubernetes manifests |

See [development](docs/development.md), [configuration](docs/configuration.md),
[architecture](docs/structure.md), [deployment](docs/deployment.md),
[migration](docs/migration.md), and [plugin protocol](docs/plugins.md), and [screenshots](docs/gallery.md). Verification results are recorded in
[validation](docs/validation.md).

## License

[MIT](LICENSE).
