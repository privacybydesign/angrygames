# angrygames

Fake game shop that demonstrates age verification with [Yivi](https://yivi.app).
Adding a game to the cart starts a Yivi disclosure session that asks for proof
of the minimum age of the game (12+, 16+ or 18+). The browser checks the
disclosed attribute (see `js/main.js`); this is by design, the shop is a demo.

## Run

```
go run .
```

The server listens on port 8080 and needs an IRMA/Yivi server to start sessions.

| Variable          | Default                 | Meaning                                                 |
|-------------------|-------------------------|---------------------------------------------------------|
| `IRMA_SERVER_URL` | `http://localhost:8088` | Base URL of the IRMA server                             |
| `API_TOKEN`       |                         | Sent as the `Authorization` header to the IRMA server   |

## Endpoints

- `GET /` and the other static files (`index.html`, `cart.html`, `css/`,
  `fonts/`, `img/`, `js/`). They are embedded in the binary.
- `GET /start-session?type=12plus|16plus|18plus` starts a disclosure session at
  the IRMA server and returns its response unchanged. Responds 400 for an
  unknown type and 500 when the IRMA server call fails.

## Test

```
go vet ./... && go test ./...
```

## Docker

```
docker build -t ghcr.io/privacybydesign/angrygames .
docker run -p 8080:8080 -e IRMA_SERVER_URL=http://host.docker.internal:8088 ghcr.io/privacybydesign/angrygames
```
