# Changelog

## Maintenance

- Remove custom metric ingestion, its exporter, and the old database importer.
- Split HTTP routes and handlers by feature and wire dependencies in bootstrap.
- Remove unused database management endpoints and update the HTTP contract.

## Go server

- Replace the .NET server with a Go implementation.
- Preserve versioned resource and telemetry APIs for the Angular application.
- Add namespace authorization, encrypted configuration, and OIDC code flow.
- Add transactional alert scheduling and a notification outbox.
- Add PostgreSQL import tooling and HTTP plugin integration.
- Add container images, local development scripts, and Kubernetes manifests.
- Translate remaining interface text and remove unused source and IDE files.
- Add unit, database integration, contract and browser checks.
