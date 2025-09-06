# DataCat UI

Angular 18 and PrimeNG workspace for the Go server.

```sh
npm ci
npm start
```

The development server runs on port 4200 and proxies `/api` to port 8080.
Authentication uses a same-origin HttpOnly session cookie. Set up the local
identity provider using the [repository instructions](../../README.md).

Run `npm run lint`, `npm run typecheck`, and `npm run build` before committing.
Browser checks are described in [development](../../docs/development.md).
