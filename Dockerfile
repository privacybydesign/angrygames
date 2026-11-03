FROM golang:1.27.2 AS build
WORKDIR /app
COPY . .

# Static binary: the distroless static image has no libc.
RUN CGO_ENABLED=0 go build -o ./server

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /app/server /app/server

# Declared explicitly to satisfy static scanners; the :nonroot tag already
# defaults to this same user (UID 65532).
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT [ "/app/server" ]
