# Development

Use Go 1.21.13, Node.js 20.19, PostgreSQL 16, and Docker Compose. Run `make dev`
for the local stack. The imported Keycloak account is intended only for this
stack. Existing `.env` files are left unchanged by `scripts/init-env.sh`.

## Checks

`make test` runs unit and contract tests. `make test-race` enables the race
detector. `make test-integration` adds HTTP and worker
scenarios. The integration database must end in `_test`; the tests replace its
DataCat resources. Do not point this variable at application data.

On newer macOS releases, Go 1.21's internal linker can produce binaries rejected
by dyld. With the Xcode command line tools installed, use:

```sh
CGO_ENABLED=1 GOTOOLCHAIN=go1.21.13 make GO_TEST_FLAGS='-ldflags=-linkmode=external' test-race
```

The Angular checks are `npm run lint`, `npm run typecheck`, and `npm run build`
in `frontend/datacat-ui`. Generated output, dependencies, IDE folders, and local
configuration are ignored.

## Browser tests

`npm run e2e` exercises a real OIDC login, HttpOnly cookies, dashboard persistence,
and navigation. It expects the UI on port 14200, the API on 18080, and a local
Keycloak client allowing `http://localhost:14200/api/v1/user/callback`.
Run the API against a separate database from the Go integration tests.

```sh
# API environment overrides:
HTTP_ADDR=:18080
PUBLIC_URL=http://localhost:14200
ALLOWED_ORIGIN=http://localhost:14200
OIDC_REDIRECT_URL=http://localhost:14200/api/v1/user/callback

# UI:
npm start -- --port 14200 --proxy-config proxy.e2e.json
npx playwright install chromium
npm run e2e
```

Override `E2E_BASE_URL`, `E2E_USERNAME`, and `E2E_PASSWORD` if needed. An installed
Chromium executable can be selected with `PLAYWRIGHT_CHROMIUM_EXECUTABLE`.

## Contracts and generated files

The Angular client preserves the versioned HTTP contract. `api/openapi.json`
describes its requests and responses. Regenerate it with
`python3 scripts/generate-openapi.py` after changing the client contract.

`make generate` updates the OpenAPI document. Contract tests compare every
versioned route with the document in both directions.

Keep comments focused on constraints, protocol details, or non-obvious choices.
Commit messages should describe the resulting change, for example `Fix alert
repeat scheduling` or `Redesign deploy scripts for K8s`.
