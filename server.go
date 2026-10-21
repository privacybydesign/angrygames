package main

import (
	"embed"
	"log"
	"mime"
	"net/http"
)

// Only the public files are embedded. Anything not listed here (config,
// Dockerfile, tests, .git) cannot be served.
//
//go:embed index.html cart.html css fonts img js
var staticFiles embed.FS

const (
	invalidRequestMessage = "Invalid request"
	serverErrorMessage    = "Internal server error"
)

func init() {
	// The distroless runtime has no /etc/mime.types and Go's built-in table
	// has no font types, so without these the fonts are served sniffed.
	for ext, mimeType := range map[string]string{
		".eot":   "application/vnd.ms-fontobject",
		".otf":   "font/otf",
		".ttf":   "font/ttf",
		".woff":  "font/woff",
		".woff2": "font/woff2",
	} {
		if err := mime.AddExtensionType(ext, mimeType); err != nil {
			log.Fatal(err)
		}
	}
}

func newHandler(irma *irmaClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /start-session", startSessionHandler(irma))
	mux.Handle("GET /", http.FileServerFS(staticFiles))
	return mux
}

func startSessionHandler(irma *irmaClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, ok := sessionRequests[r.URL.Query().Get("type")]
		if !ok {
			http.Error(w, invalidRequestMessage, http.StatusBadRequest)
			return
		}

		response, err := irma.startSession(r.Context(), request)
		if err != nil {
			log.Printf("starting IRMA session: %v", err)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("writing session response: %v", err)
		}
	}
}
