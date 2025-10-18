# Deployment

Build both images from the repository root:

```sh
docker build -f .ops/docker/server.Dockerfile -t datacat-server:local .
docker build -f .ops/docker/ui.Dockerfile -t datacat-ui:local .
```

The API image runs as a non-root user. The UI image serves static files through
an unprivileged nginx process and proxies `/api` to `datacat-server:8080`.

## Kubernetes

The manifests in `.ops/k8s` are a base for an existing cluster. Configure the
registry images, public hosts, OIDC issuer, and upstream allowlist before use.
Create a `datacat` namespace, a `datacat` Secret containing `DATABASE_URL` and
`SECRET_KEY` (plus `OIDC_CLIENT_SECRET` when required), and a `datacat-tls` TLS
Secret. Use your normal secret manager; do not commit their values.

`scripts/deploy.sh` renders the manifests for review by default. Passing
`--apply` applies them to the current kubectl context and waits for rollouts.
It does not build or publish images.

The HTTP ingress terminates TLS and forwards to the UI. Expose gRPC separately
through a TLS ingress or gateway that supports HTTP/2 gRPC. Internal gRPC uses
port 4317. Do not expose the raw internal listener directly on a public network.

Startup migrations take a PostgreSQL advisory transaction lock. The `-migrate`
flag applies additive schema setup before serving requests. A separate
`-migrate-only` invocation can be used before rollout instead. The database
account must have the required DDL rights when running migrations.

## Operations

- `/health/live` reports process availability.
- `/health/ready` checks PostgreSQL with a bounded timeout.
- `/metrics` exposes request and application-error counters for internal scraping.
- Logs are JSON and contain request IDs, methods, and paths; request bodies and
  token values are not logged.

Back up PostgreSQL and the encryption key before migrations. The legacy importer
uses a separate `datacat` schema and leaves the old tables intact. Keep a tested
restore procedure rather than invoking schema rollback over HTTP.

Inspect notifications that exhausted their retries:

```sql
SELECT id, namespace, channel_id, attempts, created_at
FROM datacat.outbox
WHERE delivered_at IS NULL AND attempts >= 8;
```

After correcting the channel or upstream service, requeue selected IDs by setting
`attempts=0`, `available_at=now()`, and `last_error=NULL`. Notification delivery
is at least once; downstream consumers must tolerate duplicates.
