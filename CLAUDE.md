# angrygames — repo notes

Go server (standard library only) that serves a fake game shop and starts Yivi
age-verification sessions.

## Build & test

- `go vet ./... && go test ./...`. The Go version is set in `go.mod`.
- Tests use a fake IRMA server (`httptest`); no IRMA server is needed.

## Gotchas

- Static files are embedded with `//go:embed` in `server.go`. A new public
  file or directory must be added to that list, and to the allowlist in
  `TestEmbeddedFilesPublicOnly`. Nothing else in the repo may be served.
- The runtime image is distroless and has no `/etc/mime.types`, so font MIME
  types are registered in `server.go`. Add any new asset extension that Go's
  built-in table lacks.
- `js/main.js` calls `start-session?type=<age>plus`. Keep the route in
  `server.go` and the URL in `js/main.js` in step.
- The age check happens in the browser (`js/main.js`), on purpose.
