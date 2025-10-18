# Migration from the .NET server

The Go server keeps the versioned resource and telemetry routes used by the
Angular application. `docs/legacy-api.json` records the route inventory, and a
contract test checks registration. The unused `latest` and `series` source
stubs were never exposed by the previous server.

## Database import

Back up the source database and encryption material first. Stop writes to the
old application during the import. Use a fresh target database; an initialized
or nonempty DataCat schema is rejected. The database account needs rights to
create the `datacat` schema.

```sh
export LEGACY_DATABASE_URL='postgres://user:password@host/old_database'
export DATABASE_URL='postgres://user:password@host/new_database'
export SECRET_KEY="$(openssl rand -base64 32)"
go run ./cmd/import-legacy
```

The importer reads the final legacy schema in one repeatable-read snapshot. It
preserves resource IDs, tags, timestamps, role IDs, manual role assignments,
external role mappings, notification settings, and alert durations. A .NET tick
is converted to 100 nanoseconds. Target resources are written in one transaction.
The importer prints counts without credential values.

Legacy data sources were global. Their settings are copied into each namespace,
with the original ID, then encrypted with that namespace's scope. The old
`public` tables remain unchanged. Plugin metadata is imported in a disabled
state because DLLs need an HTTP replacement.

Secrets stored outside PostgreSQL are not part of this import. Copy them to the
namespace-scoped database secret API or the configured Vault KV v2 mount before
enabling notifications. Retain the new encryption key.

## Intentional contract changes

- Password login accepts JSON `{email,password}`. Passwords are no longer query
  parameters. Browser sessions use HttpOnly cookies.
- Data sources and notification settings are namespace-scoped. Editing their
  credentials requires a namespace Admin or global administrator.
- HTTP plugins replace uploaded .NET assemblies. Register an endpoint, then
  enable the plugin. The server never loads arbitrary native binaries.
- The unused SignalR chat service was removed. Metric panels continue to use the
  metric query API.
- The debug rollback route rejects requests. Database schema changes are
  performed through the deployment process.
- SQL source text is not exposed over HTTP. Database metadata routes require a
  global administrator and describe the Go storage layout.
- Metadata-specific legacy permission assignments are not imported. Review
  administrative access explicitly through realm and namespace roles.

OIDC roles from the realm, signed namespace claims, and external mappings are
resolved on each request. This replaces the previous periodic role sync job.

Verify dashboards, panels, queries, and notifications against the imported target
before switching traffic. Keep the old database available for rollback until the
new deployment is accepted.
