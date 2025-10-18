# HTTP plugins

Plugins run as separate services. An administrator registers a name, version,
author, endpoint, and optional settings through `POST /api/v1/plugin/add`.
The endpoint authority must be allowed by `OUTBOUND_ALLOWED_HOSTS`. Plugins are
registered disabled and enabled with `POST /api/v1/plugin/{id}?isActive=true`.

To use a plugin as a data source, register a data source type with the same name,
then create a data source with that type. Notification destinations follow the
same naming convention.

The server sends a POST request to the configured endpoint:

```json
{
  "version": 1,
  "operation": "metrics.query",
  "namespaceId": "00000000-0000-0000-0000-000000000000",
  "connectionSettings": {"serverUrl":"https://upstream.example.com"},
  "settings": "plugin configuration",
  "input": {"query":"up"}
}
```

Supported operations:

| Operation | Input | Response |
| --- | --- | --- |
| `metrics.query` | query | MetricPoint array |
| `metrics.queryRange` | query, from, to, step | TimeSeries array |
| `logs.search` | LogSearch request | Page of LogEntry |
| `traces.names` | service; empty for services | string array |
| `traces.get` | traceId | TraceEntry |
| `traces.search` | TraceSearch request | TraceEntry array |
| `notifications.send` | message | any successful response |

Return JSON matching the corresponding HTTP schema. Non-2xx responses fail the
operation. Requests have a 15-second HTTP timeout and a 16 MiB response limit.
Redirects are rejected. Use TLS and your deployment's service authentication when
crossing a trust boundary. A registered plugin receives the configured upstream
credentials, so registration is an administrative operation.

The previous DLL interface cannot be loaded into the Go process. Move its query
or notification implementation behind this protocol and update the registration.
For imported plugin metadata, remove the disabled registration and register its
HTTP replacement before enabling dependent resources.
