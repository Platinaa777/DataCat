# Architecture

The server has one executable and one PostgreSQL database. HTTP uses OIDC authentication. A background worker evaluates alerts and delivers
queued notifications.

## Boundaries

`cmd` parses process options and handles signals. `bootstrap` wires dependencies
with ordinary constructors:

- `dependencies.go` opens PostgreSQL, cache, and external clients.
- `router.go` constructs feature handlers and registers them with the router.
- `http.go` configures OIDC and HTTP server limits.
- `app.go` starts the server and worker and coordinates shutdown.
- `run.go` loads configuration and selects migration or serving mode.

Each package under `api/http` represents a feature: dashboards, panels, alerts,
variables, namespaces, data sources, notifications, plugins, secrets, or telemetry
queries. Its `routes.go` declares routes, `handler.go` contains named handlers,
and `models.go` defines request payloads. The shared `transport` package handles
JSON decoding, responses, errors, namespace checks, and pagination. The root
router adds authentication, middleware, health checks, and scrape counters.

`api` owns request decoding, access checks, response mapping, and transport errors.
A request carries one namespace in `X-Namespace`; a body or query parameter cannot
override it. Global integration management requires the `datacat-admin` realm
role. Namespace roles grant access only to their matching namespace.

`usecases` owns validation, resource relationships, transactions, template
rendering, and query variable substitution. Repository ports live in
`internal/usecases/contract`.

`repository/postgres` stores typed resource payloads in JSONB envelopes identified
by namespace, kind, and ID. Named resources have a unique name within their kind
and namespace. Searches use bounded SQL pagination, allowed fields, and bound
parameters. Version checks reject stale updates. Multi-resource writes use
serializable transactions.

`extapi` contains explicit adapters for Prometheus, Elasticsearch, Jaeger, SMTP,
Telegram, webhooks, Vault, and HTTP plugins. Outbound HTTP calls have deadlines,
response size limits, an exact host allowlist, and no automatic redirects.

## Alerts and delivery

The worker locks due alerts with `FOR UPDATE SKIP LOCKED`, allowing multiple
replicas to share work. It stores alert transitions and notification outbox rows
in the same transaction. A failed query enters Error rather than reporting a
successful recovery.

Pending alerts wait for the configured interval before firing. Repeat intervals
limit notifications while a condition remains true. Muting postpones evaluation.
A successful recovery resets the pending and repeat timers.

Outbox delivery is at least once. A process can stop after a remote service
accepts a message and before the delivery transaction commits. Consumers should
tolerate duplicates. Failed messages retry with exponential delays; after eight
attempts they remain in the outbox for operator review.

## Credentials and caching

Database secrets, data source credentials, and notification settings use AES-GCM.
Associated data binds ciphertext to the namespace, resource type, and ID. Vault
KV v2 can replace database secret storage. Notification credentials are referenced
by secret key rather than embedded in the channel configuration.

Query caching has a five-second TTL. Keys include the namespace and query
parameters. The default cache has a bounded capacity; Redis is optional and fails
open on cache outages. Access is checked before cached values are returned.

## Telemetry queries

Metrics, logs, and traces stay in Prometheus, Elasticsearch, Jaeger, or a provider
behind an HTTP plugin. DataCat queries those providers and stores only application
configuration, alert state, and the notification outbox in PostgreSQL.

Delivered outbox entries are retained for 30 days. Cleanup runs hourly while the
worker is running.
