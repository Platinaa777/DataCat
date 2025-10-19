# Configuration

The server reads environment variables and fails startup when required values are
missing. It never reads `.env` implicitly; development scripts load that file.

| Variable | Default or requirement |
| --- | --- |
| `DATABASE_URL` | Required PostgreSQL connection string |
| `SECRET_KEY` | Required base64-encoded 32-byte key |
| `HTTP_ADDR` | `:8080` |
| `PUBLIC_URL` | `http://localhost:4200`; browser return URL |
| `ALLOWED_ORIGIN` | `http://localhost:4200`; exact browser origin |
| `OIDC_ISSUER` | Required outside explicit development token mode |
| `OIDC_CLIENT_ID` | Required; also the expected access token audience |
| `OIDC_CLIENT_SECRET` | Optional for public clients |
| `OIDC_REDIRECT_URL` | Required OIDC callback URL |
| `OUTBOUND_ALLOWED_HOSTS` | Comma-separated exact authorities, including non-default ports |
| `REDIS_URL` | Optional Redis URL; otherwise use bounded memory caching |
| `VAULT_ADDR` | Optional Vault address; enables KV v2 secret storage |
| `VAULT_TOKEN` | Required with `VAULT_ADDR` |
| `VAULT_MOUNT` | `secret` |
| `APP_ENV` | Set `development` only for local development |
| `DEV_TOKEN` | Optional local API token, at least 32 characters; requires `APP_ENV=development` |

Generate a key with `openssl rand -base64 32`. Keep the key stable across restarts
and replicas. Replacing it without re-encrypting stored credentials makes them
unreadable. Keep a backup alongside the database backup, with separate access.

## Identity

Configure an OIDC client with Authorization Code flow, PKCE S256, and the exact
callback URL. Access tokens must include the DataCat client in their audience.
The local Keycloak realm includes an audience mapper.

The `datacat-admin` realm role grants global administration. Other users can
receive namespace roles through a signed `namespace_roles` claim:

```json
{"namespace_roles":{"00000000-0000-0000-0000-000000000000":["Viewer"]}}
```

Alternatively, administrators can create external role mappings through
`POST /api/v1/external-role`. Internal role IDs remain `1=Admin`, `2=Viewer`,
`3=Editor`. External mappings are resolved on each request. Imported manual role
assignments are preserved by identity subject. The server does not use a role
from one namespace to authorize another.

The browser uses a same-origin HttpOnly cookie. API clients
use `Authorization: Bearer <token>`.
Terminate TLS at the deployment ingress and expose only the intended service
ports. Password login is rate limited; Authorization Code login is the normal
browser flow.

## Data sources

Connection settings remain JSON strings to match the web client.

```json
{"serverUrl":"http://prometheus:9090","authType":"None"}
```

Prometheus and Jaeger support `None`, `Basic`, and `Bearer` authentication, using
`username`, `password`, or `authToken`. Elasticsearch uses `clusterUrl`,
`indexPattern`, `userName`, and `password`. Add each upstream authority to
`OUTBOUND_ALLOWED_HOSTS`. Do not add arbitrary user-controlled hosts.

## Notifications

Channel settings use the previous property names:

```json
{"Url":"https://hooks.example.com/alerts"}
```

```json
{"TelegramTokenPath":"telegram-token","ChatId":"123456"}
```

```json
{"DestinationEmail":"alerts@example.com","SmtpServer":"smtp.example.com","Port":465,"PasswordPath":"smtp-password"}
```

SMTP uses TLS on port 465 and requires STARTTLS on other ports. The destination
email is also the SMTP account and sender. Add `smtp.example.com:465` and
`api.telegram.org` to the outbound allowlist as needed. Secrets are scoped to the
channel namespace. Vault stores each value at `<mount>/data/<namespace>/<key>`
under the `value` property.
